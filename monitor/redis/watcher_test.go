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
