package consul

import (
	"log"
	"os"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taxibeat/harvester/change"
	"github.com/taxibeat/harvester/watcher"
)

const (
	addr = "127.0.0.1:8501"
)

func TestMain(m *testing.M) {
	config := api.DefaultConfig()
	config.Address = addr
	consul, err := api.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	err = cleanup(consul)
	if err != nil {
		log.Fatal(err)
	}
	err = setup(consul)
	if err != nil {
		log.Fatal(err)
	}
	ret := m.Run()
	err = cleanup(consul)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(ret)
}

func TestWatch(t *testing.T) {
	ch := make(chan []*change.Change)
	chErr := make(chan error)
	cfg, err := NewConfig(addr, "", "", ch, chErr)
	require.NoError(t, err)
	w, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, w)
	defer w.Stop()

	err = w.Watch(watcher.NewPrefixItem("prefix1"), watcher.NewKeyItem("key1"))
	require.NoError(t, err)

	for i := 0; i < 1; i++ {
		cc := <-ch
		for _, cng := range cc {
			switch cng.Key() {
			case "prefix1/key2":
				assert.Equal(t, "2", cng.Value())
			case "prefix1/key3":
				assert.Equal(t, "3", cng.Value())
			case "key1":
				assert.Equal(t, "1", cng.Value())
			default:
				assert.Fail(t, "key invalid", cng.Key())
			}
			assert.True(t, cng.Version() > 0)
		}
	}
}

func cleanup(consul *api.Client) error {
	_, err := consul.KV().Delete("key1", nil)
	if err != nil {
		return err
	}
	_, err = consul.KV().DeleteTree("prefix1", nil)
	if err != nil {
		return err
	}
	return nil
}

func setup(consul *api.Client) error {
	_, err := consul.KV().Put(&api.KVPair{Key: "key1", Value: []byte("1")}, nil)
	if err != nil {
		return err
	}
	_, err = consul.KV().Put(&api.KVPair{Key: "prefix1/key2", Value: []byte("2")}, nil)
	if err != nil {
		return err
	}
	_, err = consul.KV().Put(&api.KVPair{Key: "prefix1/key3", Value: []byte("3")}, nil)
	if err != nil {
		return err
	}
	return nil
}
