package harvester

import (
	"context"
)

// Monitorer defines a monitoring interface.
type Monitorer interface {
	Monitor(ctx context.Context)
}

// Harvester interface.
type Harvester interface {
	Harvest(ctx context.Context, cfg interface{}) error
}
