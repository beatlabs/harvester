//go:build integration

package redis

import (
	"context"
	"testing"
	"time"

	"github.com/beatlabs/harvester/change"
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

	set(t, client, key1, val1)
	set(t, client, key2, val2)
	set(t, client, key3, val3)
	defer func() {
		del(t, client, key1)
		del(t, client, key2)
		del(t, client, key3)
	}()

	ch := make(chan []*change.Change)
	w, err := New(client, 10*time.Millisecond, []string{key1, key2, key3})
	require.NoError(t, err)
	require.NotNil(t, w)
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()
	err = w.Watch(ctx, ch)
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		cc := <-ch
		for _, cng := range cc {
			switch cng.Key() {
			case key1:
				assert.Equal(t, val1, cng.Value())
			case key2:
				assert.Equal(t, val2, cng.Value())
			case key3:
				assert.Equal(t, val3, cng.Value())
			default:
				assert.Fail(t, "key invalid", cng.Key())
			}
			assert.True(t, cng.Version() == 0)
		}
	}
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
