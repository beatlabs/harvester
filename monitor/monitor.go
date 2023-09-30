// Package monitor handles config value monitoring and changing.
package monitor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/beatlabs/harvester/change"
	"github.com/beatlabs/harvester/config"
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
		for source, val := range f.Sources() {
			if source == config.SourceSeed {
				continue
			}
			_, ok := mp[source]
			if !ok {
				mp[source] = map[string]*config.Field{val: f}
			} else {
				_, ok := mp[source][val]
				if ok {
					return nil, fmt.Errorf("%s key %s already exists in monitor map", source, val)
				}
				mp[source][val] = f
			}
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
			slog.Debug("source not found", "source", c.Source())
			continue
		}
		fld, ok := mp[c.Key()]
		if !ok {
			slog.Debug("key not found", "key", c.Key())
			continue
		}

		err := fld.Set(c.Value(), c.Version())
		if err != nil {
			slog.Error("failed to set value", "value", c.Value(), "type", fld.Type(), "name", fld.Name(),
				"source", c.Source(), "err", err)
			continue
		}
	}
}
