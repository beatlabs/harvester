package harvester

import (
	"context"

	"github.com/beatlabs/harvester/config"
	"github.com/beatlabs/harvester/monitor"
	"github.com/beatlabs/harvester/seed"
)

// Seeder interface for seeding initial values of the configuration.
type Seeder interface {
	Seed(*config.Config) error
}

// Monitor defines a interface for monitoring configuration changes from various sources.
type Monitor interface {
	Monitor(context.Context) error
}

// Harvester interface.
type Harvester interface {
	Harvest(context.Context) error
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

// New constructor with functional options support.
func New(cfg interface{}, ch chan<- config.ChangeNotification, oo ...OptionFunc) (Harvester, error) {
	hCfg, err := config.New(cfg, ch)
	if err != nil {
		return nil, err
	}

	opt := &options{
		cfg: hCfg,
	}

	for _, option := range oo {
		err = option(opt)
		if err != nil {
			return nil, err
		}
	}

	sd := seed.New(opt.seedParams...)

	mon, err := monitor.New(opt.cfg, opt.monitorParams...)
	if err != nil {
		return nil, err
	}

	return &harvester{seeder: sd, monitor: mon}, nil
}
