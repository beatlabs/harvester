package monitor

import (
	"context"
	"errors"
	"fmt"

	"github.com/taxibeat/harvester/change"
	"github.com/taxibeat/harvester/config"
	"github.com/taxibeat/harvester/log"
)

// Watcher interface definition.
type Watcher interface {
	Watch(ctx context.Context, ch <-chan *change.Change, ii []Item) error
}

// Item definition.
type Item struct {
	Source config.Source
	Type   string
	Key    string
}

// NewKeyItem creates a new key watch item for the watcher.
func NewKeyItem(src config.Source, key string) Item {
	return Item{Type: "key", Key: key}
}

// NewPrefixItem creates a prefix key watch item for the watcher.
func NewPrefixItem(src config.Source, key string) Item {
	return Item{Type: "keyprefix", Key: key}
}

type sourceMap map[config.Source]map[string]*config.Field

// Monitor for configuration changes.
type Monitor struct {
	cfg   *config.Config
	items []Item
	mp    sourceMap
	ww    map[config.Source]Watcher
}

// New constructor.
func New(cfg *config.Config, ii []Item, ww map[config.Source]Watcher) (*Monitor, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	if len(ii) == 0 {
		return nil, errors.New("items are empty")
	}
	if len(ww) == 0 {
		return nil, errors.New("watchers are empty")
	}
	mp, err := generateMap(cfg.Fields)
	if err != nil {
		return nil, err
	}
	return &Monitor{cfg: cfg, items: ii, mp: mp}, nil
}

func generateMap(ff []*config.Field) (sourceMap, error) {
	mp := make(sourceMap)
	for _, f := range ff {
		key, ok := f.Sources[config.SourceConsul]
		if !ok {
			continue
		}
		_, ok = mp[config.SourceConsul]
		if !ok {
			mp[config.SourceConsul] = map[string]*config.Field{key: f}
		} else {
			_, ok := mp[config.SourceConsul][key]
			if ok {
				return nil, fmt.Errorf("consul key %s already exist in monitor map", key)
			}
			mp[config.SourceConsul][key] = f
		}
	}
	return mp, nil
}

// Monitor configuration changes by starting watchers per source.
func (m *Monitor) Monitor(ctx context.Context) error {
	ch := make(chan *change.Change)
	go m.monitor(ctx, ch)

	for src, ii := range generateSourceItems(m.items) {
		wtc, ok := m.ww[src]
		if !ok {
			return fmt.Errorf("source watcher %s not available", src)
		}
		err := wtc.Watch(ctx, ch, ii)
		if err != nil {
			return err
		}
	}

	return nil
}

func generateSourceItems(ii []Item) map[config.Source][]Item {
	sourceItems := make(map[config.Source][]Item)
	for _, i := range ii {
		items, ok := sourceItems[i.Source]
		if !ok {
			items = []Item{i}
		} else {
			items = append(items, i)
		}
		sourceItems[i.Source] = items
	}
	return sourceItems
}

func (m *Monitor) monitor(ctx context.Context, ch <-chan *change.Change) {
	for {
		select {
		case <-ctx.Done():
			return
		case c := <-ch:
			m.applyChange(c)
		}
	}
}

func (m *Monitor) applyChange(c *change.Change) {
	mp, ok := m.mp[c.Source()]
	if !ok {
		log.Warnf("source %s not found", c.Source())
		return
	}
	fld, ok := mp[c.Key()]
	if !ok {
		log.Warnf("key %s not found", c.Key)
		return
	}
	if fld.Version > c.Version() {
		log.Warnf("version %d is older than %d", c.Version, fld.Version)
		return
	}
	err := m.cfg.Set(fld.Name, c.Value(), fld.Kind)
	if err != nil {
		log.Errorf("failed to set value %s of kind %d on field %s from source %s", c.Value, fld.Kind, fld.Name, c.Source())
		return
	}
	fld.Version = c.Version()
}
