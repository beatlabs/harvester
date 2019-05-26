package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taxibeat/harvester"
)

type configAttrs struct {
	Name string `seed:"John Doe"`
	Age  int64  `seed:"18" env:"ENV_AGE"`
}

func main() {
	attrs := configAttrs{
		Name: "Jim",
	}

	h, err := harvester.New(&attrs).Create()
	if err != nil {
		fmt.Printf("Oops, something went wrong creating harvester instance: %v", err)
	}
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()
	h.Harvest(ctx)

	printAttrs(attrs)
}

func printAttrs(attrs configAttrs) {
	log.Printf("Attribute State: name: %s, age: %d\n", attrs.Name, attrs.Age)
}
