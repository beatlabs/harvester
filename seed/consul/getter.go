// Package consul handles seeding capabilities with consul.
package consul

import (
	"errors"
	"path"
	"time"

	"github.com/hashicorp/consul/api"
)

// Getter implementation of the getter interface.
type Getter struct {
	kv           *api.KV
	dc           string
	token        string
	folderPrefix string
}

// New constructor. Timeout is set to 60s when 0 is provided.
func New(addr, dc, token string, timeout time.Duration) (*Getter, error) {
	return NewWithFolderPrefix(addr, dc, token, "", timeout)
}

// NewWithFolderPrefix constructor. Timeout is set to 60s when 0 is provided.
func NewWithFolderPrefix(addr, dc, token, folderPrefix string, timeout time.Duration) (*Getter, error) {
	if addr == "" {
		return nil, errors.New("address is empty")
	}
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	config := api.DefaultConfig()
	config.Address = addr

	var err error
	config.HttpClient, err = api.NewHttpClient(config.Transport, config.TLSConfig)
	if err != nil {
		return nil, err
	}
	config.HttpClient.Timeout = timeout

	consul, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Getter{kv: consul.KV(), dc: dc, token: token, folderPrefix: folderPrefix}, nil
}

// Get the specific key value from consul.
func (g *Getter) Get(key string) (*string, uint64, error) {
	pair, _, err := g.kv.Get(path.Join(g.folderPrefix, key), &api.QueryOptions{Datacenter: g.dc, Token: g.token})
	if err != nil {
		return nil, 0, err
	}
	if pair == nil {
		return nil, 0, nil
	}
	val := string(pair.Value)
	return &val, pair.ModifyIndex, nil
}
