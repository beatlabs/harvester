package watch

import (
	"context"
	"errors"
	"fmt"

	"github.com/taxibeat/harvester/config"
)

type sourceMap map[config.Source]map[string]*config.Field

// Item definition.
type Item struct {
	Type string
	Key  string
}

// NewKeyItem creates a new key watch item for the watcher.
func NewKeyItem(key string) Item {
	return Item{Type: "key", Key: key}
}

// NewPrefixItem creates a prefix key watch item for the watcher.
func NewPrefixItem(key string) Item {
	return Item{Type: "keyprefix", Key: key}
}

// Watcher defines methods to watch for configuration changes.
type Watcher struct {
	cfg   *config.Config
	items []Item
	mp    sourceMap
}

// New constructor.
func New(cfg *config.Config, ii []Item) (*Watcher, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	if len(ii) == 0 {
		return nil, errors.New("items are empty")
	}
	mp, err := generateMap(cfg.Fields)
	if err != nil {
		return nil, err
	}
	return &Watcher{cfg: cfg, items: ii, mp: mp}, nil
}

// Watch keys and update accordingly.
func (w *Watcher) Watch(ctx context.Context, ww ...Item) error {

	//TODO: start watchers in a goroutine and exit

	return nil
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
