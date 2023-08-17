package main

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/beatlabs/harvester"
	"github.com/beatlabs/harvester/config"
	harvestersync "github.com/beatlabs/harvester/sync"
)

type cfg struct {
	IndexName      harvestersync.String `seed:"customers-v1"`
	CacheRetention harvestersync.Int64  `seed:"43200" env:"ENV_CACHE_RETENTION_SECONDS"`
	LogLevel       harvestersync.String `seed:"DEBUG" flag:"loglevel"`
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	err := os.Setenv("ENV_CACHE_RETENTION_SECONDS", "86400")
	if err != nil {
		slog.Error("failed to set env var", "err", err)
		os.Exit(1)
	}

	cfg := cfg{}
	chNotify := make(chan config.ChangeNotification)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for change := range chNotify {
			slog.Info("notification", "change", change.String())
		}
		wg.Done()
	}()

	h, err := harvester.New(&cfg).WithNotification(chNotify).Create()
	if err != nil {
		slog.Error("failed to create harvester", "err", err)
		os.Exit(1)
	}

	err = h.Harvest(ctx)
	if err != nil {
		slog.Error("failed to harvest configuration", "err", err)
		os.Exit(1)
	}

	slog.Info("config", "index", cfg.IndexName.Get(), "retention", cfg.CacheRetention.Get(), "level", cfg.LogLevel.Get())
	close(chNotify)
	wg.Wait()
}
