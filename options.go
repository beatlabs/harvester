package harvester

import (
	"errors"
	"time"

	"github.com/beatlabs/harvester/config"
	"github.com/beatlabs/harvester/monitor"
	"github.com/beatlabs/harvester/monitor/consul"
	redismon "github.com/beatlabs/harvester/monitor/redis"
	"github.com/beatlabs/harvester/seed"
	seedconsul "github.com/beatlabs/harvester/seed/consul"
	seedredis "github.com/beatlabs/harvester/seed/redis"
	"github.com/go-redis/redis/v8"
)

// TODO: Add some logging!!!

type options struct {
	cfg           *config.Config
	seedParams    []seed.Param
	monitorParams []monitor.Watcher
}

// OptionFunc is used to configure harvester in an optional manner.
type OptionFunc func(opts *options) error

// WithConsulSeedWithPrefix set's up Consul seeder to use prefixes.
func WithConsulSeedWithPrefix(addr, dataCenter, token, folderPrefix string, timeout time.Duration) OptionFunc {
	return func(opts *options) error {
		getter, err := seedconsul.NewWithFolderPrefix(addr, dataCenter, token, folderPrefix, timeout)
		if err != nil {
			return err
		}

		prm, err := seed.NewParam(config.SourceConsul, getter)
		if err != nil {
			return err
		}

		opts.seedParams = append(opts.seedParams, *prm)

		return nil
	}
}

// WithConsulSeed set's up a Consul seeder.
func WithConsulSeed(addr, dataCenter, token string, timeout time.Duration) OptionFunc {
	return WithConsulSeedWithPrefix(addr, dataCenter, token, "", timeout)
}

// WithConsulFolderPrefixMonitor set's up a Consul monitor to use prefixes.
func WithConsulFolderPrefixMonitor(addr, dataCenter, token, folderPrefix string, timeout time.Duration) OptionFunc {
	return func(opts *options) error {
		items := make([]consul.Item, 0)
		for _, field := range opts.cfg.Fields {
			consulKey, ok := field.Sources()[config.SourceConsul]
			if !ok {
				continue
			}
			items = append(items, consul.NewKeyItemWithPrefix(consulKey, folderPrefix))
		}

		prm, err := consul.New(addr, dataCenter, token, timeout, items...)
		if err != nil {
			return err
		}

		opts.monitorParams = append(opts.monitorParams, prm)

		return nil
	}
}

// WithConsulMonitor set's up a Consul monitor.
func WithConsulMonitor(addr, dataCenter, token string, timeout time.Duration) OptionFunc {
	return WithConsulFolderPrefixMonitor(addr, dataCenter, token, "", timeout)
}

// WithConsulSeed set's up a Redis seeder.
func WithRedisSeed(client redis.UniversalClient) OptionFunc {
	return func(opts *options) error {
		getter, err := seedredis.New(client)
		if err != nil {
			return err
		}

		prm, err := seed.NewParam(config.SourceRedis, getter)
		if err != nil {
			return err
		}

		opts.seedParams = append(opts.seedParams, *prm)

		return nil
	}
}

// WithRedisMonitor set's up a Redis monitor.
func WithRedisMonitor(client redis.UniversalClient, pollInterval time.Duration) OptionFunc {
	return func(opts *options) error {
		if pollInterval <= 0 {
			return errors.New("redis monitor poll interval should be a positive number")
		}

		items := make([]string, 0)
		for _, field := range opts.cfg.Fields {
			redisKey, ok := field.Sources()[config.SourceRedis]
			if !ok {
				continue
			}
			items = append(items, redisKey)
		}
		wtc, err := redismon.New(client, pollInterval, items)
		if err != nil {
			return err
		}

		opts.monitorParams = append(opts.monitorParams, wtc)
		return nil
	}
}
