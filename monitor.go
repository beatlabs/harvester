package harvester

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type field struct {
	Name    string
	Kind    reflect.Kind
	Version uint64
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

	//TODO: check for unsupported kind
	//TODO: extract tags
	//TODO: create internal map
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
