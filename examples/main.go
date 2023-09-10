package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/beatlabs/harvester"
	harvesterconfig "github.com/beatlabs/harvester/config"
	harvestersync "github.com/beatlabs/harvester/sync"
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/consul/api"
)

const (
	consulAddress = "127.0.0.1:8500"
	consulDC      = ""
	consulToken   = ""
)

type config struct {
	// IndexName demonstrates only seed.
	IndexName harvestersync.String `seed:"customers-v1"`
	// CacheRetention demonstrates seed and env var.
	CacheRetention harvestersync.Int64 `seed:"43200" env:"ENV_CACHE_RETENTION_SECONDS"`
	// LogLevel demonstrates seed and flag.
	LogLevel harvestersync.String `seed:"DEBUG" flag:"loglevel"`
	// OpeningBalance demonstrates seed, env var and redis.
	OpeningBalance harvestersync.Float64 `seed:"0.0" redis:"opening-balance"`
	// AccessToken demonstrates seed and consul for a secret.
	AccessToken harvestersync.Secret `seed:"defaultaccesstoken" consul:"harvester/example/accesstoken"`
	// Email demonstrates seed for a custom type.
	Email Email `seed:"foo@example.com"`
}

func (c *config) String() string {
	return fmt.Sprintf("config: IndexName: %s CacheRetention: %d LogLevel: %s OpeningBalance: %f AccessToken: %s Email: %s",
		c.IndexName.Get(), c.CacheRetention.Get(), c.LogLevel.Get(), c.OpeningBalance.Get(), c.AccessToken.Get(),
		c.Email.String())
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()

	setEnvVarCacheRetention()
	seedConsulAccessToken("currentaccesstoken")
	setRedisOpeningBalance(ctx, "1000")

	cfg := config{}

	chNotify := make(chan harvesterconfig.ChangeNotification)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for change := range chNotify {
			log.Printf("notification: " + change.String())
		}
		wg.Done()
	}()

	redisClient := createRedisClient()

	h, err := harvester.New(&cfg).
		WithConsulSeed(consulAddress, consulDC, consulToken, 0).WithConsulMonitor(consulAddress, consulDC, consulToken, 0).
		WithRedisSeed(redisClient).WithRedisMonitor(redisClient, 200*time.Millisecond).
		WithNotification(chNotify).Create()
	if err != nil {
		log.Fatalf("failed to create harvester: %v", err)
	}

	err = h.Harvest(ctx)
	if err != nil {
		log.Fatalf("failed to harvest configuration: %v", err)
	}

	log.Println(cfg.String())

	seedConsulAccessToken("newtaccesstoken")
	setRedisOpeningBalance(ctx, "2000")

	time.Sleep(1 * time.Second) // Wait for the data to be updated async...

	log.Println(cfg.String())
}

func setEnvVarCacheRetention() {
	err := os.Setenv("ENV_CACHE_RETENTION_SECONDS", "86400")
	if err != nil {
		log.Fatalf("failed to set env var: %v", err)
	}
}

func seedConsulAccessToken(accessToken string) {
	cl, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("failed to create consul client: %v", err)
	}

	p := &api.KVPair{Key: "harvester/example/accesstoken", Value: []byte(accessToken)}
	_, err = cl.KV().Put(p, nil)
	if err != nil {
		log.Fatalf("failed to put key value pair to consul: %v", err)
	}
}

func setRedisOpeningBalance(ctx context.Context, amount string) error {
	_, err := createRedisClient().Set(ctx, "opening-balance", amount, 0).Result()
	if err != nil {
		return err
	}
	return nil
}

func createRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{})
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
