package main

import (
	"context"
	"log"
	"os"

	"github.com/beatlabs/harvester"
	"github.com/beatlabs/harvester/sync"
	"github.com/hashicorp/consul/api"
)

type config struct {
	IndexName      sync.String  `seed:"customers-v1"`
	CacheRetention sync.Int64   `seed:"43200" env:"ENV_CACHE_RETENTION_SECONDS"`
	LogLevel       sync.String  `seed:"DEBUG" flag:"loglevel"`
	OpeningBalance sync.Float64 `seed:"0.0" env:"ENV_CONSUL_VAR" consul:"harvester/example_02/openingbalance"`
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	err := os.Setenv("ENV_CACHE_RETENTION_SECONDS", "86400")
	if err != nil {
		log.Fatalf("failed to set env var: %v", err)
	}

	seedConsulVars()

	cfg := config{}

	h, err := harvester.New(&cfg).
		WithConsulSeed("127.0.0.1:8500", "", "", 0).
		Create()
	if err != nil {
		log.Fatalf("failed to create harvester: %v", err)
	}

	err = h.Harvest(ctx)
	if err != nil {
		log.Fatalf("failed to harvest configuration: %v", err)
	}

	log.Printf("Config: IndexName: %s, CacheRetention: %d, LogLevel: %s, OpeningBalance: %f\n", cfg.IndexName.Get(), cfg.CacheRetention.Get(), cfg.LogLevel.Get(), cfg.OpeningBalance.Get())
}

func seedConsulVars() {
	cl, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("failed to create consul client: %v", err)
	}
	p := &api.KVPair{Key: "harvester/example_02/openingbalance", Value: []byte("100.0")}
	_, err = cl.KV().Put(p, nil)
	if err != nil {
		log.Fatalf("failed to put key value pair to consul: %v", err)
	}
}
