//go:build integration
// +build integration

package redis

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/beatlabs/harvester/change"
	"github.com/beatlabs/harvester/config"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatch(t *testing.T) {
	client := redis.NewClient(&redis.Options{})

	const (
		key1 = "key1"
		val1 = "value1"

		key2 = "key2"
		val2 = "value2"

		key3 = "key3"
		val3 = "value3"
	)

	defer func() {
		del(t, client, key1)
		del(t, client, key2)
		del(t, client, key3)
	}()

	w, err := New(client, 5*time.Millisecond, []string{key1, key2, key3})
	require.NoError(t, err)
	require.NotNil(t, w)

	// Initial values, set even before watching - not all keys have a value
	set(t, client, key1, val1)
	set(t, client, key2, val1)

	// Start watching
	ctx, cnl := context.WithCancel(context.Background())
	ch, err := w.Watch(ctx)
	require.NoError(t, err)
	// listens to the changes and append to the found slice
	found := make([]change.Change, 0)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for cc := range ch {
			found = append(found, cc...)
		}
	}()

	// First values update
	time.Sleep(1 * time.Second)
	set(t, client, key1, val1) // Same value
	set(t, client, key2, val2)
	set(t, client, key3, val1) // First value for this key

	// Second values update
	time.Sleep(1 * time.Second)
	set(t, client, key1, val1) // Same value
	set(t, client, key2, val1) // Second value - same as the initial value
	set(t, client, key3, val3)

	time.Sleep(1 * time.Second)
	cnl()
	wg.Wait()

	toValue := func(c *change.Change) change.Change {
		return *c
	}
	expected := []change.Change{
		// Initial values
		toValue(change.New(config.SourceRedis, key1, val1, 1)),
		toValue(change.New(config.SourceRedis, key2, val1, 1)),
		// First update
		toValue(change.New(config.SourceRedis, key2, val2, 2)),
		toValue(change.New(config.SourceRedis, key3, val1, 1)),
		// Second update
		toValue(change.New(config.SourceRedis, key2, val1, 3)),
		toValue(change.New(config.SourceRedis, key3, val3, 2)),
	}

	assert.Equal(t, expected, found)
}

func set(t *testing.T, client redis.UniversalClient, key string, val string) {
	getResult, err := client.Set(context.Background(), key, val, 0).Result()
	require.NoError(t, err)
	require.Equal(t, "OK", getResult)
}

func del(t *testing.T, client redis.UniversalClient, key string) {
	delResult, err := client.Del(context.Background(), key).Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), delResult)
}
