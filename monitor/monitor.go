// Package monitor handles config value monitoring and changing.
package monitor

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/beatlabs/harvester/change"
	"github.com/beatlabs/harvester/config"
	"github.com/beatlabs/harvester/log"
)

// Watcher interface definition.
type Watcher interface {
	Watch(ctx context.Context) (<-chan []change.Change, error)
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

// fanIn returns a channel that is the result of multiplexing the items from all
// the input channels. Output channel will be closed when all input channels are
// or when the context is done.
func fanIn(ctx context.Context, cs ...<-chan []change.Change) <-chan []change.Change {
	var wg sync.WaitGroup
	out := make(chan []change.Change)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan []change.Change) {
		defer wg.Done()
		for e := range c {
			select {
			case <-ctx.Done():
				return
			case out <- e:
			}
		}
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Monitor configuration changes by starting watchers per source.
func (m *Monitor) Monitor(ctx context.Context) error {
	watcherChans := make([]<-chan []change.Change, 0, len(m.ww))
	for _, w := range m.ww {
		c, err := w.Watch(ctx)
		if err != nil {
			return err
		}
		watcherChans = append(watcherChans, c)
	}
	ch := fanIn(ctx, watcherChans...)
	go m.monitor(ctx, ch)
	return nil
}

func (m *Monitor) monitor(ctx context.Context, ch <-chan []change.Change) {
	for {
		select {
		case <-ctx.Done():
			return
		case c := <-ch:
			m.applyChange(c)
		}
	}
}

func (m *Monitor) applyChange(cc []change.Change) {
	for _, c := range cc {
		mp, ok := m.mp[c.Source()]
		if !ok {
			log.Debugf("source %s not found", c.Source())
			continue
		}
		fld, ok := mp[c.Key()]
		if !ok {
			log.Debugf("key %s not found", c.Key())
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
