package harvester

import (
	"context"
	"sync"
)

// Monitor defines a monitoring interface.
type Monitor interface {
	Monitor(ctx context.Context, cfg interface{})
}

// MonitorFactory defines a monitoring interface.
type MonitorFactory interface {
	Create(ctx context.Context, cfg interface{}) (Monitor, error)
}

// Harvester interface.
type Harvester interface {
	Harvest(ctx context.Context, cfg interface{}) error
}

type harvester struct {
	mf MonitorFactory
}

// New constructor.
func New() (Harvester, error) {
	return &harvester{}, nil
}

func (h *harvester) Harvest(ctx context.Context, cfg interface{}) error {
	wg := sync.WaitGroup{}

	mon, err := h.mf.Create(ctx, cfg)
	if err != nil {
		return err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		mon.Monitor(ctx, cfg)
	}()

	// TODO: Watch

	wg.Wait()
	return nil
}
