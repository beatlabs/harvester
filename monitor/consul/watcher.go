package consul

import (
	"context"
	"errors"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/watch"
	"github.com/taxibeat/harvester/change"
	"github.com/taxibeat/harvester/config"
	"github.com/taxibeat/harvester/log"
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
	addr  string
	dc    string
	token string
	pp    []*watch.Plan
	ii    []Item
}

// New creates a new watcher.
func New(addr, dc, token string, ii ...Item) (*Watcher, error) {
	if addr == "" {
		return nil, errors.New("address is empty")
	}
	if len(ii) == 0 {
		return nil, errors.New("items are empty")
	}
	return &Watcher{addr: addr, dc: dc, token: token, ii: ii}, nil
}

// Watch key and prefixes for changes.
func (w *Watcher) Watch(ctx context.Context, ch chan<- []*change.Change, chErr chan<- error) error {
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
			pl, err = w.runKeyWatcher(i.key, ch, chErr)
		case "keyprefix":
			pl, err = w.runPrefixWatcher(i.key, ch, chErr)
		}
		if err != nil {
			return err
		}
		w.pp = append(w.pp, pl)
		go func(tp, key string) {
			err := pl.Run(w.addr)
			if err != nil {
				if chErr != nil {
					chErr <- err
				}
				log.Errorf("plan %s of type %s failed: %v", tp, key, err)
			}
		}(i.tp, i.key)
	}
	go func() {
		<-ctx.Done()
		for _, pl := range w.pp {
			pl.Stop()
		}
	}()

	return nil
}

func (w *Watcher) runKeyWatcher(key string, ch chan<- []*change.Change, chErr chan<- error) (*watch.Plan, error) {
	pl, err := w.getPlan("key", key)
	if err != nil {
		return nil, err
	}
	pl.Handler = func(idx uint64, data interface{}) {
		pair, ok := data.(*api.KVPair)
		if !ok {
			if chErr != nil {
				chErr <- err
			}
			log.Errorf("data is not kv pair: %v", data)
		}
		ch <- []*change.Change{change.New(config.SourceConsul, pair.Key, string(pair.Value), pair.ModifyIndex)}
	}
	return pl, nil
}

func (w *Watcher) runPrefixWatcher(key string, ch chan<- []*change.Change, chErr chan<- error) (*watch.Plan, error) {
	pl, err := w.getPlan("keyprefix", key)
	if err != nil {
		return nil, err
	}
	pl.Handler = func(idx uint64, data interface{}) {
		pp, ok := data.(api.KVPairs)
		if !ok {
			if chErr != nil {
				chErr <- err
			}
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
