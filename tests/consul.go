package tests

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/ory/dockertest/v3"
)

// ConsulRuntime is the main runtime struct.
type ConsulRuntime struct {
	consul *dockertest.Resource
	pool   *dockertest.Pool
	port   string
	tag    string
}

// NewConsulRuntime creates a new Consul dockertest runtime.
func NewConsulRuntime(version string) (*ConsulRuntime, error) {
	pool, err := createPool()
	if err != nil {
		return nil, err
	}
	return &ConsulRuntime{pool: pool, tag: version}, nil
}

// StartUp starts the consul container in the Runtime.
func (r *ConsulRuntime) StartUp() error {
	var err error
	r.consul, err = r.pool.RunWithOptions(&dockertest.RunOptions{Repository: "consul", Tag: r.tag})
	if err != nil {
		return err
	}
	r.port = r.consul.GetPort("8500/tcp")
	return r.Ready()
}

// Ready waits for the consul container to be healthy.
func (r *ConsulRuntime) Ready() error {
	return r.pool.Retry(func() error {
		health, err := http.Get(fmt.Sprintf("http://localhost:%s/v1/health/node/any", r.port))
		if err == nil && health.StatusCode == 200 {
			return nil
		}
		return err
	})
}

// TearDown purges the consul container in the Runtime.
func (r *ConsulRuntime) TearDown() error {
	return r.pool.Purge(r.consul)
}

// GetAddress returns the address of the Consul docker container.
func (r *ConsulRuntime) GetAddress() string {
	return fmt.Sprintf("localhost:%s", r.port)
}

// GetClient returns a Consul Client.
func (r *ConsulRuntime) GetClient() (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = r.GetAddress()
	return api.NewClient(config)
}

func createPool() (*dockertest.Pool, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("could not create new docker pool: %v", err)
	}
	pool.MaxWait = time.Minute * 2
	return pool, nil
}
