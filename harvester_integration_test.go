//go:build integration
// +build integration

package harvester

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/beatlabs/harvester/sync"
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	csl         *api.KV
	redisClient = redis.NewClient(&redis.Options{})
)

const (
	redisWorkTimeKey   = "work-time"
	redisWorkTimeValue = "2s"
	redisDooKey        = "doo"
	redisDooValue      = "999"
)

type testConfigWithSecret struct {
	Name     sync.Secret       `seed:"John Doe" consul:"harvester1/name"`
	Age      sync.Int64        `seed:"18" consul:"harvester/age"`
	Balance  sync.Float64      `seed:"99.9" consul:"harvester/balance"`
	HasJob   sync.Bool         `seed:"true" consul:"harvester/has-job"`
	FunTime  sync.TimeDuration `seed:"1s" consul:"harvester/fun-time"`
	WorkTime sync.TimeDuration `seed:"1s" redis:"work-time"`
	Foo      fooStruct
}

type fooStruct struct {
	Bar sync.Int64 `seed:"123" consul:"harvester/foo/bar"`
	Doo sync.Int64 `seed:"234" redis:"doo"`
}

func TestMain(m *testing.M) {
	config := api.DefaultConfig()
	config.Address = addr
	c, err := api.NewClient(config)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	csl = c.KV()

	err = cleanup()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	err = setup()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	ret := m.Run()
	err = cleanup()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	os.Exit(ret)
}

func Test_harvester_Harvest(t *testing.T) {
	cfg := testConfigWithSecret{}
	h, err := New(&cfg, nil,
		WithConsulSeed(addr, "", "", 0),
		WithConsulMonitor(addr, "", "", 0),
		WithRedisSeed(redisClient),
		WithRedisMonitor(redisClient, 10*time.Millisecond))
	require.NoError(t, err)

	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()
	err = h.Harvest(ctx)

	require.NoError(t, err)
	assert.Equal(t, "Mr. Smith", cfg.Name.Get())
	assert.Equal(t, int64(99), cfg.Age.Get())
	assert.InDelta(t, 111.1, cfg.Balance.Get(), 0.01)
	assert.False(t, cfg.HasJob.Get())
	assert.Equal(t, 1*time.Second, cfg.FunTime.Get())
	assert.Equal(t, 2*time.Second, cfg.WorkTime.Get())
	assert.Equal(t, int64(123), cfg.Foo.Bar.Get())
	assert.Equal(t, int64(999), cfg.Foo.Doo.Get())

	_, err = csl.Put(&api.KVPair{Key: "harvester1/name", Value: []byte("Mr. Anderson")}, nil)
	require.NoError(t, err)
	time.Sleep(1000 * time.Millisecond)
	assert.Equal(t, "Mr. Anderson", cfg.Name.Get())

	duration, err := time.ParseDuration("5s")
	require.NoError(t, err)
	_, err = csl.Put(&api.KVPair{Key: "harvester/fun-time", Value: []byte(duration.String())}, nil)
	require.NoError(t, err)
	time.Sleep(1000 * time.Millisecond)
	assert.Equal(t, 5*time.Second, cfg.FunTime.Get())

	_, err = csl.Put(&api.KVPair{Key: "harvester/foo/bar", Value: []byte("42")}, nil)
	require.NoError(t, err)
	time.Sleep(1000 * time.Millisecond)
	assert.Equal(t, int64(42), cfg.Foo.Bar.Get())

	res, err := redisClient.Set(context.Background(), redisWorkTimeKey, "3s", 0).Result()
	require.NoError(t, err)
	require.Equal(t, "OK", res)
	time.Sleep(1000 * time.Millisecond)
	assert.Equal(t, 3*time.Second, cfg.WorkTime.Get())
}

func cleanup() error {
	_, err := csl.Delete("harvester1/name", nil)
	if err != nil {
		return err
	}
	_, err = csl.DeleteTree("harvester", nil)
	if err != nil {
		return err
	}
	_, err = redisClient.Del(context.Background(), redisWorkTimeKey, redisDooKey).Result()
	if err != nil {
		return err
	}
	return nil
}

func setup() error {
	_, err := csl.Put(&api.KVPair{Key: "harvester1/name", Value: []byte("Mr. Smith")}, nil)
	if err != nil {
		return err
	}
	_, err = csl.Put(&api.KVPair{Key: "harvester/age", Value: []byte("99")}, nil)
	if err != nil {
		return err
	}
	_, err = csl.Put(&api.KVPair{Key: "harvester/balance", Value: []byte("111.1")}, nil)
	if err != nil {
		return err
	}
	_, err = csl.Put(&api.KVPair{Key: "harvester/has-job", Value: []byte("false")}, nil)
	if err != nil {
		return err
	}
	_, err = redisClient.Set(context.Background(), redisWorkTimeKey, redisWorkTimeValue, 0).Result()
	if err != nil {
		return err
	}
	_, err = redisClient.Set(context.Background(), redisDooKey, redisDooValue, 0).Result()
	if err != nil {
		return err
	}
	return nil
}
