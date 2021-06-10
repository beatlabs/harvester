package monitor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/beatlabs/harvester/change"
	"github.com/beatlabs/harvester/config"
	"github.com/beatlabs/harvester/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cfg, err := config.New(&testConfig{}, nil)
	require.NoError(t, err)
	errCfg, err := config.New(&testConfig{}, nil)
	require.NoError(t, err)
	errCfg.Fields[3].Sources()[config.SourceConsul] = "/config/balance"
	require.NoError(t, err)
	watchers := []Watcher{&testWatcher{}}
	type args struct {
		cfg *config.Config
		ww  []Watcher
	}
	tests := map[string]struct {
		args    args
		wantErr bool
	}{
		"success":        {args: args{cfg: cfg, ww: watchers}, wantErr: false},
		"missing cfg":    {args: args{cfg: nil, ww: watchers}, wantErr: true},
		"empty watchers": {args: args{cfg: cfg, ww: nil}, wantErr: true},
		"error watchers": {args: args{cfg: errCfg, ww: watchers}, wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	cfg, err := config.New(&testConfig{}, nil)
	require.NoError(t, err)
	watchers := []Watcher{&testWatcher{}, &testWatcher{err: true}}
	mon, err := New(cfg, watchers...)
	require.NoError(t, err)
	err = mon.Monitor(context.Background())
	assert.Error(t, err)
}

func TestMonitor_Monitor(t *testing.T) {
	c := &testConfig{}
	cfg, err := config.New(c, nil)
	require.NoError(t, err)
	watchers := []Watcher{&testWatcher{}}
	mon, err := New(cfg, watchers...)
	require.NoError(t, err)
	ctx, cnl := context.WithCancel(context.Background())
	err = mon.Monitor(ctx)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	cnl()
	assert.Equal(t, int64(25), c.Age.Get())
	assert.Equal(t, 111.11, c.Balance.Get())
	assert.Equal(t, false, c.HasJob.Get())
	assert.Equal(t, 6*time.Hour, c.WorkHours.Get())
	assert.Equal(t, 7*time.Hour, c.NonWorkHours.Get())
}

type testConfig struct {
	Name         sync.String       `seed:"John Doe" env:"ENV_NAME"`
	Age          sync.Int64        `env:"ENV_AGE" consul:"/config/age"`
	Balance      sync.Float64      `seed:"99.9" env:"ENV_BALANCE" consul:"/config/balance"`
	HasJob       sync.Bool         `seed:"true" env:"ENV_HAS_JOB" consul:"/config/has-job"`
	WorkHours    sync.TimeDuration `seed:"5h" env:"ENV_WORK_HOURS" consul:"/config/work_hours"`
	NonWorkHours sync.TimeDuration `seed:"5h" env:"ENV_NON_WORK_HOURS" redis:"/config/non_work_hours"`
}

type testWatcher struct {
	err bool
}

func (tw *testWatcher) Watch(_ context.Context, ch chan<- []*change.Change) error {
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
		change.New(config.SourceConsul, "/config/has-job", "XXX", 2),
		change.New(config.SourceConsul, "/config/work_hours", (6 * time.Hour).String(), 2),
		change.New(config.SourceRedis, "/config/non_work_hours", (7 * time.Hour).String(), 2),
	}
	return nil
}
