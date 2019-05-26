package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/taxibeat/harvester"
	"github.com/taxibeat/harvester/monitor/consul"
)

type configAttrs struct {
	Name      string `seed:"John Doe"`
	Age       int64  `seed:"18" env:"ENV_AGE"`
	ConsulVar string `seed:"b" env:"ENV_CONSUL_VAR" consul:"/harvester/example_03/consul_var"`
}

func main() {
	consulConfig := api.DefaultConfig()
	consulConfig.Token = "token"

	seedConsulVars(consulConfig, "bar")

	attrs := configAttrs{}

	ii := []consul.Item{consul.NewKeyItem("/harvester/example_03/consul_var"), consul.NewPrefixItem("")}

	h, err := harvester.New(&attrs).
		WithConsulSeed(consulConfig.Address, consulConfig.Datacenter, consulConfig.Token, 0).
		WithConsulMonitor(consulConfig.Address, consulConfig.Datacenter, consulConfig.Token, 0, ii...).
		Create()
	if err != nil {
		fmt.Printf("Oops, something went wrong creating harvester instance: %v", err)
	}

	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()
	h.Harvest(ctx)

	printAttrs(attrs)

	time.Sleep(5 * time.Second)
	seedConsulVars(consulConfig, "baz")

	time.Sleep(5 * time.Second)
	printAttrs(attrs)
}

func seedConsulVars(consulConfig *api.Config, consulVarValue string) {
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		panic(err)
	}
	// Get a handle to the KV API
	kv := consulClient.KV()

	// PUT a new KV pair
	p := &api.KVPair{Key: "harvester/example_03/consul_var", Value: []byte(consulVarValue)}
	_, err = kv.Put(p, nil)
	if err != nil {
		panic(err)
	}
}

func printAttrs(attrs configAttrs) {
	log.Printf("Attribute State: name: %s, age: %d, consulVar: %s\n", attrs.Name, attrs.Age, attrs.ConsulVar)
}
