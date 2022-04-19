// Package redis handles the monitor capabilities of harvester using redis.
package redis

import (
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"errors"
	"time"

	"github.com/beatlabs/harvester/change"
	"github.com/beatlabs/harvester/config"
	"github.com/beatlabs/harvester/log"
	"github.com/go-redis/redis/v8"
)

// Watcher of Redis changes.
type Watcher struct {
	client       redis.UniversalClient
	keys         []string
	versions     []uint64
	hashes       []string
	pollInterval time.Duration
}

// New watcher.
func New(client redis.UniversalClient, pollInterval time.Duration, keys []string) (*Watcher, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	if pollInterval <= 0 {
		return nil, errors.New("poll interval should be a positive number")
	}
	if len(keys) == 0 {
		return nil, errors.New("keys are empty")
	}

	return &Watcher{
		client:       client,
		keys:         keys,
		versions:     make([]uint64, len(keys)),
		hashes:       make([]string, len(keys)),
		pollInterval: pollInterval,
	}, nil
}

// Watch keys and changes.
func (w *Watcher) Watch(ctx context.Context) (<-chan []change.Change, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	ch := make(chan []change.Change)
	go func(pollInterval time.Duration) {
		defer close(ch)
		tickerStats := time.NewTicker(pollInterval)
		defer tickerStats.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tickerStats.C:
				if cgs := w.getValues(ctx); len(cgs) > 0 {
					ch <- cgs
				}
			}
		}
	}(w.pollInterval)
	return ch, nil
}

func (w *Watcher) getValues(ctx context.Context) []change.Change {
	values := make([]*string, len(w.keys))
	for i, key := range w.keys {
		strCmd := w.client.Get(ctx, key)
		if strCmd == nil {
			log.Errorf("failed to get value for key %s: nil strCmd", key)
			continue
		}
		if strCmd.Err() != nil {
			if !errors.Is(strCmd.Err(), redis.Nil) {
				log.Errorf("failed to get value for key %s: %s", key, strCmd.Err())
			}
			continue
		}
		val := strCmd.Val()
		values[i] = &val
	}

	changes := make([]change.Change, 0, len(w.keys))

	for i, key := range w.keys {
		if values[i] == nil {
			continue
		}

		value := *values[i]
		hash := w.hash(value)
		if hash == w.hashes[i] {
			continue
		}

		w.versions[i]++
		w.hashes[i] = hash

		cg := change.New(config.SourceRedis, key, value, w.versions[i])
		if cg != nil {
			changes = append(changes, *cg)
		}
	}

	if len(changes) == 0 {
		return nil
	}

	return changes
}

func (w *Watcher) hash(value string) string {
	hash := md5.Sum([]byte(value)) // nolint:gosec
	return hex.EncodeToString(hash[:])
}
