package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/consul/api"

	"github.com/taxibeat/harvester"
)

type configAttrs struct {
	Name      string `seed:"John Doe"`
	Age       int64  `seed:"18" env:"ENV_AGE"`
	ConsulVar string `seed:"b" env:"ENV_CONSUL_VAR" consul:"/harvester/example_02/consul_var"`
}

func main() {
	consulConfig := api.DefaultConfig()
	consulConfig.Token = "token"

	seedConsulVars(consulConfig)

	attrs := configAttrs{}

	h, err := harvester.New(&attrs).
		WithConsulSeed(consulConfig.Address, consulConfig.Datacenter, consulConfig.Token, 0).
		Create()
	if err != nil {
		fmt.Printf("Oops, something went wrong creating harvester instance: %v", err)
	}

	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()
	h.Harvest(ctx)

	printAttrs(attrs)
}

func seedConsulVars(consulConfig *api.Config) {
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		panic(err)
	}
	// Get a handle to the KV API
	kv := consulClient.KV()

	// PUT a new KV pair
	p := &api.KVPair{Key: "harvester/example_02/consul_var", Value: []byte("boo")}
	_, err = kv.Put(p, nil)
	if err != nil {
		panic(err)
	}
}

func printAttrs(attrs configAttrs) {
	log.Printf("Attribute State: name: %s, age: %d, consulVar: %s\n", attrs.Name, attrs.Age, attrs.ConsulVar)
}
