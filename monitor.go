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
	return &Monitor{ch: ch, cfg: cfg}, nil
}

func (c *Monitor) Monitor() error {

}
