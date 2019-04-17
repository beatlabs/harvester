package monitor

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taxibeat/harvester/change"
)

func TestNewMonitor(t *testing.T) {
	expectedAgeField := field{
		Name:      "Age",
		Kind:      reflect.Int64,
		EnvVarKey: "ENV_AGE",
		ConsulKey: "/config/age",
	}
	expectedBalanceField := field{
		Name:      "Balance",
		Kind:      reflect.Float64,
		SeedValue: "99.9",
		EnvVarKey: "ENV_BALANCE",
		ConsulKey: "/config/balance",
	}
	expectedHasJobField := field{
		Name:      "HasJob",
		Kind:      reflect.Bool,
		SeedValue: "true",
		EnvVarKey: "ENV_HAS_JOB",
		ConsulKey: "/config/has-job",
	}
	require.NoError(t, os.Setenv("ENV_AGE", "18"))
	require.NoError(t, os.Setenv("ENV_XXX", "aaa"))
	ch := make(chan []*change.Change)
	type args struct {
		cfg       interface{}
		ch        <-chan []*change.Change
		consulGet GetValueFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "config nil", args: args{cfg: nil, ch: ch, consulGet: nil}, wantErr: true},
		{name: "channel nil", args: args{cfg: &testConfig{}, ch: nil, consulGet: nil}, wantErr: true},
		{name: "channel nil", args: args{cfg: &testConfig{}, ch: ch, consulGet: nil}, wantErr: true},
		{name: "config not pointer", args: args{cfg: testConfig{}, ch: ch, consulGet: stubGetFunc}, wantErr: true},
		{name: "not supported data types", args: args{cfg: &testInvalidConfig{}, ch: ch, consulGet: stubGetFunc}, wantErr: true},
		{name: "duplicate consul key", args: args{cfg: &testDuplicateConfig{}, ch: ch, consulGet: stubGetFunc}, wantErr: true},
		{name: "invalid bool seed", args: args{cfg: &testInvalidIntSeedConfig{}, ch: ch, consulGet: stubGetFunc}, wantErr: true},
		{name: "invalid env var val", args: args{cfg: &testInvalidEnvVarValueConfig{}, ch: ch, consulGet: stubGetFunc}, wantErr: true},
		{name: "consul get error", args: args{cfg: &testConfig{}, ch: ch, consulGet: stubGetErrorFunc}, wantErr: true},
		{name: "invalid consul get val", args: args{cfg: &testInvalidConsulValueConfig{}, ch: ch, consulGet: stubGetFunc}, wantErr: true},
		{name: "success", args: args{cfg: &testConfig{}, ch: ch, consulGet: stubGetFunc}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.cfg, tt.args.ch, tt.args.consulGet)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				cfg := tt.args.cfg.(*testConfig)
				assert.Equal(t, "John Doe", cfg.Name)
				assert.Equal(t, int64(25), cfg.Age)
				assert.Equal(t, 99.9, cfg.Balance)
				assert.True(t, cfg.HasJob)
				assert.Contains(t, got.monitorMap, change.SourceConsul)
				assert.Len(t, got.monitorMap[change.SourceConsul], 3)
				assert.Contains(t, got.monitorMap[change.SourceConsul], "/config/age")
				assert.Equal(t, expectedAgeField, *got.monitorMap[change.SourceConsul]["/config/age"])
				assert.Contains(t, got.monitorMap[change.SourceConsul], "/config/balance")
				assert.Equal(t, expectedBalanceField, *got.monitorMap[change.SourceConsul]["/config/balance"])
				assert.Contains(t, got.monitorMap[change.SourceConsul], "/config/has-job")
				assert.Equal(t, expectedHasJobField, *got.monitorMap[change.SourceConsul]["/config/has-job"])
			}
		})
	}
}

func TestMonitor_Monitor(t *testing.T) {
	require.NoError(t, os.Setenv("ENV_AGE", "18"))
	chDone := make(chan struct{})
	ch := make(chan []*change.Change)
	cfg := &testConfig{}
	mon, err := New(cfg, ch, stubGetFunc)
	require.NoError(t, err)
	require.Equal(t, "John Doe", cfg.Name)
	require.Equal(t, int64(25), cfg.Age)
	require.Equal(t, 99.9, cfg.Balance)
	require.True(t, cfg.HasJob)
	ctx, cnl := context.WithCancel(context.Background())
	go func() {
		mon.Monitor(ctx)
		chDone <- struct{}{}
	}()
	t.Run("change age", func(t *testing.T) {
		ch <- []*change.Change{&change.Change{
			Src:     change.SourceConsul,
			Key:     "/config/age",
			Value:   "23",
			Version: 1,
		}}
		time.Sleep(10 * time.Millisecond)
		mon.Lock()
		defer mon.Unlock()
		require.Equal(t, int64(23), cfg.Age)
	})
	t.Run("age does not change due to version check", func(t *testing.T) {
		ch <- []*change.Change{&change.Change{
			Src:     change.SourceConsul,
			Key:     "/config/age",
			Value:   "99",
			Version: 0,
		}}
		time.Sleep(10 * time.Millisecond)
		mon.Lock()
		defer mon.Unlock()
		require.Equal(t, int64(23), cfg.Age)
	})
	t.Run("balance change", func(t *testing.T) {
		ch <- []*change.Change{&change.Change{
			Src:     change.SourceConsul,
			Key:     "/config/balance",
			Value:   "123.4",
			Version: 1,
		}}
		time.Sleep(10 * time.Millisecond)
		mon.Lock()
		defer mon.Unlock()
		require.Equal(t, 123.4, cfg.Balance)
	})
	t.Run("has job(bool) change", func(t *testing.T) {
		ch <- []*change.Change{&change.Change{
			Src:     change.SourceConsul,
			Key:     "/config/has-job",
			Value:   "false",
			Version: 1,
		}}
		time.Sleep(10 * time.Millisecond)
		mon.Lock()
		defer mon.Unlock()
		require.False(t, cfg.HasJob)
	})
	t.Run("invalid source, no change", func(t *testing.T) {
		ch <- []*change.Change{&change.Change{
			Src:     change.Source("XXX"),
			Key:     "/config/has-job",
			Value:   "true",
			Version: 2,
		}}
		time.Sleep(10 * time.Millisecond)
		mon.Lock()
		defer mon.Unlock()
		require.False(t, cfg.HasJob)
	})
	t.Run("key not found, no change", func(t *testing.T) {
		ch <- []*change.Change{&change.Change{
			Src:     change.SourceConsul,
			Key:     "/config/has-job1",
			Value:   "true",
			Version: 2,
		}}
		time.Sleep(10 * time.Millisecond)
		mon.Lock()
		defer mon.Unlock()
		require.False(t, cfg.HasJob)
	})
	t.Run("invalid bool, no change", func(t *testing.T) {
		ch <- []*change.Change{&change.Change{
			Src:     change.SourceConsul,
			Key:     "/config/has-job",
			Value:   "XXX",
			Version: 2,
		}}
		time.Sleep(10 * time.Millisecond)
		mon.Lock()
		defer mon.Unlock()
		require.False(t, cfg.HasJob)
	})
	t.Run("invalid int, no change", func(t *testing.T) {
		ch <- []*change.Change{&change.Change{
			Src:     change.SourceConsul,
			Key:     "/config/age",
			Value:   "XXX",
			Version: 4,
		}}
		time.Sleep(10 * time.Millisecond)
		mon.Lock()
		defer mon.Unlock()
		require.Equal(t, int64(23), cfg.Age)
	})
	t.Run("invalid float, no change", func(t *testing.T) {
		ch <- []*change.Change{&change.Change{
			Src:     change.SourceConsul,
			Key:     "/config/balance",
			Value:   "XXX",
			Version: 5,
		}}
		time.Sleep(10 * time.Millisecond)
		mon.Lock()
		defer mon.Unlock()
		require.Equal(t, 123.4, cfg.Balance)
	})
	cnl()
	<-chDone
}

var stubGetFunc = func(key string) (string, error) {
	switch key {
	case "/config/age":
		return "25", nil
	case "/config/balance":
		return "999.99", nil
	case "/config/has-job":
		return "false", nil
	case "/config/xxx":
		return "xxx", nil
	}
	return "", errors.New("should not happen")
}

var stubGetErrorFunc = func(_ string) (string, error) {
	return "", errors.New("should not happen")
}

type testConfig struct {
	Name    string  `seed:"John Doe" env:"ENV_NAME"`
	Age     int64   `env:"ENV_AGE" consul:"/config/age"`
	Balance float64 `seed:"99.9" env:"ENV_BALANCE" consul:"/config/balance"`
	HasJob  bool    `seed:"true" env:"ENV_HAS_JOB" consul:"/config/has-job"`
}

type testInvalidConfig struct {
	Name    string  `seed:"John Doe" env:"ENV_NAME" consul:"/config/name"`
	Age     int     `seed:"18" env:"ENV_AGE" consul:"/config/age"`
	Balance float32 `seed:"99.9" env:"ENV_BALANCE" consul:"/config/balance"`
	HasJob  bool    `seed:"true" env:"ENV_HAS_JOB" consul:"/config/has-job"`
}

type testDuplicateConfig struct {
	Name string `seed:"John Doe" env:"ENV_NAME"`
	Age1 int64  `env:"ENV_AGE" consul:"/config/age"`
	Age2 int64  `env:"ENV_AGE" consul:"/config/age"`
}

type testInvalidIntSeedConfig struct {
	Age int64 `seed:"XXX" consul:"/config/age"`
}

type testInvalidEnvVarValueConfig struct {
	Age int64 `env:"ENV_XXX" consul:"/config/age"`
}

type testInvalidConsulValueConfig struct {
	Age int64 `consul:"/config/xxx"`
}
