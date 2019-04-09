package consul

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/hashicorp/consul/watch"
	"github.com/taxibeat/harvester"
)

// Watcher of Consul changes.
type Watcher struct {
	params              map[string]interface{}
	address             string
	ch                  chan<- *harvester.Change
	chErr               chan<- error
	ignoreInitialChange bool
	pl                  *watch.Plan
}

// New creates a new watcher.
func New(address string, params map[string]interface{}, ch chan<- *harvester.Change,
	chErr chan<- error, ign bool) (*Watcher, error) {
	if address == "" {
		return nil, errors.New("address is empty")
	}
	if len(params) == 0 {
		return nil, errors.New("params are empty")
	}
	if ch == nil {
		return nil, errors.New("channel is nil")
	}
	if chErr == nil {
		return nil, errors.New("error channel is nil")
	}
	return &Watcher{address: address, params: params, ch: ch, chErr: chErr, ignoreInitialChange: ign}, nil
}

// Watch the setup key and prefixes for changes.
func (w *Watcher) Watch() error {
	var err error
	w.pl, err = watch.Parse(w.params)
	if err != nil {
		return err
	}
	w.pl.Handler = func(idx uint64, data interface{}) {
		if w.ignoreInitialChange {
			w.ignoreInitialChange = false
			return
		}
		buf, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			w.chErr <- err
			return
		}
		mp := make(map[string]interface{}, 0)
		err = json.Unmarshal(buf, &mp)
		if err != nil {
			w.chErr <- err
			return
		}
		w.ch <- &harvester.Change{
			Key:     mp["Key"].(string),
			Value:   base64.StdEncoding.EncodeToString([]byte(mp["Value"].(string))),
			Version: mp["ModifyIndex"].(int),
		}
	}
	return w.pl.Run(w.address)
}

// Stop the watcher.
func (w *Watcher) Stop() error {
	return w.Stop()
}
