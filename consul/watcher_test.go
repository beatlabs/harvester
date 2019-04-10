package consul

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taxibeat/harvester"
)

func TestNew(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				address: "addr",
			},
			wantErr: false,
		},
		{
			name: "missing address",
			args: args{
				address: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.address, "datacenter", "token", false)
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
	ch := make(chan *harvester.Change)
	chErr := make(chan error)
	ww := []WatchItem{}
	type args struct {
		ch    chan<- *harvester.Change
		chErr chan<- error
		ww    []WatchItem
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "missing channel", args: args{ch: nil, chErr: chErr, ww: ww}, wantErr: true},
		{name: "missing error channel", args: args{ch: ch, chErr: nil, ww: ww}, wantErr: true},
		{name: "missing watch items", args: args{ch: ch, chErr: chErr, ww: nil}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := New("address", "datacenter", "token", false)
			require.NoError(t, err)
			err = w.Watch(tt.args.ch, tt.args.chErr, tt.args.ww...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
