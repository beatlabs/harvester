package main

import (
	"context"
	"log"
	"os"

	"github.com/beatlabs/harvester"
	"github.com/beatlabs/harvester/sync"
)

type config struct {
	Name sync.String `seed:"John Doe"`
	Age  sync.Int64  `seed:"18" env:"ENV_AGE"`
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	err := os.Setenv("ENV_AGE", "25")
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

	log.Printf("Config : Name: %s, Age: %d\n", cfg.Name.Get(), cfg.Age.Get())
}
