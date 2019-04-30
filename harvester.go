package harvester

import (
	"context"
	"errors"

	"github.com/taxibeat/harvester/config"
	"github.com/taxibeat/harvester/monitor"
	"github.com/taxibeat/harvester/monitor/consul"
	"github.com/taxibeat/harvester/seed"
)

// Seeder interface for seeding initial values of the configuration.
type Seeder interface {
	Seed(cfg *config.Config) error
}

// Monitor defines a interface for monitoring configuration changes from various sources.
type Monitor interface {
	Monitor(ctx context.Context, chErr chan<- error) error
}

// Harvester interface.
type Harvester interface {
	Harvest(ctx context.Context, cfg interface{}) error
}

type harvester struct {
	seeder  Seeder
	monitor Monitor
	chErr   chan<- error
}

// Harvest take the configuration object, initializes it and monitors for changes.
func (h *harvester) Harvest(ctx context.Context, cfg interface{}) error {
	c, err := config.New(cfg)
	if err != nil {
		return err
	}
	err = h.seeder.Seed(c)
	if err != nil {
		return err
	}
	return h.monitor.Monitor(ctx, h.chErr)
}

// Builder of a harvester instance.
type Builder struct {
	cfg         *config.Config
	consulGet   seed.GetValueFunc
	consulItems []consul.Item
	watchers    map[config.Source]monitor.Watcher
	consulAddr  string
	consulDC    string
	consulToken string
	err         error
}

// New constructor.
func New(cfg interface{}) *Builder {
	b := &Builder{}
	c, err := config.New(cfg)
	if err != nil {
		b.err = err
	}
	b.cfg = c
	return b
}

// WithConsul enables support for consul seed and monitor.
func (b *Builder) WithConsul(addr, dc, token string) *Builder {
	if addr == "" {
		b.err = errors.New("consul address is empty")
	}
	b.consulAddr = addr
	b.consulDC = dc
	b.consulToken = token
	return b
}

// WithConsulSeed enables support for seeding values with consul.
func (b *Builder) WithConsulSeed() *Builder {
	if b.err != nil {
		return b
	}
	// TODO: create consul seeder
	return b
}

// WithConsulMonitor enables support for monitoring key/prefixes on consul.
func (b *Builder) WithConsulMonitor(ii ...consul.Item) *Builder {
	if b.err != nil {
		return b
	}
	b.consulItems = ii

	// TODO: create consul monitor
	return b
}

// Create the harvester instance.
func (b *Builder) Create() (Harvester, error) {
	if b.err != nil {
		return nil, b.err
	}
	chErr := make(chan<- error)
	seed := seed.New(b.consulGet)
	wtc, err := consul.New(b.consulAddr, b.consulDC, b.consulToken, b.consulItems...)
	if err != nil {
		return nil, err
	}
	mon, err := monitor.New(b.cfg, wtc)
	if err != nil {
		return nil, err
	}
	return &harvester{seeder: seed, monitor: mon, chErr: chErr}, nil
}
