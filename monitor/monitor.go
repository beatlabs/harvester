package monitor

import (
	"context"
	"errors"
	"fmt"

	"github.com/beatlabs/harvester/change"
	"github.com/beatlabs/harvester/config"
	"github.com/beatlabs/harvester/log"
)

// Watcher interface definition.
type Watcher interface {
	Watch(ctx context.Context, ch chan<- []*change.Change) error
}

type sourceMap map[config.Source]map[string]*config.Field

// Monitor for configuration changes.
type Monitor struct {
	cfg *config.Config
	mp  sourceMap
	ww  []Watcher
}

// New constructor.
func New(cfg *config.Config, ww ...Watcher) (*Monitor, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	if len(ww) == 0 {
		return nil, errors.New("watchers are empty")
	}
	mp, err := generateMap(cfg.Fields)
	if err != nil {
		return nil, err
	}
	return &Monitor{cfg: cfg, mp: mp, ww: ww}, nil
}

func generateMap(ff []*config.Field) (sourceMap, error) {
	mp := make(sourceMap)
	for _, f := range ff {
		key, ok := f.Sources()[config.SourceConsul]
		if !ok {
			continue
		}
		_, ok = mp[config.SourceConsul]
		if !ok {
			mp[config.SourceConsul] = map[string]*config.Field{key: f}
		} else {
			_, ok := mp[config.SourceConsul][key]
			if ok {
				return nil, fmt.Errorf("consul key %s already exists in monitor map", key)
			}
			mp[config.SourceConsul][key] = f
		}
	}
	return mp, nil
}

// Monitor configuration changes by starting watchers per source.
func (m *Monitor) Monitor(ctx context.Context) error {
	ch := make(chan []*change.Change)
	go m.monitor(ctx, ch)

	for _, w := range m.ww {
		err := w.Watch(ctx, ch)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Monitor) monitor(ctx context.Context, ch <-chan []*change.Change) {
	for {
		select {
		case <-ctx.Done():
			return
		case c := <-ch:
			m.applyChange(c)
		}
	}
}

func (m *Monitor) applyChange(cc []*change.Change) {
	for _, c := range cc {
		mp, ok := m.mp[c.Source()]
		if !ok {
			log.Warnf("source %s not found", c.Source())
			continue
		}
		fld, ok := mp[c.Key()]
		if !ok {
			log.Warnf("key %s not found", c.Key())
			continue
		}

		err := fld.Set(c.Value(), c.Version())
		if err != nil {
			log.Errorf("failed to set value %s of type %s on field %s from source %s: %v",
				c.Value(), fld.Type(), fld.Name(), c.Source(), err)
			continue
		}
	}
}
