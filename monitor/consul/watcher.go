// Package consul handles the monitor capabilities of harvester using ConsulLogger.
package consul

import (
	"context"
	"errors"
	"log/slog"
	"path"
	"time"

	"github.com/beatlabs/harvester/change"
	"github.com/beatlabs/harvester/config"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

// Item definition.
type Item struct {
	tp     string
	key    string
	prefix string
}

// NewKeyItem creates a new key watch item for the watcher.
func NewKeyItem(key string) Item {
	return Item{tp: "key", key: key}
}

// NewKeyItemWithPrefix creates a new key item for a given key and prefix.
func NewKeyItemWithPrefix(key, prefix string) Item {
	return Item{tp: "key", key: key, prefix: prefix}
}

// NewPrefixItem creates a prefix key watch item for the watcher.
func NewPrefixItem(key string) Item {
	return Item{tp: "keyprefix", key: key}
}

// Watcher of ConsulLogger changes.
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
	cfg := api.DefaultConfig()
	cfg.Address = addr
	if timeout > 0 {
		cfg.WaitTime = timeout
	}

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
			pl, err = w.createKeyPlanWithPrefix(i.key, i.prefix, ch)
		case "keyprefix":
			pl, err = w.createKeyPrefixPlan(i.key, ch)
		}
		if err != nil {
			return err
		}
		w.pp = append(w.pp, pl)
		go func(tp, key string) {
			err := pl.RunWithClientAndHclog(w.cl, logger)
			if err != nil {
				slog.Error("plan failed", "plan", tp, "type", key, "err", err)
			} else {
				slog.Debug("plan running", "plan", tp, "type", key)
			}
		}(i.tp, i.key)
	}
	go func() {
		<-ctx.Done()
		for _, pl := range w.pp {
			pl.Stop()
		}
		slog.Debug("all watch plans have been stopped")
	}()

	return nil
}

func (w *Watcher) createKeyPlanWithPrefix(key, prefix string, ch chan<- []*change.Change) (*watch.Plan, error) {
	pl, err := w.getPlan("key", path.Join(prefix, key))
	if err != nil {
		return nil, err
	}
	pl.Handler = func(idx uint64, data interface{}) {
		if data == nil {
			return
		}
		pair, ok := data.(*api.KVPair)
		if !ok {
			slog.Error("data is not a kv pair", "data", data)
		} else {
			ch <- []*change.Change{change.New(config.SourceConsul, key, string(pair.Value), pair.ModifyIndex)}
		}
	}
	slog.Debug("plan created", "key", key)
	return pl, nil
}

func (w *Watcher) createKeyPrefixPlan(keyPrefix string, ch chan<- []*change.Change) (*watch.Plan, error) {
	pl, err := w.getPlan("keyprefix", keyPrefix)
	if err != nil {
		return nil, err
	}
	pl.Handler = func(idx uint64, data interface{}) {
		if data == nil {
			return
		}
		pp, ok := data.(api.KVPairs)
		if !ok {
			slog.Error("data is not a kv pairs", "data", data)
		} else {
			cc := make([]*change.Change, len(pp))
			for i := 0; i < len(pp); i++ {
				cc[i] = change.New(config.SourceConsul, pp[i].Key, string(pp[i].Value), pp[i].ModifyIndex)
			}
			ch <- cc
		}
	}
	slog.Debug("plan created", "keyPrefix", keyPrefix)
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
