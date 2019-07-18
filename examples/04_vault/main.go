package main

import (
	"context"
	"log"
	"time"

	"github.com/beatlabs/harvester"
	"github.com/beatlabs/harvester/sync"
	"github.com/hashicorp/vault/api"
)

type config struct {
	Secret sync.String `vault:"secret/data/harvester/example_04/app_secret"`
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	address := "http://0.0.0.0:8200"
	token := "root"
	createSecretInVault(address, token, "secret/data/harvester/example_04", "app_secret", "@pps3cr3t")

	cfg := config{}
	h, err := harvester.New(&cfg).
		WithVaultSeed(address, token, 0).
		Create()
	if err != nil {
		log.Fatalf("failed to create harvester: %v", err)
	}

	err = h.Harvest(ctx)
	if err != nil {
		log.Fatalf("failed to harvest configuration: %v", err)
	}

	log.Printf("Secret: %s", cfg.Secret.Get())
}

func createSecretInVault(address, token, path, secretKey, secretValue string) {
	cfg := api.DefaultConfig()
	cfg.Address = address
	cfg.Timeout = 5 * time.Second
	client, err := api.NewClient(cfg)
	if err != nil {
		log.Fatalf("failed to create Vault client: %v", err)
	}
	client.SetToken(token)

	_, err = client.Logical().Delete(path)
	if err != nil {
		log.Fatalf("could not reset Vault key: %v", err)
	}

	_, err = client.Logical().Write(
		path,
		map[string]interface{}{
			"data": map[string]interface{}{secretKey: secretValue},
		},
	)
	if err != nil {
		log.Fatalf("could not write Vault key: %v", err)
	}
}
