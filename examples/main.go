package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/beatlabs/harvester"
	harvestersync "github.com/beatlabs/harvester/sync"
	"github.com/hashicorp/consul/api"
)

type config struct {
	IndexName      harvestersync.String  `seed:"customers-v1"`
	CacheRetention harvestersync.Int64   `seed:"43200" env:"ENV_CACHE_RETENTION_SECONDS"`
	LogLevel       harvestersync.String  `seed:"DEBUG" flag:"loglevel"`
	OpeningBalance harvestersync.Float64 `seed:"0.0" env:"ENV_CONSUL_VAR" consul:"harvester/example_02/openingbalance"`
	AccessToken    harvestersync.Secret  `seed:"defaultaccesstoken" consul:"harvester/example_04/accesstoken"`
	Email          Email                 `seed:"foo@example.com" env:"ENV_EMAIL"`
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	setEnvVars()

	seedConsulVars("currentaccesstoken")

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

	log.Printf("Config: IndexName: %s, CacheRetention: %d, LogLevel: %s, OpeningBalance: %f\n", cfg.IndexName.Get(), cfg.CacheRetention.Get(), cfg.LogLevel.Get(), cfg.OpeningBalance.Get())
}

func setEnvVars() {
	err := os.Setenv("ENV_CACHE_RETENTION_SECONDS", "86400")
	if err != nil {
		log.Fatalf("failed to set env var: %v", err)
	}

	err = os.Setenv("ENV_EMAIL", "bar@example.com")
	if err != nil {
		log.Fatalf("failed to set env var: %v", err)
	}
}

func seedConsulVars(accessToken string) {
	cl, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("failed to create consul client: %v", err)
	}
	p := &api.KVPair{Key: "harvester/example_02/openingbalance", Value: []byte("100.0")}
	_, err = cl.KV().Put(p, nil)
	if err != nil {
		log.Fatalf("failed to put key value pair to consul: %v", err)
	}

	p = &api.KVPair{Key: "harvester/example_04/accesstoken", Value: []byte(accessToken)}
	_, err = cl.KV().Put(p, nil)
	if err != nil {
		log.Fatalf("failed to put key value pair to consul: %v", err)
	}
}

// regex to validate an email value.
const emailPattern = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"

// Email represents a custom config structure.
type Email struct {
	m      sync.RWMutex
	v      string
	name   string
	domain string
}

// SetString performs basic validation and sets a config value from string typed value.
func (t *Email) SetString(v string) error {
	re := regexp.MustCompile(emailPattern)
	if !re.MatchString(v) {
		return fmt.Errorf("%s is not a valid email address", v)
	}

	t.m.Lock()
	defer t.m.Unlock()

	t.v = v
	parts := strings.Split(v, "@")
	t.name = parts[0]
	t.domain = parts[1]

	return nil
}

// Get returns the stored value.
func (t *Email) Get() string {
	t.m.RLock()
	defer t.m.RUnlock()

	return t.v
}

// GetName returns name part of the stored email.
func (t *Email) GetName() string {
	t.m.RLock()
	defer t.m.RUnlock()

	return t.name
}

// GetDomain returns domain part of the stored email.
func (t *Email) GetDomain() string {
	t.m.RLock()
	defer t.m.RUnlock()

	return t.domain
}

// String represents golang Stringer interface.
func (t *Email) String() string {
	return t.Get()
}
