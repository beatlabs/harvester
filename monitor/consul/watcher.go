package consul

import (
	"context"
	"errors"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/watch"
	"github.com/taxibeat/harvester/change"
	"github.com/taxibeat/harvester/config"
	"github.com/taxibeat/harvester/log"
	"github.com/taxibeat/harvester/monitor"
)

// Watcher of Consul changes.
type Watcher struct {
	addr  string
	dc    string
	token string
	pp    []*watch.Plan
}

// New creates a new watcher.
func New(addr, dc, token string) (*Watcher, error) {
	if addr == "" {
		return nil, errors.New("address is empty")
	}
	return &Watcher{addr: addr, dc: dc, token: token}, nil
}

// Watch key and prefixes for changes.
func (w *Watcher) Watch(ctx context.Context, ii []monitor.Item, ch chan<- []*change.Change) error {
	if len(ii) == 0 {
		return errors.New("items are empty")
	}
	if ch == nil {
		return errors.New("change channel is nil")
	}
	for _, i := range ii {
		var pl *watch.Plan
		var err error
		switch i.Type {
		case "key":
			pl, err = w.runKeyWatcher(i.Key, ch)
		case "keyprefix":
			pl, err = w.runPrefixWatcher(i.Key, ch)
		}
		if err != nil {
			return err
		}
		w.pp = append(w.pp, pl)
		go func() {
			err := pl.Run(w.addr)
			if err != nil {
				//TODO: error handling in monitor...
				log.Errorf("", err)
			}
		}()
	}
	go func() {
		<-ctx.Done()
		for _, pl := range w.pp {
			pl.Stop()
		}
	}()

	return nil
}

func (w *Watcher) runKeyWatcher(key string, ch chan<- []*change.Change) (*watch.Plan, error) {
	pl, err := w.getPlan("key", key)
	if err != nil {
		return nil, err
	}
	pl.Handler = func(idx uint64, data interface{}) {
		pair, ok := data.(*api.KVPair)
		if !ok {
			log.Errorf("data is not kv pair: %v", data)
		}
		ch <- []*change.Change{change.New(config.SourceConsul, pair.Key, string(pair.Value), pair.ModifyIndex)}
	}
	return pl, nil
}

func (w *Watcher) runPrefixWatcher(key string, ch chan<- []*change.Change) (*watch.Plan, error) {
	pl, err := w.getPlan("keyprefix", key)
	if err != nil {
		return nil, err
	}
	pl.Handler = func(idx uint64, data interface{}) {
		pp, ok := data.(api.KVPairs)
		if !ok {
			log.Errorf("data is not kv pairs: %v", data)
		}
		cc := make([]*change.Change, len(pp))
		for i := 0; i < len(pp); i++ {
			cc[i] = change.New(config.SourceConsul, pp[i].Key, string(pp[i].Value), pp[i].ModifyIndex)
		}
		ch <- cc
	}
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
