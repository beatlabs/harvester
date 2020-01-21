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
	Email     Email       `seed:"foo@example.com" env:"ENV_EMAIL"`
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

	log.Printf("Config : IndexName: %s, Email: %s, Email.Name: %s, Email.Domain: %s\n", cfg.IndexName.Get(), cfg.Email.Get(), cfg.Email.GetName(), cfg.Email.GetDomain())
}

//regex to validate an email value
const emailPattern = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"

// Email represents a custom config structure
type Email struct {
	m      gosync.RWMutex
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
