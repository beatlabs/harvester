package consul

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/beatlabs/harvester/change"
	"github.com/beatlabs/harvester/config"
	harvesterlog "github.com/beatlabs/harvester/log"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

// Item definition.
type Item struct {
	tp  string
	key string
}

// NewKeyItem creates a new key watch item for the watcher.
func NewKeyItem(key string) Item {
	return Item{tp: "key", key: key}
}

// NewPrefixItem creates a prefix key watch item for the watcher.
func NewPrefixItem(key string) Item {
	return Item{tp: "keyprefix", key: key}
}

// Watcher of Consul changes.
type Watcher struct {
	cl    *api.Client
	dc    string
	token string
	pp    []*watch.Plan
	ii    []Item
}

// New creates a new watcher.
func New(addr, dc, token string, timeout time.Duration, ii ...Item) (*Watcher, error) {
	if addr == "" {
		return nil, errors.New("address is empty")
	}
	if len(ii) == 0 {
		return nil, errors.New("items are empty")
	}
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	cfg := api.DefaultConfig()
	cfg.Address = addr
	var err error
	cfg.HttpClient, err = api.NewHttpClient(cfg.Transport, cfg.TLSConfig)
	if err != nil {
		return nil, err
	}
	cfg.HttpClient.Timeout = timeout

	cl, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Watcher{cl: cl, dc: dc, token: token, ii: ii}, nil
}

// Watch key and prefixes for changes.
func (w *Watcher) Watch(ctx context.Context, ch chan<- []*change.Change) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	if ch == nil {
		return errors.New("change channel is nil")
	}
	for _, i := range w.ii {
		var pl *watch.Plan
		var err error
		switch i.tp {
		case "key":
			pl, err = w.createKeyPlan(i.key, ch)
		case "keyprefix":
			pl, err = w.createKeyPrefixPlan(i.key, ch)
		}
		if err != nil {
			return err
		}
		w.pp = append(w.pp, pl)
		go func(tp, key string) {
			logger := log.New(harvesterlog.Writer(), "", 0)
			err := pl.RunWithClientAndLogger(w.cl, logger)
			if err != nil {
				harvesterlog.Errorf("plan %s of type %s failed: %v", tp, key, err)
			} else {
				harvesterlog.Infof("plan %s of type %s is running", tp, key)
			}
		}(i.tp, i.key)
	}
	go func() {
		<-ctx.Done()
		for _, pl := range w.pp {
			pl.Stop()
		}
		harvesterlog.Infof("all watch plans have been stopped")
	}()

	return nil
}

func (w *Watcher) createKeyPlan(key string, ch chan<- []*change.Change) (*watch.Plan, error) {
	pl, err := w.getPlan("key", key)
	if err != nil {
		return nil, err
	}
	pl.Handler = func(idx uint64, data interface{}) {
		pair, ok := data.(*api.KVPair)
		if !ok {
			harvesterlog.Errorf("data is not kv pair: %v", data)
		} else {
			ch <- []*change.Change{change.New(config.SourceConsul, pair.Key, string(pair.Value), pair.ModifyIndex)}
		}
	}
	harvesterlog.Infof("plan for key %s created", key)
	return pl, nil
}

func (w *Watcher) createKeyPrefixPlan(keyPrefix string, ch chan<- []*change.Change) (*watch.Plan, error) {
	pl, err := w.getPlan("keyprefix", keyPrefix)
	if err != nil {
		return nil, err
	}
	pl.Handler = func(idx uint64, data interface{}) {
		pp, ok := data.(api.KVPairs)
		if !ok {
			harvesterlog.Errorf("data is not kv pairs: %v", data)
		} else {
			cc := make([]*change.Change, len(pp))
			for i := 0; i < len(pp); i++ {
				cc[i] = change.New(config.SourceConsul, pp[i].Key, string(pp[i].Value), pp[i].ModifyIndex)
			}
			ch <- cc
		}
	}
	harvesterlog.Infof("plan for keyprefix %s created", keyPrefix)
	return pl, nil
}

func (w *Watcher) getPlan(tp, key string) (*watch.Plan, error) {
	params := map[string]interface{}{}
	params["datacenter"] = w.dc
	params["token"] = w.token
	if tp == "key" {
		params["key"] = key
	} else {
		params["prefix"] = key
	}
	params["type"] = tp
	return watch.Parse(params)
}
