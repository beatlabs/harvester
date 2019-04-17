package consul

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taxibeat/harvester/watcher"
)

func TestNewConfig(t *testing.T) {
	ch := make(chan []*watcher.Change)
	chErr := make(chan error)
	type args struct {
		address string
		ch      chan []*watcher.Change
		chErr   chan error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "success", args: args{address: "addr", ch: ch, chErr: chErr}, wantErr: false},
		{name: "missing address", args: args{address: "", ch: ch, chErr: chErr}, wantErr: true},
		{name: "missing channel", args: args{address: "addr", ch: nil, chErr: chErr}, wantErr: true},
		{name: "missing error channel", args: args{address: "addr", ch: ch, chErr: nil}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfig(tt.args.address, "datacenter", "token", tt.args.ch, tt.args.chErr)
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

func TestNew(t *testing.T) {
	type args struct {
		cfg *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "success", args: args{cfg: &Config{}}, wantErr: false},
		{name: "missing config", args: args{cfg: nil}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.cfg)
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
	type args struct {
		ww []watcher.Item
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "missing watch items", args: args{ww: nil}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := New(&Config{})
			require.NoError(t, err)
			err = w.Watch(tt.args.ww...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
