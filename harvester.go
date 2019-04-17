package harvester

import (
	"context"

	"github.com/taxibeat/harvester/config"
)

// Monitor defines a monitoring interface.
type Monitor interface {
	Monitor(ctx context.Context, cfg interface{})
}

// Harvester interface.
type Harvester interface {
	Harvest(ctx context.Context, cfg interface{}) error
}

type harvester struct {
	cfg *config.Config
}

// New constructor.
func New() (Harvester, error) {

	//TODO: support optional consul parameters (address etc.)

	return &harvester{}, nil
}

// Harvest take the configuration object, initializes it and monitors for changes.
func (h *harvester) Harvest(ctx context.Context, cfg interface{}) error {

	// TODO: create the Value object

	// TODO: initialize the value object

	// TODO: monitor and change the value object in a goroutine

	return nil
}
