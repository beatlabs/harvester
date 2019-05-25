package main

import (
	"fmt"
	"log"
	"reflect"

	"github.com/taxibeat/harvester/config"
	"github.com/taxibeat/harvester/seed"
)

// ConfigAttrs sample config struct
type ConfigAttrs struct {
	Name string `seed:"John Doe"`
	Age  int64  `seed:"18" env:"ENV_AGE"`
}

func main() {
	attrs := ConfigAttrs{}

	cfg, err := config.New(&attrs)
	if err != nil {
		fmt.Printf("Oops, something went wrong creating config: %v", err)
	}

	printAttrs(attrs)

	s := seed.New()
	s.Seed(cfg)

	printAttrs(attrs)

	fmt.Println("Setting name to: boo")
	err = cfg.Set("Name", "boo", reflect.String)
	if err != nil {
		fmt.Printf("Oops, something went wrong setting field: %v", err)
	}

	printAttrs(attrs)
}

func printAttrs(attrs ConfigAttrs) {
	log.Printf("Attribute State: name: %s, age: %d\n", attrs.Name, attrs.Age)
}
