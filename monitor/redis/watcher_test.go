package redis

import (
	"context"
	"errors"
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
		ctx    context.Context
		cancel context.CancelFunc
	}
	withCancel := func() args {
		ctx, cnl := context.WithCancel(context.Background())
		return args{ctx: ctx, cancel: cnl}
	}
	tests := map[string]struct {
		args    args
		wantErr bool
	}{
		"success":         {args: withCancel()},
		"missing context": {args: args{}, wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ch, err := w.Watch(tt.args.ctx)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, ch)
			} else {
				assert.NoError(t, err)
				// by cancelling the context the channel should close, if it's
				// not closed within the timeout of 1 second the test will fail
				tt.args.cancel()
				tickerStats := time.NewTicker(1 * time.Second)
				for {
					select {
					case _, opened := <-ch:
						if !opened {
							// success (channel closed before timeout)
							return
						}
					case <-tickerStats.C:
						assert.Fail(t, "channel should have been closed")
					}
				}
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
	toValue := func(ch *change.Change) change.Change {
		return *ch
	}
	expected := [][]change.Change{
		{
			toValue(change.New(config.SourceRedis, "key1", "val1.1", 1)),
			toValue(change.New(config.SourceRedis, "key2", "val2.1", 1)),
			toValue(change.New(config.SourceRedis, "key3", "val3.1", 1)),
		},
		{
			toValue(change.New(config.SourceRedis, "key2", "val2.2", 2)),
			toValue(change.New(config.SourceRedis, "key3", "val3.2", 2)),
		},
	}

	client := clientStub{t: t, m: sync.Mutex{}, watchedKeys: watchedKeys}
	for _, mv := range redisInternalState {
		client.AppendMockValues(mv)
	}
	pollingIntervalMillis := 5
	w, err := New(&client, time.Duration(pollingIntervalMillis)*time.Millisecond, []string{"key1", "key2", "key3"})
	require.NoError(t, err)
	assert.Equal(t, []uint64{0, 0, 0}, w.versions)
	assert.Equal(t, []string{"", "", ""}, w.hashes)

	ctx, cancel := context.WithCancel(context.Background())

	ch, err := w.Watch(ctx)
	assert.NoError(t, err)

	found := make([][]change.Change, 0)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for cc := range ch {
			found = append(found, cc)
		}
	}()
	intervalForAllPollingToHappen := time.Duration(pollingIntervalMillis*len(redisInternalState)) * time.Millisecond
	intervalForAllPollingToHappen += 10 * time.Millisecond // to be sure
	time.Sleep(intervalForAllPollingToHappen)
	cancel()
	wg.Wait()

	assert.Equal(t, expected, found)
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

func (c *clientStub) rollInternalRedisState() {
	// replace redis virtual state every len(watchedKeys) calls to Get
	if len(c.keyToCmd) > 0 && (c.internalGetCalls)%len(c.watchedKeys) == 0 {
		c.keyToCmd = c.keyToCmd[1:]
	}
}
