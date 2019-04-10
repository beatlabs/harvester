package harvester

import "errors"

type field struct {
	Name string
	Type string
}

// Monitor definition.
type Monitor struct {
	cfg interface{}
	ch  <-chan *Change
	mp  map[Source]map[string]*field
}

func NewMonitor(cfg interface{}, ch <-chan *Change) (*Monitor, error) {
	if cfg == nil {
		return nil, errors.New("configuration is nil")
	}
	if ch == nil {
		return nil, errors.New("change channel is nil")
	}
	m := &Monitor{ch: ch, cfg: cfg}
	if err := m.init(); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *Monitor) Monitor() error {
	//TODO: for range the channel for changes
	return nil
}

func (c *Monitor) init() error {
	//TODO: extract tags
	//TODO: create internal map
	return nil
}
