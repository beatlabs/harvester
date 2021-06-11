package harvester

import (
	"context"
	"errors"
	"time"

	"github.com/beatlabs/harvester/config"
	"github.com/beatlabs/harvester/log"
	"github.com/beatlabs/harvester/monitor"
	"github.com/beatlabs/harvester/monitor/consul"
	redismon "github.com/beatlabs/harvester/monitor/redis"
	"github.com/beatlabs/harvester/seed"
	seedconsul "github.com/beatlabs/harvester/seed/consul"
	seedredis "github.com/beatlabs/harvester/seed/redis"
	"github.com/go-redis/redis/v8"
)

// Seeder interface for seeding initial values of the configuration.
type Seeder interface {
	Seed(cfg *config.Config) error
}

// Monitor defines a interface for monitoring configuration changes from various sources.
type Monitor interface {
	Monitor(ctx context.Context) error
}

// Harvester interface.
type Harvester interface {
	Harvest(ctx context.Context) error
}

type harvester struct {
	cfg     *config.Config
	seeder  Seeder
	monitor Monitor
}

// Harvest take the configuration object, initializes it and monitors for changes.
func (h *harvester) Harvest(ctx context.Context) error {
	err := h.seeder.Seed(h.cfg)
	if err != nil {
		return err
	}
	if h.monitor == nil {
		return nil
	}
	return h.monitor.Monitor(ctx)
}

type consulConfig struct {
	addr, dataCenter, token string
	timeout                 time.Duration
}

// Builder of a harvester instance.
type Builder struct {
	cfg                      interface{}
	seedConsulCfg            *consulConfig
	monitorConsulCfg         *consulConfig
	err                      error
	chNotify                 chan<- config.ChangeNotification
	monitorRedisClient       redis.UniversalClient
	seedRedisClient          redis.UniversalClient
	monitorRedisPollInterval time.Duration
}

// New constructor.
func New(cfg interface{}) *Builder {
	return &Builder{cfg: cfg}
}

// WithNotification constructor.
func (b *Builder) WithNotification(chNotify chan<- config.ChangeNotification) *Builder {
	if b.err != nil {
		return b
	}
	if chNotify == nil {
		b.err = errors.New("notification channel is nil")
		return b
	}
	b.chNotify = chNotify
	return b
}

// WithConsulSeed enables support for seeding values with consul.
func (b *Builder) WithConsulSeed(addr, dataCenter, token string, timeout time.Duration) *Builder {
	if b.err != nil {
		return b
	}
	b.seedConsulCfg = &consulConfig{
		addr:       addr,
		dataCenter: dataCenter,
		token:      token,
		timeout:    timeout,
	}
	return b
}

// WithConsulMonitor enables support for monitoring key/prefixes on ConsulLogger. It automatically parses the config
// and monitors every field found tagged with ConsulLogger.
func (b *Builder) WithConsulMonitor(addr, dataCenter, token string, timeout time.Duration) *Builder {
	if b.err != nil {
		return b
	}
	b.monitorConsulCfg = &consulConfig{
		addr:       addr,
		dataCenter: dataCenter,
		token:      token,
		timeout:    timeout,
	}
	return b
}

// WithRedisSeed enables support for seeding values with redis.
func (b *Builder) WithRedisSeed(client redis.UniversalClient) *Builder {
	if b.err != nil {
		return b
	}
	if client == nil {
		b.err = errors.New("redis seed client is nil")
		return b
	}
	b.seedRedisClient = client
	return b
}

// WithRedisMonitor enables support for monitoring keys in Redis. It automatically parses the config
// and monitors every field found tagged with ConsulLogger.
func (b *Builder) WithRedisMonitor(client redis.UniversalClient, pollInterval time.Duration) *Builder {
	if b.err != nil {
		return b
	}
	if client == nil {
		b.err = errors.New("redis monitor client is nil")
		return b
	}
	if pollInterval <= 0 {
		b.err = errors.New("redis monitor poll interval should be a positive number")
		return b
	}
	b.monitorRedisClient = client
	b.monitorRedisPollInterval = pollInterval
	return b
}

// Create the harvester instance.
func (b *Builder) Create() (Harvester, error) {
	if b.err != nil {
		return nil, b.err
	}

	cfg, err := config.New(b.cfg, b.chNotify)
	if err != nil {
		return nil, err
	}

	sd, err := b.setupSeeding()
	if err != nil {
		return nil, err
	}
	mon, err := b.setupMonitoring(cfg)
	if err != nil {
		return nil, err
	}

	return &harvester{seeder: sd, monitor: mon, cfg: cfg}, nil
}

func (b *Builder) setupSeeding() (Seeder, error) {
	pp := make([]seed.Param, 0)

	consulSeedParam, err := b.setupConsulSeeding()
	if err != nil {
		return nil, err
	}
	if consulSeedParam != nil {
		pp = append(pp, *consulSeedParam)
	}

	redisSeedParam, err := b.setupRedisSeeding()
	if err != nil {
		return nil, err
	}
	if redisSeedParam != nil {
		pp = append(pp, *redisSeedParam)
	}

	return seed.New(pp...), nil
}

func (b *Builder) setupConsulSeeding() (*seed.Param, error) {
	if b.seedConsulCfg == nil {
		return nil, nil
	}

	getter, err := seedconsul.New(b.seedConsulCfg.addr, b.seedConsulCfg.dataCenter, b.seedConsulCfg.token,
		b.seedConsulCfg.timeout)
	if err != nil {
		return nil, err
	}

	return seed.NewParam(config.SourceConsul, getter)
}

func (b *Builder) setupRedisSeeding() (*seed.Param, error) {
	if b.seedRedisClient == nil {
		return nil, nil
	}

	getter, err := seedredis.New(b.seedRedisClient)
	if err != nil {
		return nil, err
	}

	return seed.NewParam(config.SourceRedis, getter)
}

func (b *Builder) setupMonitoring(cfg *config.Config) (Monitor, error) {
	var watchers []monitor.Watcher

	consulWatcher, err := b.setupConsulMonitoring(cfg)
	if err != nil {
		return nil, err
	}
	if consulWatcher != nil {
		watchers = append(watchers, consulWatcher)
	}

	redisWatcher, err := b.setupRedisMonitoring(cfg)
	if err != nil {
		return nil, err
	}
	if redisWatcher != nil {
		watchers = append(watchers, redisWatcher)
	}

	if len(watchers) == 0 {
		return nil, nil
	}

	return monitor.New(cfg, watchers...)
}

func (b *Builder) setupConsulMonitoring(cfg *config.Config) (*consul.Watcher, error) {
	if b.monitorConsulCfg == nil {
		return nil, nil
	}
	items := make([]consul.Item, 0)
	for _, field := range cfg.Fields {
		consulKey, ok := field.Sources()[config.SourceConsul]
		if !ok {
			continue
		}
		log.Infof(`automatically monitoring consul key "%s"`, consulKey)
		items = append(items, consul.NewKeyItem(consulKey))
	}
	return consul.New(b.monitorConsulCfg.addr, b.monitorConsulCfg.dataCenter, b.monitorConsulCfg.token,
		b.monitorConsulCfg.timeout, items...)
}

func (b *Builder) setupRedisMonitoring(cfg *config.Config) (*redismon.Watcher, error) {
	if b.monitorRedisClient == nil {
		return nil, nil
	}
	items := make([]string, 0)
	for _, field := range cfg.Fields {
		redisKey, ok := field.Sources()[config.SourceRedis]
		if !ok {
			continue
		}
		log.Infof(`automatically monitoring redis key "%s"`, redisKey)
		items = append(items, redisKey)
	}
	return redismon.New(b.monitorRedisClient, b.monitorRedisPollInterval, items)
}
