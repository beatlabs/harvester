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
	ignoreInitialChange bool
}

// New creates a new watcher.
func New(address string, params map[string]interface{}, ch chan<- *harvester.Change, ign bool) (*Watcher, error) {
	if address == "" {
		return nil, errors.New("address is empty")
	}
	if len(params) == 0 {
		return nil, errors.New("params are empty")
	}
	if ch == nil {
		return nil, errors.New("channel is nil")
	}
	return &Watcher{address: address, params: params, ch: ch, ignoreInitialChange: ign}, nil
}

// Watch the setup key and prefices for changes.
func (w *Watcher) Watch() error {
	pl, err := watch.Parse(w.params)
	if err != nil {
		return err
	}
	pl.Handler = func(idx uint64, data interface{}) {
		if w.ignoreInitialChange {
			w.ignoreInitialChange = false
			return
		}
		buf, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			//TODO: error handling
		}
		mp := make(map[string]interface{}, 0)
		err = json.Unmarshal(buf, mp)
		if err != nil {
			//TODO: error handling
		}
		w.ch <- &harvester.Change{
			Key:     mp["Key"].(string),
			Value:   base64.StdEncoding.EncodeToString([]byte(mp["Value"].(string))),
			Version: mp["ModifyIndex"].(int),
		}
	}
	return pl.Run(w.address)
}
