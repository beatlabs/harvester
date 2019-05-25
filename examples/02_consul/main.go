package main

import (
	"log"

	"github.com/hashicorp/consul/api"

	"github.com/taxibeat/harvester/config"
	"github.com/taxibeat/harvester/seed"
	"github.com/taxibeat/harvester/seed/consul"
)

type fields struct {
	consulParam *seed.Param
}

// ConfigAttrs sample config struct
type ConfigAttrs struct {
	Name      string `seed:"John Doe"`
	Age       int64  `seed:"18" env:"ENV_AGE"`
	ConsulVar string `seed:"b" env:"ENV_CONSUL_VAR" consul:"/harvester/example_02/consul_var"`
}

func main() {
	consulConfig := api.DefaultConfig()
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		panic(err)
	}
	// Get a handle to the KV API
	kv := consulClient.KV()

	// PUT a new KV pair
	p := &api.KVPair{Key: "harvester/example_02/consul_var", Value: []byte("bar")}
	_, err = kv.Put(p, nil)
	if err != nil {
		panic(err)
	}

	attrs := ConfigAttrs{}

	cfg, err := config.New(&attrs)

	printAttrs(attrs)

	getter, err := consul.New(consulConfig.Address, "", "token", 0)

	prmFoo, err := seed.NewParam(config.SourceConsul, getter)

	s := seed.New(*prmFoo)
	s.Seed(cfg)
	printAttrs(attrs)
}

func printAttrs(attrs ConfigAttrs) {
	log.Printf("Attribute State: name: %s, age: %d, consulVar: %s\n", attrs.Name, attrs.Age, attrs.ConsulVar)
}
