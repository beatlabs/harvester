package harvester

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type field struct {
	Name    string
	Kind    reflect.Kind
	Version uint64
	Tag     reflect.StructTag
}

type tag struct {
	Src Source
	Key string
}

// Monitor definition.
type Monitor struct {
	cfg reflect.Value
	ch  <-chan *Change
	mp  map[Source]map[string]*field
}

// NewMonitor creates a new monitor.
func NewMonitor(cfg interface{}, ch <-chan *Change) (*Monitor, error) {
	if cfg == nil {
		return nil, errors.New("configuration is nil")
	}
	if ch == nil {
		return nil, errors.New("change channel is nil")
	}
	m := &Monitor{ch: ch, cfg: reflect.ValueOf(cfg).Elem()}
	if err := m.init(cfg); err != nil {
		return nil, err
	}
	return m, nil
}

// Monitor changes and apply them.
func (m *Monitor) Monitor() {
	for c := range m.ch {
		m.applyChange(c)
	}
}

func (m *Monitor) applyChange(c *Change) {
	mp, ok := m.mp[c.Src]
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

	err := m.setValue(fld.Name, c.Value, fld.Kind)
	if err != nil {
		logErrorf("failed to set value %s of kind %d on field %s", c.Value, fld.Kind, fld.Name)
		return
	}
	fld.Version = c.Version
}

func (m *Monitor) setValue(name, value string, kind reflect.Kind) error {
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

func (m *Monitor) init(cfg interface{}) error {
	tp := reflect.TypeOf(cfg)
	if tp.Kind() != reflect.Ptr {
		return errors.New("configuration should be a pointer type")
	}
	ff, err := getFields(tp.Elem())
	if err != nil {
		return err
	}
	err = validate(ff)
	if err != nil {
		return err
	}
	err = m.applySeedValues(ff)
	if err != nil {
		return err
	}
	err = m.applyEnvVarValues(ff)
	if err != nil {
		return err
	}
	//TODO: extract tags
	//TODO: create internal map
	return nil
}

func getFields(tp reflect.Type) ([]*field, error) {
	var ff []*field
	for i := 0; i < tp.NumField(); i++ {
		ff = append(ff, &field{
			Version: 0,
			Name:    tp.Field(i).Name,
			Kind:    tp.Field(i).Type.Kind(),
			Tag:     tp.Field(i).Tag,
		})
	}
	return ff, nil
}

func validate(ff []*field) error {
	sb := strings.Builder{}
	for _, f := range ff {
		if !isSupportedKind(f.Kind) {
			sb.WriteString(fmt.Sprintf("field %s of kind %d is not supported\n", f.Name, f.Kind))
		}
	}
	if sb.Len() == 0 {
		return nil
	}
	return errors.New(sb.String())
}

func (m *Monitor) applySeedValues(ff []*field) error {
	for _, f := range ff {
		value, ok := f.Tag.Lookup(string(SourceSeed))
		if !ok {
			continue
		}
		err := m.setValue(f.Name, value, f.Kind)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Monitor) applyEnvVarValues(ff []*field) error {
	for _, f := range ff {
		value, ok := f.Tag.Lookup(string(SourceEnv))
		if !ok {
			continue
		}
		value, ok = os.LookupEnv(value)
		if !ok {
			continue
		}
		err := m.setValue(f.Name, value, f.Kind)
		if err != nil {
			return err
		}
	}
	return nil
}

func isSupportedKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool, reflect.Int64, reflect.Float64, reflect.String:
		return true
	default:
		return false
	}
}
