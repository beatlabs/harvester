package harvester

import (
	"errors"
	"reflect"
	"strconv"
)

type field struct {
	Name    string
	Type    reflect.Kind
	Version uint64
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
		//TODO: log
		return
	}
	fld, ok := mp[c.Key]
	if !ok {
		//TODO: log
		return
	}
	if fld.Version > c.Version {
		//TODO: log
		return
	}

	f := m.cfg.FieldByName(fld.Name)
	switch fld.Type {
	case reflect.Bool:
		b, err := strconv.ParseBool(c.Value)
		if err != nil {
			//TODO: log error
			return
		}
		f.SetBool(b)
	case reflect.String:
		f.SetString(c.Value)
	case reflect.Int64:
		v, err := strconv.ParseInt(c.Value, 10, 64)
		if err != nil {
			//TODO: log error
			return
		}
		f.SetInt(v)
	case reflect.Float64:
		v, err := strconv.ParseFloat(c.Value, 64)
		if err != nil {
			//TODO: log error
			return
		}
		f.SetFloat(v)
	default:
		//TODO: log error!!!!
		return
	}
	fld.Version = c.Version
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
