package harvester

import (
	"context"
	"sync"
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
	mon Monitor
}

// New constructor.
func New() (Harvester, error) {
	return &harvester{}, nil
}

func (h *harvester) Harvest(ctx context.Context, cfg interface{}) error {

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.mon.Monitor(ctx, cfg)
	}()

	wg.Wait()
	return nil
}
