package harvester

import (
	"context"
	"testing"

	"github.com/beatlabs/harvester/monitor/consul"
	"github.com/beatlabs/harvester/sync"
	"github.com/stretchr/testify/assert"
)

const (
	addr = "127.0.0.1:8500"
)

func TestCreateWithConsul(t *testing.T) {
	ii := []consul.Item{consul.NewKeyItem("harvester1/name"), consul.NewPrefixItem("harvester")}
	type args struct {
		cfg   interface{}
		addr  string
		items []consul.Item
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "invalid cfg", args: args{cfg: "test", addr: addr, items: ii}, wantErr: true},
		{name: "invalid address", args: args{cfg: &testConfig{}, addr: "", items: ii}, wantErr: true},
		{name: "missing items", args: args{cfg: &testConfig{}, addr: addr, items: []consul.Item{}}, wantErr: true},
		{name: "success", args: args{cfg: &testConfig{}, addr: addr, items: ii}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.cfg).
				WithConsulSeed(tt.args.addr, "", "", 0).
				WithConsulMonitor(tt.args.addr, "", "", 0, tt.args.items...).
				Create()
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

func TestCreate_NoConsul(t *testing.T) {
	cfg := &testConfigNoConsul{}
	got, err := New(cfg).Create()
	assert.NoError(t, err)
	assert.NotNil(t, got)
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()
	err = got.Harvest(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "John Doe", cfg.Name.Get())
	assert.Equal(t, int64(18), cfg.Age.Get())
	assert.Equal(t, 99.9, cfg.Balance.Get())
	assert.Equal(t, true, cfg.HasJob.Get())
}

func TestCreate_SeedError(t *testing.T) {
	cfg := &testConfigSeedError{}
	got, err := New(cfg).Create()
	assert.NoError(t, err)
	assert.NotNil(t, got)
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()
	err = got.Harvest(ctx)
	assert.Error(t, err)
}

type testConfig struct {
	Name    sync.String  `seed:"John Doe" consul:"harvester1/name"`
	Age     sync.Int64   `seed:"18"  consul:"harvester/age"`
	Balance sync.Float64 `seed:"99.9"  consul:"harvester/balance"`
	HasJob  sync.Bool    `seed:"true"  consul:"harvester/has-job"`
}

type testConfigNoConsul struct {
	Name    sync.String  `seed:"John Doe"`
	Age     sync.Int64   `seed:"18"`
	Balance sync.Float64 `seed:"99.9"`
	HasJob  sync.Bool    `seed:"true"`
}

type testConfigSeedError struct {
	Name    sync.String  `seed:"John Doe"`
	Age     sync.Int64   `seed:"XXX"`
	Balance sync.Float64 `seed:"99.9"`
	HasJob  sync.Bool    `seed:"true"`
}
