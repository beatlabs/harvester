package harvester

import (
	"context"
	"testing"

	"github.com/beatlabs/harvester/sync"
	"github.com/stretchr/testify/assert"
)

const (
	addr = "127.0.0.1:8500"
)

func TestCreateWithConsul(t *testing.T) {
	type args struct {
		cfg  interface{}
		addr string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "invalid cfg", args: args{cfg: "test", addr: addr}, wantErr: true},
		{name: "invalid address", args: args{cfg: &testConfig{}, addr: ""}, wantErr: true},
		{name: "success", args: args{cfg: &testConfig{}, addr: addr}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.cfg).
				WithConsulSeed(tt.args.addr, "", "", 0).
				WithConsulMonitor(tt.args.addr, "", "", 0).
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

func TestWithNotification(t *testing.T) {
	type args struct {
		cfg      interface{}
		chNotify chan<- string
	}
	tests := map[string]struct {
		args    args
		wantErr bool
	}{
		"nil notify channel": {args: args{cfg: &testConfig{}, chNotify: nil}, wantErr: true},
		"success":            {args: args{cfg: &testConfig{}, chNotify: make(chan string, 0)}, wantErr: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := New(tt.args.cfg).WithNotification(tt.args.chNotify).Create()
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
	assert.Equal(t, int64(8000), cfg.Position.Salary.Get())
	assert.Equal(t, int64(24), cfg.Position.Place.RoomNumber.Get())
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
	Name    sync.String       `seed:"John Doe" consul:"harvester1/name"`
	Age     sync.Int64        `seed:"18"  consul:"harvester/age"`
	Balance sync.Float64      `seed:"99.9"  consul:"harvester/balance"`
	HasJob  sync.Bool         `seed:"true"  consul:"harvester/has-job"`
	FunTime sync.TimeDuration `seed:"1s" consul:"harvester/fun-time"`
}

type testConfigNoConsul struct {
	Name     sync.String       `seed:"John Doe"`
	Age      sync.Int64        `seed:"18"`
	Balance  sync.Float64      `seed:"99.9"`
	HasJob   sync.Bool         `seed:"true"`
	FunTime  sync.TimeDuration `seed:"3s"`
	Position struct {
		Salary sync.Int64 `seed:"8000"`
		Place  struct {
			RoomNumber sync.Int64        `seed:"24"`
			SignTime   sync.TimeDuration `seed:"8s"`
		}
	}
}

type testConfigSeedError struct {
	Name    sync.String       `seed:"John Doe"`
	Age     sync.Int64        `seed:"XXX"`
	Balance sync.Float64      `seed:"99.9"`
	HasJob  sync.Bool         `seed:"true"`
	FunTime sync.TimeDuration `seed:"1s"`
}
