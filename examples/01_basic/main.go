package main

import (
	"context"
	"log"
	"os"

	"github.com/beatlabs/harvester"
	"github.com/beatlabs/harvester/sync"
)

type config struct {
	IndexName      sync.String `seed:"customers-v1"`
	CacheRetention sync.Int64  `seed:"43200" env:"ENV_CACHE_RETENTION_SECONDS"`
	LogLevel       sync.String `seed:"DEBUG" flag:"loglevel"`
	DbPassword     sync.String `seed:"mylocalpassword" env:"ENV_DB_PASSWORD" secret:"true"`
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	err := os.Setenv("ENV_CACHE_RETENTION_SECONDS", "86400")
	if err != nil {
		log.Fatalf("failed to set env var: %v", err)
	}
	err = os.Setenv("ENV_DB_PASSWORD", "dfs89sSFD*89SDFV7VSD6^SDVvvss")
	if err != nil {
		log.Fatalf("failed to set env var: %v", err)
	}

	cfg := config{}

	h, err := harvester.New(&cfg).Create()
	if err != nil {
		log.Fatalf("failed to create harvester: %v", err)
	}

	err = h.Harvest(ctx)
	if err != nil {
		log.Fatalf("failed to harvest configuration: %v", err)
	}

	log.Printf("Config : IndexName: %s, CacheRetention: %d, LogLevel: %s\n", cfg.IndexName.Get(), cfg.CacheRetention.Get(), cfg.LogLevel.Get())
}
