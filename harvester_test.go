package harvester

import (
	"context"
	"testing"
	"time"

	"github.com/beatlabs/harvester/config"
	"github.com/beatlabs/harvester/sync"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

const (
	addr = "127.0.0.1:8500"
)

func TestCreateWithConsulAndRedis(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{})
	type args struct {
		cfg                interface{}
		consulAddress      string
		seedRedisClient    *redis.Client
		monitorRedisClient *redis.Client
	}
	tests := map[string]struct {
		args        args
		expectedErr string
	}{
		"invalid config": {
			args: args{
				cfg:                "test",
				consulAddress:      addr,
				seedRedisClient:    redisClient,
				monitorRedisClient: redisClient,
			}, expectedErr: "configuration should be a pointer type",
		},
		"invalid consul address": {
			args: args{
				cfg:                &testConfig{},
				consulAddress:      "",
				seedRedisClient:    redisClient,
				monitorRedisClient: redisClient,
			}, expectedErr: "address is empty",
		},
		"invalid redis seed client": {
			args: args{
				cfg:                &testConfig{},
				consulAddress:      addr,
				seedRedisClient:    nil,
				monitorRedisClient: redisClient,
			}, expectedErr: "redis seed client is nil",
		},
		"invalid redis monitor client": {
			args: args{
				cfg:                &testConfig{},
				consulAddress:      addr,
				seedRedisClient:    redisClient,
				monitorRedisClient: nil,
			}, expectedErr: "redis monitor client is nil",
		},
		"success": {
			args: args{
				cfg:                &testConfig{},
				consulAddress:      addr,
				seedRedisClient:    redisClient,
				monitorRedisClient: redisClient,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := New(tt.args.cfg).
				WithConsulSeed(tt.args.consulAddress, "", "", 0).
				WithConsulMonitor(tt.args.consulAddress, "", "", 0).
				WithRedisSeed(tt.args.seedRedisClient).
				WithRedisMonitor(tt.args.monitorRedisClient, 10*time.Millisecond).
				Create()
			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
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
		chNotify chan<- config.ChangeNotification
	}
	tests := map[string]struct {
		args    args
		wantErr bool
	}{
		"nil notify channel": {args: args{cfg: &testConfig{}, chNotify: nil}, wantErr: true},
		"success":            {args: args{cfg: &testConfig{}, chNotify: make(chan config.ChangeNotification)}, wantErr: false},
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

func TestCreate_NoConsulOrRedis(t *testing.T) {
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
	IsAdult sync.Bool         `seed:"false" redis:"is-adult"`
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
