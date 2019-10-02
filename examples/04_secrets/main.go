package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/beatlabs/harvester"
	"github.com/beatlabs/harvester/monitor/consul"
	"github.com/beatlabs/harvester/sync"
	"github.com/hashicorp/consul/api"
)

type config struct {
	IndexName      sync.String `seed:"customers-v1"`
	CacheRetention sync.Int64  `seed:"43200" env:"ENV_CACHE_RETENTION_SECONDS"`
	LogLevel       sync.String `seed:"DEBUG" flag:"loglevel"`
	AccessToken    sync.Secret `seed:"defaultaccesstoken" consul:"harvester/example_04/accesstoken"`
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	err := os.Setenv("ENV_CACHE_RETENTION_SECONDS", "86400")
	if err != nil {
		log.Fatalf("failed to set env var: %v", err)
	}

	seedConsulAccessToken("currentaccesstoken")

	cfg := config{}

	ii := []consul.Item{
		consul.NewKeyItem("harvester/example_04/accesstoken"),
	}

	h, err := harvester.New(&cfg).
		WithConsulSeed("127.0.0.1:8500", "", "", 0).
		WithConsulMonitor("127.0.0.1:8500", "", "", 0, ii...).
		Create()
	if err != nil {
		log.Fatalf("failed to create harvester: %v", err)
	}

	err = h.Harvest(ctx)
	if err != nil {
		log.Fatalf("failed to harvest configuration: %v", err)
	}

	log.Printf("Config: IndexName: %s, CacheRetention: %d, LogLevel: %s, AccessToken: %s\n", cfg.IndexName.Get(), cfg.CacheRetention.Get(), cfg.LogLevel.Get(), cfg.AccessToken.Get())

	time.Sleep(time.Second)
	seedConsulAccessToken("newaccesstoken")

	time.Sleep(time.Second)
	log.Printf("Config: IndexName: %s, CacheRetention: %d, LogLevel: %s, AccessToken: %s\n", cfg.IndexName.Get(), cfg.CacheRetention.Get(), cfg.LogLevel.Get(), cfg.AccessToken.Get())
}

func seedConsulAccessToken(accessToken string) {
	cl, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("failed to create consul client: %v", err)
	}
	p := &api.KVPair{Key: "harvester/example_04/accesstoken", Value: []byte(accessToken)}
	_, err = cl.KV().Put(p, nil)
	if err != nil {
		log.Fatalf("failed to put key value pair to consul: %v", err)
	}
}
