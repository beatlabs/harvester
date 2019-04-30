package monitor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taxibeat/harvester/change"
	"github.com/taxibeat/harvester/config"
)

func TestNew(t *testing.T) {
	cfg, err := config.New(&testConfig{})
	require.NoError(t, err)
	errCfg, err := config.New(&testConfig{})
	errCfg.Fields[3].Sources[config.SourceConsul] = "/config/balance"
	require.NoError(t, err)
	watchers := []Watcher{&testWatcher{}}
	type args struct {
		cfg *config.Config
		ww  []Watcher
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "success", args: args{cfg: cfg, ww: watchers}, wantErr: false},
		{name: "missing cfg", args: args{cfg: nil, ww: watchers}, wantErr: true},
		{name: "empty watchers", args: args{cfg: cfg, ww: nil}, wantErr: true},
		{name: "error watchers", args: args{cfg: errCfg, ww: watchers}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.cfg, tt.args.ww...)
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

func TestMonitor_Monitor_Error(t *testing.T) {
	cfg, err := config.New(&testConfig{})
	require.NoError(t, err)
	watchers := []Watcher{&testWatcher{}, &testWatcher{err: true}}
	mon, err := New(cfg, watchers...)
	require.NoError(t, err)
	chErr := make(chan error)
	err = mon.Monitor(context.Background(), chErr)
	assert.Error(t, err)
}

func TestMonitor_Monitor(t *testing.T) {
	c := &testConfig{}
	cfg, err := config.New(c)
	require.NoError(t, err)
	watchers := []Watcher{&testWatcher{}}
	mon, err := New(cfg, watchers...)
	require.NoError(t, err)
	chErr := make(chan error)
	ctx, cnl := context.WithCancel(context.Background())
	err = mon.Monitor(ctx, chErr)
	assert.NoError(t, err)
	time.Sleep(1000 * time.Millisecond)
	cnl()
	assert.Equal(t, int64(25), c.Age)
	assert.Equal(t, 111.11, c.Balance)
	assert.Equal(t, false, c.HasJob)
}

type testConfig struct {
	Name    string  `seed:"John Doe" env:"ENV_NAME"`
	Age     int64   `env:"ENV_AGE" consul:"/config/age"`
	Balance float64 `seed:"99.9" env:"ENV_BALANCE" consul:"/config/balance"`
	HasJob  bool    `seed:"true" env:"ENV_HAS_JOB" consul:"/config/has-job"`
}

type testWatcher struct {
	err bool
}

func (tw *testWatcher) Watch(ctx context.Context, ch chan<- []*change.Change, chErr chan<- error) error {
	if tw.err {
		return errors.New("TEST")
	}
	ch <- []*change.Change{
		change.New(config.SourceConsul, "/config/age", "25", 1),
		change.New(config.SourceConsul, "/config/balance", "111.11", 1),
		change.New(config.SourceConsul, "/config/has-job", "false", 1),
		change.New(config.SourceEnv, "/config/has-job", "false", 1),
		change.New(config.SourceConsul, "/config/has-job1", "false", 1),
		change.New(config.SourceConsul, "/config/has-job", "false", 0),
	}
	return nil
}
