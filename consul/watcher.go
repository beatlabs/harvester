package consul

import (
	"errors"
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/watch"
	"github.com/taxibeat/harvester"
)

// WatchItem definition.
type WatchItem struct {
	Type string
	Key  string
}

// NewKeyWatchItem creates a new key watch item for the watcher.
func NewKeyWatchItem(key string) WatchItem {
	return WatchItem{Type: "key", Key: key}
}

// NewPrefixWatchItem creates a prefix key watch item for the watcher.
func NewPrefixWatchItem(key string) WatchItem {
	return WatchItem{Type: "keyprefix", Key: key}
}

// Watcher of Consul changes.
type Watcher struct {
	datacenter          string
	token               string
	address             string
	ignoreInitialChange bool
	pp                  []*watch.Plan
}

// New creates a new watcher.
func New(address, datacenter, token string, ign bool) (*Watcher, error) {
	if address == "" {
		return nil, errors.New("address is empty")
	}
	return &Watcher{address: address, ignoreInitialChange: ign}, nil
}

// Watch the setup key and prefixes for changes.
func (w *Watcher) Watch(ch chan<- *harvester.Change, chErr chan<- error, ww ...WatchItem) error {
	if ch == nil {
		return errors.New("channel is nil")
	}
	if chErr == nil {
		return errors.New("error channel is nil")
	}
	if len(ww) == 0 {
		return errors.New("watch items are empty")
	}

	for _, wi := range ww {
		var pl *watch.Plan
		var err error
		switch wi.Type {
		case "key":
			pl, err = w.runKeyWatcher(ch, chErr, wi.Key)
		case "keyprefix":
			pl, err = w.runPrefixWatcher(ch, chErr, wi.Key)
		}
		if err != nil {
			return err
		}
		w.pp = append(w.pp, pl)
		go func() {
			err := pl.Run(w.address)
			if err != nil {
				chErr <- err
			}
		}()
	}
	return nil
}

// Stop the watcher.
func (w *Watcher) Stop() {
	for _, p := range w.pp {
		p.Stop()
	}
}

func (w *Watcher) runKeyWatcher(ch chan<- *harvester.Change, chErr chan<- error, key string) (*watch.Plan, error) {
	pl, err := w.getPlan("key", key)
	if err != nil {
		return nil, err
	}
	pl.Handler = func(idx uint64, data interface{}) {
		if w.ignoreInitialChange {
			w.ignoreInitialChange = false
			return
		}
		pair, ok := data.(*api.KVPair)
		if !ok {
			chErr <- fmt.Errorf("data is not kv pair: %v", data)
		}

		ch <- &harvester.Change{
			Key:     pair.Key,
			Value:   string(pair.Value),
			Version: pair.ModifyIndex,
		}
	}
	return pl, nil
}

func (w *Watcher) runPrefixWatcher(ch chan<- *harvester.Change, chErr chan<- error, key string) (*watch.Plan, error) {
	pl, err := w.getPlan("keyprefix", key)
	if err != nil {
		return nil, err
	}
	pl.Handler = func(idx uint64, data interface{}) {
		if w.ignoreInitialChange {
			w.ignoreInitialChange = false
			return
		}
		pp, ok := data.(api.KVPairs)
		if !ok {
			chErr <- fmt.Errorf("data is not kv pairs: %v", data)
		}
		for _, p := range pp {
			ch <- &harvester.Change{
				Key:     p.Key,
				Value:   string(p.Value),
				Version: p.ModifyIndex,
			}
		}
	}
	return pl, nil
}

func (w *Watcher) getPlan(tp, key string) (*watch.Plan, error) {
	params := map[string]interface{}{}
	params["datacenter"] = w.datacenter
	params["token"] = w.token
	if tp == "key" {
		params["key"] = key
	} else {
		params["prefix"] = key
	}
	params["type"] = tp
	return watch.Parse(params)
}
