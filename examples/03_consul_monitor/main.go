package main

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/beatlabs/harvester"
	"github.com/beatlabs/harvester/monitor/consul"
)

type config struct {
	Name    string  `seed:"John Doe"`
	Age     int64   `seed:"18" env:"ENV_AGE"`
	Balance float64 `seed:"99.9" env:"ENV_CONSUL_VAR" consul:"harvester/example_03/balance"`
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

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

	log.Printf("Config: Name: %s, Age: %d, Balance: %f\n", cfg.Name, cfg.Age, cfg.Balance)

	time.Sleep(time.Second)
	seedConsulBalance("999.99")

	time.Sleep(time.Second)
	log.Printf("Config: Name: %s, Age: %d, Balance: %f\n", cfg.Name, cfg.Age, cfg.Balance)
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
