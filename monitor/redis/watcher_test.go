package redis

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/beatlabs/harvester/change"
	"github.com/beatlabs/harvester/config"
	"github.com/redis/go-redis/v9"
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
				require.EqualError(t, err, tt.expectedErr)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
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
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWatcher_Versioning(t *testing.T) {
	watchedKeys := []string{"key1", "key2", "key3"}
	// each element represent the state of redis server at each subsequent poll
	redisInternalState := []map[string]interface{}{
		{
			// watch triggers change in key1, key2 and key3
			"key1": "val1.1",
			"key2": "val2.1",
			"key3": "val3.1",
		},
		{
			// watch triggers change in key2 and key3
			"key1": "val1.1", // no change
			"key2": "val2.2", // change
			"key3": "val3.2", // change
		},
		{
			// whole watch does not trigger change (but errors will be logged as != redis.Nil)
			"key1": errors.New("error key1"), // error occurred -> no change should be triggered
			"key2": errors.New("error key2"), // error occurred -> no change should be triggered
			"key3": errors.New("error key3"), // error occurred -> no change should be triggered
		},
		{
			// watch does not trigger change or log because key1 watch will lead to redis.Nil
			"key2": "val2.2", // no change from previous
			"key3": "val3.2", // no change from previous
		},
		{
			// all subscribed keys deleted -> do not trigger change or log because redis.Nil is ignored
			"key4": "val4.1", // no change -> not subscribed to this key
		},
		{
			// all subscribed keys deleted -> do not trigger change or log because redis.Nil is ignored
			"key4": "val4.2", // no change -> not subscribed to this key
		},
		{
			// all subscribed keys deleted -> do not trigger change but triggers
			// log because error is different than redis.Nil is ignored
			"key2": errors.New("error"), // error occurred -> no change should be triggered
		},
	}
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
	}

	client := clientStub{t: t, m: sync.Mutex{}, watchedKeys: watchedKeys}
	for _, mv := range redisInternalState {
		client.AppendMockValues(mv)
	}
	w, err := New(&client, 5*time.Millisecond, []string{"key1", "key2", "key3"})
	require.NoError(t, err)
	assert.Equal(t, []uint64{0, 0, 0}, w.versions)
	assert.Equal(t, []string{"", "", ""}, w.hashes)

	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan []*change.Change, 10)
	err = w.Watch(ctx, ch)
	require.NoError(t, err)

	// time for completing all the polling for the different states
	time.Sleep(100 * time.Millisecond)

	found := make([][]*change.Change, 0)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case cc := <-ch:
				found = append(found, cc)
			default:
				return
			}
		}
	}()
	cancel()
	wg.Wait()

	assert.Equal(t, expected, found)
}

func TestWatcher_GetValues_EdgeCases(t *testing.T) {
	ctx := context.Background()
	ch := make(chan []*change.Change, 1)

	t.Run("MGet returns nil command", func(t *testing.T) {
		c := &failClientStub{
			mGetFn: func(ctx context.Context, keys ...string) *redis.SliceCmd {
				return nil
			},
		}
		w, _ := New(c, time.Second, []string{"key1"})
		w.getValues(ctx, ch)
		assert.Empty(t, ch)
	})

	t.Run("MGet returns error", func(t *testing.T) {
		c := &failClientStub{
			mGetFn: func(ctx context.Context, keys ...string) *redis.SliceCmd {
				return redis.NewSliceResult(nil, errors.New("boom"))
			},
		}
		w, _ := New(c, time.Second, []string{"key1"})
		w.getValues(ctx, ch)
		assert.Empty(t, ch)
	})

	t.Run("MGet returns unexpected number of results", func(t *testing.T) {
		c := &failClientStub{
			mGetFn: func(ctx context.Context, keys ...string) *redis.SliceCmd {
				return redis.NewSliceResult([]interface{}{"val1"}, nil)
			},
		}
		w, _ := New(c, time.Second, []string{"key1", "key2"})
		w.getValues(ctx, ch)
		assert.Empty(t, ch)
	})

	t.Run("MGet returns invalid value type", func(t *testing.T) {
		c := &failClientStub{
			mGetFn: func(ctx context.Context, keys ...string) *redis.SliceCmd {
				return redis.NewSliceResult([]interface{}{123}, nil)
			},
		}
		w, _ := New(c, time.Second, []string{"key1"})
		w.getValues(ctx, ch)
		assert.Empty(t, ch)
	})
}

type failClientStub struct {
	*redis.Client
	mGetFn func(ctx context.Context, keys ...string) *redis.SliceCmd
}

func (c *failClientStub) MGet(ctx context.Context, keys ...string) *redis.SliceCmd {
	if c.mGetFn != nil {
		return c.mGetFn(ctx, keys...)
	}
	return nil
}

type clientStub struct {
	t *testing.T
	*redis.Client
	m                sync.Mutex
	watchedKeys      []string
	internalGetCalls int

	keyToCmd []map[string]*redis.StringCmd
}

func (c *clientStub) AppendMockValues(values map[string]interface{}) *clientStub {
	c.m.Lock()
	defer c.m.Unlock()
	mockResp := make(map[string]*redis.StringCmd)
	for k, v := range values {
		if v == nil {
			mockResp[k] = nil
			continue
		}
		if s, ok := v.(string); ok {
			mockResp[k] = redis.NewStringResult(s, nil)
			continue
		}
		if e, ok := v.(error); ok {
			mockResp[k] = redis.NewStringResult("", e)
			continue
		}
		mockResp[k] = redis.NewStringResult("", errors.New("Unknown Error"))
	}

	if c.keyToCmd == nil {
		c.keyToCmd = make([]map[string]*redis.StringCmd, 0)
	}
	c.keyToCmd = append(c.keyToCmd, mockResp)
	return c
}

func (c *clientStub) Get(_ context.Context, key string) *redis.StringCmd {
	c.m.Lock()
	defer c.m.Unlock()
	c.internalGetCalls++
	defer c.rollInternalRedisState()
	if len(c.keyToCmd) == 0 {
		return redis.NewStringResult("", redis.Nil)
	}
	shifted := c.keyToCmd[0]
	if v, ok := shifted[key]; ok {
		return v
	}

	return redis.NewStringResult("", redis.Nil)
}

func (c *clientStub) MGet(_ context.Context, keys ...string) *redis.SliceCmd {
	c.m.Lock()
	defer c.m.Unlock()
	c.internalGetCalls++
	defer c.rollInternalRedisStateMGet()

	if len(c.keyToCmd) == 0 {
		// Return slice of nils for all keys
		results := make([]interface{}, len(keys))
		return redis.NewSliceResult(results, nil)
	}

	shifted := c.keyToCmd[0]
	results := make([]interface{}, len(keys))

	for i, key := range keys {
		if v, ok := shifted[key]; ok {
			switch {
			case v == nil:
				results[i] = nil
			case v.Err() != nil:
				// For errors, we skip this key (MGet doesn't return individual errors)
				results[i] = nil
			default:
				results[i] = v.Val()
			}
		} else {
			results[i] = nil
		}
	}

	return redis.NewSliceResult(results, nil)
}

func (c *clientStub) rollInternalRedisState() {
	// replace redis virtual state every len(watchedKeys) calls to Get
	if len(c.keyToCmd) > 0 && (c.internalGetCalls)%len(c.watchedKeys) == 0 {
		c.keyToCmd = c.keyToCmd[1:]
	}
}

func (c *clientStub) rollInternalRedisStateMGet() {
	// replace redis virtual state after each MGet call (since it gets all keys at once)
	if len(c.keyToCmd) > 0 {
		c.keyToCmd = c.keyToCmd[1:]
	}
}
