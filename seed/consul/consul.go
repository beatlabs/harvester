package consul

import (
	"errors"

	"github.com/hashicorp/consul/api"
)

// Getter implementation of the getter interface.
type Getter struct {
	kv    *api.KV
	dc    string
	token string
}

// New constructor.
func New(addr, dc, token string) (*Getter, error) {
	if addr == "" {
		return nil, errors.New("address is empty")
	}
	config := api.DefaultConfig()
	config.Address = addr
	consul, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Getter{kv: consul.KV()}, nil
}

// Get the specific key value from consul.
func (g *Getter) Get(key string) (string, error) {
	pair, _, err := g.kv.Get(key, &api.QueryOptions{Datacenter: g.dc, Token: g.token})
	if err != nil {
		return "", err
	}
	return string(pair.Value), nil
}
