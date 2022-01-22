// Package consul handles the monitor capabilities of harvester using ConsulLogger.
package consul

import (
	"context"
	"errors"
	"path"
	"sync"
	"time"

	"github.com/beatlabs/harvester/change"
	"github.com/beatlabs/harvester/config"
	"github.com/beatlabs/harvester/log"
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
func (w *Watcher) Watch(ctx context.Context) (<-chan []change.Change, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	// out channel is injected in every plan handler to communicate back the
	// changes. It can only be closed when all the plan watchers are done
	out := make(chan []change.Change, len(w.ii))
	plans := make([]*watch.Plan, 0, len(w.ii))
	for _, item := range w.ii {
		pl, err := w.createPlanForItem(item, out)
		if err != nil {
			close(out)
			return nil, err
		}
		plans = append(plans, pl)
	}

	go func() {
		dispatchWatcherPlansAndWaitCancellation(ctx, w, plans, out)
	}()

	return out, nil
}

func dispatchWatcherPlansAndWaitCancellation(ctx context.Context, w *Watcher, plans []*watch.Plan, out chan []change.Change) {
	// dispatch the plans
	wg := sync.WaitGroup{}
	wg.Add(len(plans))
	for i, pl := range plans {
		item := w.ii[i]
		go func(tp, key string, plan *watch.Plan) {
			err := plan.RunWithClientAndHclog(w.cl, log.ConsulLogger())
			if err != nil {
				log.Errorf("plan %s of type %s failed: %v", key, tp, err)
			} else {
				log.Debugf("plan %s of type %s is running", key, tp)
			}
			wg.Done()
		}(item.tp, item.key, pl)
	}
	// wait for cancellation signal
	<-ctx.Done()
	for _, pl := range plans {
		pl.Stop()
	}
	// wait for all plans to stop to close channel (plans write to chanel)
	wg.Wait()
	close(out)
	log.Debugf("all watch plans have been stopped")
}

func (w *Watcher) createPlanForItem(i Item, out chan []change.Change) (*watch.Plan, error) {
	switch i.tp {
	case "key":
		return w.createKeyPlanWithPrefix(i.key, i.prefix, out)
	case "keyprefix":
		return w.createKeyPrefixPlan(i.key, out)
	}
	return nil, errors.New("unknown item type")
}

func (w *Watcher) createKeyPlanWithPrefix(key, prefix string, ch chan<- []change.Change) (*watch.Plan, error) {
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
			log.Errorf("data is not kv pair: %v", data)
		} else {
			cg := change.New(config.SourceConsul, key, string(pair.Value), pair.ModifyIndex)
			if cg != nil {
				ch <- []change.Change{*cg}
			}
		}
	}
	log.Debugf("plan for key %s created", key)
	return pl, nil
}

func (w *Watcher) createKeyPrefixPlan(keyPrefix string, ch chan<- []change.Change) (*watch.Plan, error) {
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
			log.Errorf("data is not kv pairs: %v", data)
		} else {
			cc := make([]change.Change, 0, len(pp))
			for i := 0; i < len(pp); i++ {
				cg := change.New(config.SourceConsul, pp[i].Key, string(pp[i].Value), pp[i].ModifyIndex)
				if cg != nil {
					cc = append(cc, *cg)
				}
			}
			ch <- cc
		}
	}
	log.Debugf("plan for keyprefix %s created", keyPrefix)
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
