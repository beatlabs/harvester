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
	Name    sync.String  `seed:"John Doe"`
	Age     sync.Int64   `seed:"18" env:"ENV_AGE"`
	Balance sync.Float64 `seed:"99.9" consul:"harvester/example_03/balance"`
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	err := os.Setenv("ENV_AGE", "25")
	if err != nil {
		log.Fatalf("failed to set env var: %v", err)
	}

	seedConsulBalance("123.45")

	cfg := config{}

	ii := []consul.Item{consul.NewKeyItem("harvester/example_03/balance")}

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

	log.Printf("Config: Name: %s, Age: %d, Balance: %f\n", cfg.Name.Get(), cfg.Age.Get(), cfg.Balance.Get())

	time.Sleep(time.Second)
	seedConsulBalance("999.99")

	time.Sleep(time.Second)
	log.Printf("Config: Name: %s, Age: %d, Balance: %f\n", cfg.Name.Get(), cfg.Age.Get(), cfg.Balance.Get())
}

func seedConsulBalance(balance string) {
	cl, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("failed to create consul client: %v", err)
	}
	p := &api.KVPair{Key: "harvester/example_03/balance", Value: []byte(balance)}
	_, err = cl.KV().Put(p, nil)
	if err != nil {
		log.Fatalf("failed to put key value pair to consul: %v", err)
	}
}
