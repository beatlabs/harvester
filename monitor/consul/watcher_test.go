package consul

import (
	"context"
	"testing"
	"time"

	"github.com/beatlabs/harvester/change"
	"github.com/beatlabs/harvester/config"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	ii := []Item{{}}
	type args struct {
		addr    string
		timeout time.Duration
		ii      []Item
	}
	tests := map[string]struct {
		args    args
		wantErr bool
	}{
		"success":                 {args: args{addr: "xxx", timeout: 1 * time.Second, ii: ii}, wantErr: false},
		"success default timeout": {args: args{addr: "xxx", timeout: 0, ii: ii}, wantErr: false},
		"empty address":           {args: args{addr: "", timeout: 1 * time.Second, ii: ii}, wantErr: true},
		"empty items":             {args: args{addr: "xxx", timeout: 1 * time.Second, ii: nil}, wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := New(tt.args.addr, "dc", "token", tt.args.timeout, tt.args.ii...)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

func TestWatcher_Watch(t *testing.T) {
	w, err := New("xxx", "", "", 0, Item{})
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

func TestItems(t *testing.T) {
	t.Run("NewKeyItem", func(t *testing.T) {
		item := NewKeyItem("key1")
		assert.Equal(t, Item{tp: "key", key: "key1"}, item)
	})
	t.Run("NewKeyItemWithPrefix", func(t *testing.T) {
		item := NewKeyItemWithPrefix("key1", "prefix1")
		assert.Equal(t, Item{tp: "key", key: "key1", prefix: "prefix1"}, item)
	})
	t.Run("NewPrefixItem", func(t *testing.T) {
		item := NewPrefixItem("prefix1")
		assert.Equal(t, Item{tp: "keyprefix", key: "prefix1"}, item)
	})
}

func TestWatcher_createKeyPlanWithPrefix(t *testing.T) {
	w, err := New("xxx", "dc", "token", 0, NewKeyItem("key"))
	require.NoError(t, err)

	ch := make(chan []*change.Change, 1)
	pl, err := w.createKeyPlanWithPrefix("key", "prefix", ch)
	require.NoError(t, err)
	require.NotNil(t, pl)

	pl.Handler(0, nil)
	assert.Empty(t, ch)

	pl.Handler(0, "not a kv pair")
	assert.Empty(t, ch)

	pl.Handler(0, &api.KVPair{Key: "prefix/key", Value: []byte("value"), ModifyIndex: 42})
	require.Len(t, ch, 1)
	changes := <-ch
	require.Len(t, changes, 1)
	assert.Equal(t, config.SourceConsul, changes[0].Source())
	assert.Equal(t, "key", changes[0].Key())
	assert.Equal(t, "value", changes[0].Value())
	assert.Equal(t, uint64(42), changes[0].Version())
}

func TestWatcher_createKeyPrefixPlan(t *testing.T) {
	w, err := New("xxx", "dc", "token", 0, NewPrefixItem("prefix"))
	require.NoError(t, err)

	ch := make(chan []*change.Change, 1)
	pl, err := w.createKeyPrefixPlan("prefix", ch)
	require.NoError(t, err)
	require.NotNil(t, pl)

	pl.Handler(0, nil)
	assert.Empty(t, ch)

	pl.Handler(0, "not kv pairs")
	assert.Empty(t, ch)

	pl.Handler(0, api.KVPairs{
		{Key: "prefix/one", Value: []byte("one"), ModifyIndex: 11},
		{Key: "prefix/two", Value: []byte("two"), ModifyIndex: 12},
	})
	require.Len(t, ch, 1)
	changes := <-ch
	require.Len(t, changes, 2)
	assert.Equal(t, "prefix/one", changes[0].Key())
	assert.Equal(t, "one", changes[0].Value())
	assert.Equal(t, uint64(11), changes[0].Version())
	assert.Equal(t, "prefix/two", changes[1].Key())
	assert.Equal(t, "two", changes[1].Value())
	assert.Equal(t, uint64(12), changes[1].Version())
}
