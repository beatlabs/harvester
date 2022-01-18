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

func TestNew(t *testing.T) {
	type args struct {
		client       redis.UniversalClient
		pollInterval time.Duration
		keys         []string
	}
	tests := map[string]struct {
		args        args
		expectedErr string
	}{
		"success":               {args: args{client: &redis.Client{}, pollInterval: 1 * time.Second, keys: []string{"1"}}},
		"client nil":            {args: args{client: nil, pollInterval: 1 * time.Second, keys: []string{"1"}}, expectedErr: "client is nil"},
		"poll interval invalid": {args: args{client: &redis.Client{}, pollInterval: 0 * time.Second, keys: []string{"1"}}, expectedErr: "poll interval should be a positive number"},
		"keys are missing":      {args: args{client: &redis.Client{}, pollInterval: 1 * time.Second, keys: nil}, expectedErr: "keys are empty"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := New(tt.args.client, tt.args.pollInterval, tt.args.keys)
			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

func TestWatcher_Watch(t *testing.T) {
	w, err := New(&redis.Client{}, time.Second, []string{"1"})
	require.NoError(t, err)
	type args struct {
		ctx context.Context
		ch  chan<- []*change.Change
	}
	tests := map[string]struct {
		args    args
		wantErr bool
	}{
		"missing context": {args: args{}, wantErr: true},
		"missing chan":    {args: args{ctx: context.Background()}, wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err = w.Watch(tt.args.ctx, tt.args.ch)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWatcher_Versioning(t *testing.T) {
	client := (&clientStub{t: t}).
		WithValues("val1.1", "val2.1", "val3.1"). // Initial values
		WithValues("val1.1", "val2.2", "val3.2"). // Only keys 2 and 3 are updated
		WithValues("val1.1", "val2.1", "val3.2")  // Only 2 is updated, to its previous value

	expected := [][]*change.Change{
		{
			change.New(config.SourceRedis, "key1", "val1.1", 1),
			change.New(config.SourceRedis, "key2", "val2.1", 1),
			change.New(config.SourceRedis, "key3", "val3.1", 1),
		},
		{
			change.New(config.SourceRedis, "key2", "val2.2", 2),
			change.New(config.SourceRedis, "key3", "val3.2", 2),
		},
		{
			change.New(config.SourceRedis, "key2", "val2.1", 3),
		},
	}

	w, err := New(client, 1*time.Millisecond, []string{"key1", "key2", "key3"})
	require.NoError(t, err)
	assert.Equal(t, []uint64{0, 0, 0}, w.versions)
	assert.Equal(t, []string{"", "", ""}, w.hashes)

	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan []*change.Change, 10)
	err = w.Watch(ctx, ch)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	cancel()

	found := make([][]*change.Change, 0)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for {
			select {
			case cc := <-ch:
				if len(cc) == 0 {
					break
				}
				found = append(found, cc)
			default:
				wg.Done()
				return
			}
		}
	}()
	wg.Wait()

	assert.Equal(t, expected, found)
}

type clientStub struct {
	t *testing.T
	*redis.Client

	cmds []*redis.SliceCmd
}

func (c *clientStub) WithValues(values ...interface{}) *clientStub {
	c.cmds = append(c.cmds, redis.NewSliceResult(values, nil))
	return c
}

func (c *clientStub) MGet(_ context.Context, keys ...string) *redis.SliceCmd {
	if len(c.cmds) == 0 {
		return redis.NewSliceResult(make([]interface{}, len(keys)), nil)
	}
	shifted := c.cmds[0]
	c.cmds = c.cmds[1:]
	return shifted
}
