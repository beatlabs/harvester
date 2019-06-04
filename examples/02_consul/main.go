package main

import (
	"context"
	"log"

	"github.com/hashicorp/consul/api"

	"github.com/beatlabs/harvester"
)

type config struct {
	Name    string  `seed:"John Doe"`
	Age     int64   `seed:"18" env:"ENV_AGE"`
	Balance float64 `seed:"99.9" env:"ENV_CONSUL_VAR" consul:"harvester/example_02/balance"`
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

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

	log.Printf("Config: Name: %s, Age: %d, Balance: %f\n", cfg.Name, cfg.Age, cfg.Balance)
}

func seedConsulVars() {
	cl, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("failed to create consul client: %v", err)
	}
	p := &api.KVPair{Key: "harvester/example_02/balance", Value: []byte("123.45")}
	_, err = cl.KV().Put(p, nil)
	if err != nil {
		log.Fatalf("failed to put key value pair to consul: %v", err)
	}
}
