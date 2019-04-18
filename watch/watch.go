package watch

import (
	"context"

	"github.com/taxibeat/harvester/config"
)

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
}

// Watch keys and update accordingly.
func (w *Watcher) Watch(ctx context.Context, cfg *config.Config, ww ...Item) error {

	return nil
}
