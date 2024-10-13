package consul

import (
	"context"
	"testing"
	"time"

	"github.com/beatlabs/harvester/change"
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
