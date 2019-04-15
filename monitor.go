package harvester

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"sync"
)

// GetFunc function definition for getting a value for a key.
type GetFunc func(string) (string, error)

type field struct {
	Name      string
	Kind      reflect.Kind
	Version   uint64
	SeedValue string
	EnvVarKey string
	ConsulKey string
}

type tag struct {
	Src Source
	Key string
}

// Monitor defines a monitoring interface.
type Monitor interface {
	Monitor()
}

// TypeMonitor definition.
type TypeMonitor struct {
	ch         <-chan *Change
	monitorMap map[Source]map[string]*field
	consulGet  GetFunc
	sync.Mutex
	cfg reflect.Value
}

// NewMonitor creates a new monitor.
func NewMonitor(cfg interface{}, ch <-chan *Change, consulGet GetFunc) (*TypeMonitor, error) {
	if cfg == nil {
		return nil, errors.New("configuration is nil")
	}
	if ch == nil {
		return nil, errors.New("change channel is nil")
	}
	if consulGet == nil {
		return nil, errors.New("consul get is nil")
	}
	tp := reflect.TypeOf(cfg)
	if tp.Kind() != reflect.Ptr {
		return nil, errors.New("configuration should be a pointer type")
	}
	m := &TypeMonitor{
		ch:         ch,
		cfg:        reflect.ValueOf(cfg).Elem(),
		monitorMap: make(map[Source]map[string]*field),
		consulGet:  consulGet,
	}
	if err := m.setup(tp); err != nil {
		return nil, err
	}
	return m, nil
}

// Monitor changes and apply them.
func (tm *TypeMonitor) Monitor() {
	for c := range tm.ch {
		tm.applyChange(c)
	}
}

func (tm *TypeMonitor) applyChange(c *Change) {
	mp, ok := tm.monitorMap[c.Src]
	if !ok {
		logWarnf("source %s not found", c.Src)
		return
	}
	fld, ok := mp[c.Key]
	if !ok {
		logWarnf("key %s not found", c.Key)
		return
	}
	if fld.Version > c.Version {
		logWarnf("version %d is older than %d", c.Version, fld.Version)
		return
	}

	err := tm.setValue(fld.Name, c.Value, fld.Kind)
	if err != nil {
		logErrorf("failed to set value %s of kind %d on field %s from source %s", c.Value, fld.Kind, fld.Name, c.Src)
		return
	}
	fld.Version = c.Version
}

func (tm *TypeMonitor) setValue(name, value string, kind reflect.Kind) error {
	tm.Lock()
	defer tm.Unlock()
	f := tm.cfg.FieldByName(name)
	switch kind {
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		f.SetBool(b)
	case reflect.String:
		f.SetString(value)
	case reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		f.SetInt(v)
	case reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		f.SetFloat(v)
	default:
		return fmt.Errorf("unsupported kind: %v", kind)
	}
	return nil
}

func (tm *TypeMonitor) setup(tp reflect.Type) error {
	ff, err := getFields(tp.Elem())
	if err != nil {
		return err
	}
	err = tm.applyInitialValues(ff)
	if err != nil {
		return err
	}
	err = tm.createMonitorMap(ff)
	if err != nil {
		return err
	}
	return nil
}

func getFields(tp reflect.Type) ([]*field, error) {
	var ff []*field
	for i := 0; i < tp.NumField(); i++ {
		fld := tp.Field(i)
		kind := fld.Type.Kind()
		if !isKindSupported(kind) {
			return nil, fmt.Errorf("field %s is not supported(only bool, int64, float64 and string)", fld.Name)
		}
		f := &field{
			Name:    fld.Name,
			Kind:    kind,
			Version: 0,
		}
		value, ok := fld.Tag.Lookup(string(SourceSeed))
		if ok {
			f.SeedValue = value
		}
		value, ok = fld.Tag.Lookup(string(SourceEnv))
		if ok {
			f.EnvVarKey = value
		}
		value, ok = fld.Tag.Lookup(string(SourceConsul))
		if ok {
			f.ConsulKey = value
		}
		ff = append(ff, f)
	}
	return ff, nil
}

func (tm *TypeMonitor) applyInitialValues(ff []*field) error {
	for _, f := range ff {
		if f.SeedValue != "" {
			err := tm.setValue(f.Name, f.SeedValue, f.Kind)
			if err != nil {
				return err
			}
		}
		if f.EnvVarKey != "" {
			value, ok := os.LookupEnv(f.EnvVarKey)
			if !ok {
				continue
			}
			err := tm.setValue(f.Name, value, f.Kind)
			if err != nil {
				return err
			}
		}
		if f.ConsulKey != "" {
			value, err := tm.consulGet(f.ConsulKey)
			if err != nil {
				return err
			}
			err = tm.setValue(f.Name, value, f.Kind)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (tm *TypeMonitor) createMonitorMap(ff []*field) error {
	for _, f := range ff {
		if f.ConsulKey == "" {
			continue
		}
		_, ok := tm.monitorMap[SourceConsul]
		if !ok {
			tm.monitorMap[SourceConsul] = map[string]*field{f.ConsulKey: f}
		} else {
			_, ok := tm.monitorMap[SourceConsul][f.ConsulKey]
			if ok {
				return fmt.Errorf("consul key %s already exist in monitor map", f.ConsulKey)
			}
			tm.monitorMap[SourceConsul][f.ConsulKey] = f
		}
	}
	return nil
}

func isKindSupported(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool, reflect.Int64, reflect.Float64, reflect.String:
		return true
	default:
		return false
	}
}
