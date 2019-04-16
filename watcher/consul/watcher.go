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

// Config for configuring the watcher.
type Config struct {
	Address    string
	Datacenter string
	Token      string
	ch         chan<- []*harvester.Change
	chErr      chan<- error
}

// NewConfig constructor.
func NewConfig(addr, dc, token string, ch chan<- []*harvester.Change, chErr chan<- error) (*Config, error) {
	if addr == "" {
		return nil, errors.New("address is empty")
	}
	if ch == nil {
		return nil, errors.New("channel is nil")
	}
	if chErr == nil {
		return nil, errors.New("error channel is nil")
	}
	return &Config{
		Address:    addr,
		Datacenter: dc,
		Token:      token,
		ch:         ch,
		chErr:      chErr,
	}, nil
}

// Watcher of Consul changes.
type Watcher struct {
	cfg *Config
	pp  []*watch.Plan
}

// New creates a new watcher.
func New(cfg *Config) (*Watcher, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	return &Watcher{cfg: cfg}, nil
}

// Watch the setup key and prefixes for changes.
func (w *Watcher) Watch(ww ...WatchItem) error {
	if len(ww) == 0 {
		return errors.New("watch items are empty")
	}

	for _, wi := range ww {
		var pl *watch.Plan
		var err error
		switch wi.Type {
		case "key":
			pl, err = w.runKeyWatcher(wi.Key)
		case "keyprefix":
			pl, err = w.runPrefixWatcher(wi.Key)
		}
		if err != nil {
			return err
		}
		w.pp = append(w.pp, pl)
		go func() {
			err := pl.Run(w.cfg.Address)
			if err != nil {
				w.cfg.chErr <- err
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

func (w *Watcher) runKeyWatcher(key string) (*watch.Plan, error) {
	pl, err := w.getPlan("key", key)
	if err != nil {
		return nil, err
	}
	pl.Handler = func(idx uint64, data interface{}) {
		pair, ok := data.(*api.KVPair)
		if !ok {
			w.cfg.chErr <- fmt.Errorf("data is not kv pair: %v", data)
		}

		w.cfg.ch <- []*harvester.Change{&harvester.Change{
			Src:     harvester.SourceConsul,
			Key:     pair.Key,
			Value:   string(pair.Value),
			Version: pair.ModifyIndex,
		}}
	}
	return pl, nil
}

func (w *Watcher) runPrefixWatcher(key string) (*watch.Plan, error) {
	pl, err := w.getPlan("keyprefix", key)
	if err != nil {
		return nil, err
	}
	pl.Handler = func(idx uint64, data interface{}) {
		pp, ok := data.(api.KVPairs)
		if !ok {
			w.cfg.chErr <- fmt.Errorf("data is not kv pairs: %v", data)
		}
		cc := make([]*harvester.Change, len(pp))
		for _, p := range pp {
			cc = append(cc, &harvester.Change{
				Src:     harvester.SourceConsul,
				Key:     p.Key,
				Value:   string(p.Value),
				Version: p.ModifyIndex,
			})
		}
		w.cfg.ch <- cc
	}
	return pl, nil
}

func (w *Watcher) getPlan(tp, key string) (*watch.Plan, error) {
	params := map[string]interface{}{}
	params["datacenter"] = w.cfg.Datacenter
	params["token"] = w.cfg.Token
	if tp == "key" {
		params["key"] = key
	} else {
		params["prefix"] = key
	}
	params["type"] = tp
	return watch.Parse(params)
}
