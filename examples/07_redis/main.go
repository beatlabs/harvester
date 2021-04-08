package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/beatlabs/harvester"
	"github.com/beatlabs/harvester/sync"
	"github.com/go-redis/redis/v8"
)

type config struct {
	IndexName      sync.String  `seed:"customers-v1"`
	CacheRetention sync.Int64   `seed:"43200" env:"ENV_CACHE_RETENTION_SECONDS"`
	LogLevel       sync.String  `seed:"DEBUG" flag:"loglevel"`
	OpeningBalance sync.Float64 `seed:"0.0" env:"ENV_CONSUL_VAR" redis:"opening-balance"`
}

var redisClient = redis.NewClient(&redis.Options{})

func main() {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	err := os.Setenv("ENV_CACHE_RETENTION_SECONDS", "86400")
	if err != nil {
		log.Fatalf("failed to set env var: %v", err)
	}

	cfg := config{}

	err = setBalance(ctx, "1000")
	if err != nil {
		log.Fatalf("failed to seed balance in redis: %v", err)
	}

	h, err := harvester.New(&cfg).WithRedisSeed(redisClient).WithRedisMonitor(redisClient, 200*time.Millisecond).Create()
	if err != nil {
		log.Fatalf("failed to create harvester: %v", err)
	}

	err = h.Harvest(ctx)
	if err != nil {
		log.Fatalf("failed to harvest configuration: %v", err)
	}

	log.Printf("Initial Config: IndexName: %s, CacheRetention: %d, LogLevel: %s, OpeningBalance: %f\n",
		cfg.IndexName.Get(), cfg.CacheRetention.Get(), cfg.LogLevel.Get(), cfg.OpeningBalance.Get())

	err = setBalance(ctx, "2000")
	if err != nil {
		log.Fatalf("failed to change balance in redis: %v", err)
	}

	time.Sleep(1 * time.Second)

	log.Printf("Change balance. Config: IndexName: %s, CacheRetention: %d, LogLevel: %s, OpeningBalance: %f\n",
		cfg.IndexName.Get(), cfg.CacheRetention.Get(), cfg.LogLevel.Get(), cfg.OpeningBalance.Get())

	err = setBalance(ctx, "1000")
	if err != nil {
		log.Fatalf("failed to change balance in redis: %v", err)
	}

	time.Sleep(1 * time.Second)

	log.Printf("Revert balance. Config: IndexName: %s, CacheRetention: %d, LogLevel: %s, OpeningBalance: %f\n",
		cfg.IndexName.Get(), cfg.CacheRetention.Get(), cfg.LogLevel.Get(), cfg.OpeningBalance.Get())
}

func setBalance(ctx context.Context, amount string) error {
	_, err := redisClient.Set(ctx, "opening-balance", amount, 0).Result()
	if err != nil {
		return err
	}
	return nil
}
