package consul

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taxibeat/harvester/change"
)

func TestNew(t *testing.T) {
	ii := []Item{Item{}}
	type args struct {
		addr string
		ii   []Item
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "success", args: args{addr: "xxx", ii: ii}, wantErr: false},
		{name: "empty address", args: args{addr: "", ii: ii}, wantErr: true},
		{name: "empty items", args: args{addr: "xxx", ii: nil}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.addr, "dc", "token", tt.args.ii...)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

func TestWatcher_Watch(t *testing.T) {
	w, err := New("xxx", "", "", Item{})
	require.NoError(t, err)
	chErr := make(chan error)
	type args struct {
		ctx context.Context
		ch  chan<- []*change.Change
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "missing context", args: args{}, wantErr: true},
		{name: "missing chan", args: args{ctx: context.Background()}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = w.Watch(tt.args.ctx, tt.args.ch, chErr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
