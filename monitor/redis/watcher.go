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
	sleep        func(context.Context, time.Duration) bool
}

const maxBackoff = 30 * time.Second

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
		sleep:        sleepContext,
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
	interval := w.pollInterval
	consecutiveErrors := 0

	for {
		if !w.sleep(ctx, interval) {
			return
		}

		if w.getValues(ctx, ch) {
			consecutiveErrors = 0
			interval = w.pollInterval
			continue
		}

		consecutiveErrors++
		interval = w.backoffInterval(consecutiveErrors)
	}
}

func (w *Watcher) getValues(ctx context.Context, ch chan<- []*change.Change) bool {
	// Use MGET to fetch all values in a single round-trip
	sliceCmd := w.client.MGet(ctx, w.keys...)
	if sliceCmd == nil {
		slog.Error("failed to get values", "err", "nil sliceCmd")
		return false
	}
	if sliceCmd.Err() != nil {
		slog.Error("failed to get values", "err", sliceCmd.Err())
		return false
	}

	results := sliceCmd.Val()
	if len(results) != len(w.keys) {
		slog.Error("mget returned unexpected number of results", "expected", len(w.keys), "got", len(results))
		return false
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
		return true
	}

	ch <- changes
	return true
}

func (w *Watcher) hash(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func (w *Watcher) backoffInterval(consecutiveErrors int) time.Duration {
	if consecutiveErrors <= 0 {
		return w.pollInterval
	}

	interval := w.pollInterval
	for range consecutiveErrors {
		if interval >= maxBackoff/2 {
			return maxBackoff
		}
		interval *= 2
	}

	return interval
}

func sleepContext(ctx context.Context, interval time.Duration) bool {
	timer := time.NewTimer(interval)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
