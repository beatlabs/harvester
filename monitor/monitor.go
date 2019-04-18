package monitor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"sync"

	"github.com/taxibeat/harvester/change"
	"github.com/taxibeat/harvester/log"
)

// GetValueFunc function definition for getting a value for a key from a source.
type GetValueFunc func(key string) (string, error)

type field struct {
	Name      string
	Kind      reflect.Kind
	Version   uint64
	SeedValue string
	EnvVarKey string
	ConsulKey string
}

// Monitor definition.
type Monitor struct {
	ch         <-chan []*change.Change
	monitorMap map[change.Source]map[string]*field
	consulGet  GetValueFunc
	name       string
	sync.Mutex
	cfg reflect.Value
}

// New creates a new monitor.
func New(cfg interface{}, ch <-chan []*change.Change, consulGet GetValueFunc) (*Monitor, error) {
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
	m := &Monitor{
		ch:         ch,
		cfg:        reflect.ValueOf(cfg).Elem(),
		monitorMap: make(map[change.Source]map[string]*field),
		consulGet:  consulGet,
		name:       tp.Name(),
	}
	if err := m.setup(tp); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Monitor) setup(tp reflect.Type) error {
	ff, err := getFields(tp.Elem())
	if err != nil {
		return err
	}
	err = m.applyInitialValues(ff)
	if err != nil {
		return err
	}
	err = m.createMonitorMap(ff)
	if err != nil {
		return err
	}
	return nil
}

func (m *Monitor) applyInitialValues(ff []*field) error {
	for _, f := range ff {
		if f.SeedValue != "" {
			err := m.setValue(f.Name, f.SeedValue, f.Kind)
			if err != nil {
				return err
			}
		}
		if f.EnvVarKey != "" {
			value, ok := os.LookupEnv(f.EnvVarKey)
			if !ok {
				continue
			}
			err := m.setValue(f.Name, value, f.Kind)
			if err != nil {
				return err
			}
		}
		if f.ConsulKey != "" {
			if m.consulGet == nil {
				return errors.New("consul getter required")
			}
			value, err := m.consulGet(f.ConsulKey)
			if err != nil {
				return err
			}
			err = m.setValue(f.Name, value, f.Kind)
			if err != nil {
				return err
			}
		}
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
		value, ok := fld.Tag.Lookup(string(change.SourceSeed))
		if ok {
			f.SeedValue = value
		}
		value, ok = fld.Tag.Lookup(string(change.SourceEnv))
		if ok {
			f.EnvVarKey = value
		}
		value, ok = fld.Tag.Lookup(string(change.SourceConsul))
		if ok {
			f.ConsulKey = value
		}
		ff = append(ff, f)
	}
	return ff, nil
}

func (m *Monitor) createMonitorMap(ff []*field) error {
	for _, f := range ff {
		if f.ConsulKey == "" {
			continue
		}
		_, ok := m.monitorMap[change.SourceConsul]
		if !ok {
			m.monitorMap[change.SourceConsul] = map[string]*field{f.ConsulKey: f}
		} else {
			_, ok := m.monitorMap[change.SourceConsul][f.ConsulKey]
			if ok {
				return fmt.Errorf("consul key %s already exist in monitor map", f.ConsulKey)
			}
			m.monitorMap[change.SourceConsul][f.ConsulKey] = f
		}
	}
	return nil
}

func (m *Monitor) setValue(name, value string, kind reflect.Kind) error {
	m.Lock()
	defer m.Unlock()
	f := m.cfg.FieldByName(name)
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

func isKindSupported(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool, reflect.Int64, reflect.Float64, reflect.String:
		return true
	default:
		return false
	}
}

// Monitor changes and apply them.
func (m *Monitor) Monitor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Infof("exiting configuration monitor for %s", m.name)
			return
		case cc := <-m.ch:
			for _, c := range cc {
				m.applyChange(c)
			}
		}
	}
}

func (m *Monitor) applyChange(c *change.Change) {
	mp, ok := m.monitorMap[c.Src]
	if !ok {
		log.Warnf("source %s not found", c.Src)
		return
	}
	fld, ok := mp[c.Key]
	if !ok {
		log.Warnf("key %s not found", c.Key)
		return
	}
	if fld.Version > c.Version {
		log.Warnf("version %d is older than %d", c.Version, fld.Version)
		return
	}

	err := m.setValue(fld.Name, c.Value, fld.Kind)
	if err != nil {
		log.Errorf("failed to set value %s of kind %d on field %s from source %s", c.Value, fld.Kind, fld.Name, c.Src)
		return
	}
	fld.Version = c.Version
}
