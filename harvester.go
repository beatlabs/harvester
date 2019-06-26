package harvester

import (
	"context"
	"time"

	"github.com/beatlabs/harvester/config"
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

// Builder of a harvester instance.
type Builder struct {
	cfg        *config.Config
	watchers   []monitor.Watcher
	seedParams []seed.Param
	err        error
}

// New constructor.
func New(cfg interface{}) *Builder {
	b := &Builder{}
	c, err := config.New(cfg)
	if err != nil {
		b.err = err
		return b
	}
	b.cfg = c
	b.seedParams = []seed.Param{}
	return b
}

// WithConsulSeed enables support for seeding values with consul.
func (b *Builder) WithConsulSeed(addr, dataCenter, token string, timeout time.Duration) *Builder {
	if b.err != nil {
		return b
	}
	getter, err := seedConsul.New(addr, dataCenter, token, timeout)
	if err != nil {
		b.err = err
		return b
	}
	p, err := seed.NewParam(config.SourceConsul, getter)
	if err != nil {
		b.err = err
		return b
	}
	b.seedParams = append(b.seedParams, *p)
	return b
}

// WithConsulMonitor enables support for monitoring key/prefixes on consul.
func (b *Builder) WithConsulMonitor(addr, dc, token string, timeout time.Duration, ii ...consul.Item) *Builder {
	if b.err != nil {
		return b
	}
	wtc, err := consul.New(addr, dc, token, timeout, ii...)
	if err != nil {
		b.err = err
		return b
	}
	b.watchers = append(b.watchers, wtc)
	return b
}

// Create the harvester instance.
func (b *Builder) Create() (Harvester, error) {
	if b.err != nil {
		return nil, b.err
	}
	sd := seed.New(b.seedParams...)

	var mon Monitor
	if len(b.watchers) == 0 {
		return &harvester{seeder: sd, cfg: b.cfg}, nil
	}
	mon, err := monitor.New(b.cfg, b.watchers...)
	if err != nil {
		return nil, err
	}
	return &harvester{seeder: sd, monitor: mon, cfg: b.cfg}, nil
}
