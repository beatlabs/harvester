package consul

import (
	"encoding/base64"
	"encoding/json"
	"errors"

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
	return WatchItem{Type: "prefix", Key: key}
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

	for _, wi := range ww {
		var pl *watch.Plan
		var err error
		switch wi.Type {
		case "key":
			pl, err = w.runKeyWatcher(ch, chErr, wi.Key)
		case "prefix":
			pl, err = w.runPrefixWatcher(ch, chErr, wi.Key)
		}
		if err != nil {
			return err
		}
		w.pp = append(w.pp, pl)
	}
	return nil
}

// Stop the watcher.
func (w *Watcher) Stop() error {

	return w.Stop()
}

func (w *Watcher) runKeyWatcher(ch chan<- *harvester.Change, chErr chan<- error, key string) (*watch.Plan, error) {
	params := map[string]interface{}{}
	params["datacenter"] = w.datacenter
	params["token"] = w.token
	params["key"] = key
	params["type"] = "key"
	pl, err := watch.Parse(params)
	if err != nil {
		return nil, err
	}
	pl.Handler = func(idx uint64, data interface{}) {
		if w.ignoreInitialChange {
			w.ignoreInitialChange = false
			return
		}
		buf, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			chErr <- err
			return
		}
		mp := make(map[string]interface{}, 0)
		err = json.Unmarshal(buf, &mp)
		if err != nil {
			chErr <- err
			return
		}
		ch <- &harvester.Change{
			Key:     mp["Key"].(string),
			Value:   base64.StdEncoding.EncodeToString([]byte(mp["Value"].(string))),
			Version: mp["ModifyIndex"].(int),
		}
	}
	err = pl.Run(w.address)
	if err != nil {
		return nil, err
	}
	return pl, nil
}

func (w *Watcher) runPrefixWatcher(ch chan<- *harvester.Change, chErr chan<- error, key string) (*watch.Plan, error) {
	params := map[string]interface{}{}
	params["datacenter"] = w.datacenter
	params["token"] = w.token
	params["key"] = key
	params["type"] = "prefix"
	pl, err := watch.Parse(params)
	if err != nil {
		return nil, err
	}
	pl.Handler = func(idx uint64, data interface{}) {
		if w.ignoreInitialChange {
			w.ignoreInitialChange = false
			return
		}
		buf, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			chErr <- err
			return
		}
		//TODO: this might be a array...
		mp := make(map[string]interface{}, 0)
		err = json.Unmarshal(buf, &mp)
		if err != nil {
			chErr <- err
			return
		}
		ch <- &harvester.Change{
			Key:     mp["Key"].(string),
			Value:   base64.StdEncoding.EncodeToString([]byte(mp["Value"].(string))),
			Version: mp["ModifyIndex"].(int),
		}
	}
	err = pl.Run(w.address)
	if err != nil {
		return nil, err
	}
	return pl, nil
}
