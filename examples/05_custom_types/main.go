package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	gosync "sync"

	"github.com/beatlabs/harvester"
	"github.com/beatlabs/harvester/sync"
)

type config struct {
	IndexName sync.String `seed:"customers-v1"`
	EMail     EMail       `seed:"foo@example.com" env:"ENV_EMAIL"`
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	err := os.Setenv("ENV_EMAIL", "bar@example.com")
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

	log.Printf("Config : IndexName: %s, EMail: %s, EMail.Name: %s, EMail.Domain: %s\n", cfg.IndexName.Get(), cfg.EMail.Get(), cfg.EMail.GetName(), cfg.EMail.GetDomain())
}

//regex to validate an email value
const emailPattern = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"

// EMail represents a custom config structure
type EMail struct {
	m      gosync.RWMutex
	v      string
	name   string
	domain string
}

// SetString performs basic validation and sets a config value from string typed value.
func (t *EMail) SetString(v string) error {
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
func (t *EMail) Get() string {
	t.m.RLock()
	defer t.m.RUnlock()

	return t.v
}

// GetName returns name part of the stored email.
func (t *EMail) GetName() string {
	t.m.RLock()
	defer t.m.RUnlock()

	return t.name
}

// GetDomain returns domain part of the stored email.
func (t *EMail) GetDomain() string {
	t.m.RLock()
	defer t.m.RUnlock()

	return t.domain
}

// String represents golang Stringer interface.
func (t *EMail) String() string {
	return t.Get()
}
