package harvester

import (
	"context"
	"errors"

	"github.com/taxibeat/harvester/config"
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

// New constructor.
func New(s Seeder, m Monitor, chErr chan<- error) (Harvester, error) {
	if s == nil {
		return nil, errors.New("seeder is nil")
	}
	if m == nil {
		return nil, errors.New("monitor is nil")
	}
	return &harvester{seeder: s, monitor: m, chErr: chErr}, nil
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
