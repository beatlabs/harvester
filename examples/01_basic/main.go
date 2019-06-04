package main

import (
	"context"
	"log"

	"github.com/taxibeat/harvester"
)

type config struct {
	Name string `seed:"John Doe"`
	Age  int64  `seed:"18" env:"ENV_AGE"`
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()
	cfg := config{}

	h, err := harvester.New(&cfg).Create()
	if err != nil {
		log.Fatalf("failed to create harvester: %v", err)
	}

	err = h.Harvest(ctx)
	if err != nil {
		log.Fatalf("failed to harvest configuration: %v", err)
	}

	log.Printf("Config : Name: %s, Age: %d\n", cfg.Name, cfg.Age)
}
