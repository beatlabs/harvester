//go:build integration
// +build integration

package redis

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetter_Get(t *testing.T) {
	client := redis.NewClient(&redis.Options{})

	const key = "key"
	const val = "value"

	getResult, err := client.Set(context.Background(), key, val, 0).Result()
	require.NoError(t, err)
	require.Equal(t, "OK", getResult)

	gtr, err := New(client)
	require.NoError(t, err)
	got, _, err := gtr.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, val, *got)

	delResult, err := client.Del(context.Background(), key).Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), delResult)
}
