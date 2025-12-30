// Package redis handles the monitor capabilities of harvester using redis.
package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"github.com/beatlabs/harvester/change"
	"github.com/beatlabs/harvester/config"
	"github.com/redis/go-redis/v9"
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
func (w *Watcher) Watch(ctx context.Context, ch chan<- []*change.Change) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if ch == nil {
		return errors.New("change channel is nil")
	}

	go w.monitor(ctx, ch)
	return nil
}

func (w *Watcher) monitor(ctx context.Context, ch chan<- []*change.Change) {
	tickerStats := time.NewTicker(w.pollInterval)
	defer tickerStats.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tickerStats.C:
			w.getValues(ctx, ch)
		}
	}
}

func (w *Watcher) getValues(ctx context.Context, ch chan<- []*change.Change) {
	// Use MGET to fetch all values in a single round-trip
	sliceCmd := w.client.MGet(ctx, w.keys...)
	if sliceCmd == nil {
		slog.Error("failed to get values", "err", "nil sliceCmd")
		return
	}
	if sliceCmd.Err() != nil {
		slog.Error("failed to get values", "err", sliceCmd.Err())
		return
	}

	results := sliceCmd.Val()
	if len(results) != len(w.keys) {
		slog.Error("mget returned unexpected number of results", "expected", len(w.keys), "got", len(results))
		return
	}

	changes := make([]*change.Change, 0, len(w.keys))

	for i, key := range w.keys {
		// MGet returns nil for keys that don't exist
		if results[i] == nil {
			continue
		}

		value, ok := results[i].(string)
		if !ok {
			slog.Error("failed to convert value to string", "key", key, "value", results[i])
			continue
		}

		hash := w.hash(value)
		if hash == w.hashes[i] {
			continue
		}

		w.versions[i]++
		w.hashes[i] = hash

		changes = append(changes, change.New(config.SourceRedis, key, value, w.versions[i]))
	}

	if len(changes) == 0 {
		return
	}

	ch <- changes
}

func (w *Watcher) hash(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}
