package consul

import (
	"context"
	"testing"
	"time"

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
		ctx    context.Context
		cancel context.CancelFunc
	}
	withCancel := func() args {
		ctx, cnl := context.WithCancel(context.Background())
		return args{ctx: ctx, cancel: cnl}
	}
	tests := map[string]struct {
		args    args
		items   []Item
		wantErr bool
	}{
		"invalid item (missing type)": {
			args:    withCancel(),
			items:   []Item{{}},
			wantErr: true,
		},
		"missing context": {
			args: args{},
			items: []Item{
				{
					tp:     "key",
					key:    "test",
					prefix: "",
				},
			},
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			w, err := New("xxx", "", "", 0, tt.items...)
			require.NoError(t, err)
			ch, err := w.Watch(tt.args.ctx)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, ch)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
