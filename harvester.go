package harvester

import (
	"context"
	"errors"
	"time"

	"github.com/beatlabs/harvester/config"
	"github.com/beatlabs/harvester/log"
	"github.com/beatlabs/harvester/monitor"
	"github.com/beatlabs/harvester/monitor/consul"
	"github.com/beatlabs/harvester/seed"
	seedConsul "github.com/beatlabs/harvester/seed/consul"
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
	cfg              interface{}
	seedConsulCfg    *consulConfig
	monitorConsulCfg *consulConfig
	err              error
	chNotify         chan<- string
}

// New constructor.
func New(cfg interface{}) *Builder {
	return &Builder{cfg: cfg}
}

// WithNotification constructor.
func (b *Builder) WithNotification(chNotify chan<- string) *Builder {
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
	if b.seedConsulCfg != nil {

		getter, err := seedConsul.New(b.seedConsulCfg.addr, b.seedConsulCfg.dataCenter, b.seedConsulCfg.token, b.seedConsulCfg.timeout)
		if err != nil {
			return nil, err
		}

		p, err := seed.NewParam(config.SourceConsul, getter)
		if err != nil {
			return nil, err
		}
		pp = append(pp, *p)
	}

	return seed.New(pp...), nil
}

func (b *Builder) setupMonitoring(cfg *config.Config) (Monitor, error) {
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
	wtc, err := consul.New(b.monitorConsulCfg.addr, b.monitorConsulCfg.dataCenter, b.monitorConsulCfg.token, b.monitorConsulCfg.timeout, items...)
	if err != nil {
		return nil, err
	}

	mon, err := monitor.New(cfg, wtc)
	if err != nil {
		return nil, err
	}
	return mon, nil
}
