package main

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/beatlabs/harvester"
	harvestersync "github.com/beatlabs/harvester/sync"
)

type config struct {
	IndexName      harvestersync.String `seed:"customers-v1"`
	CacheRetention harvestersync.Int64  `seed:"43200" env:"ENV_CACHE_RETENTION_SECONDS"`
	LogLevel       harvestersync.String `seed:"DEBUG" flag:"loglevel"`
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	err := os.Setenv("ENV_CACHE_RETENTION_SECONDS", "86400")
	if err != nil {
		log.Fatalf("failed to set env var: %v", err)
	}

	cfg := config{}
	chNotify := make(chan string)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for change := range chNotify {
			log.Printf("notification: " + change)
		}
		wg.Done()
	}()

	h, err := harvester.New(&cfg).WithNotification(chNotify).Create()
	if err != nil {
		log.Fatalf("failed to create harvester: %v", err)
	}

	err = h.Harvest(ctx)
	if err != nil {
		log.Fatalf("failed to harvest configuration: %v", err)
	}

	log.Printf("Config : IndexName: %s, CacheRetention: %d, LogLevel: %s\n", cfg.IndexName.Get(), cfg.CacheRetention.Get(), cfg.LogLevel.Get())
	close(chNotify)
	wg.Wait()
}
