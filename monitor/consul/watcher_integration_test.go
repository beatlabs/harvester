//go:build integration
// +build integration

package consul

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/beatlabs/harvester/change"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	addr = "127.0.0.1:8500"
)

func TestMain(m *testing.M) {
	config := api.DefaultConfig()
	config.Address = addr
	consul, err := api.NewClient(config)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	err = cleanup(consul)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	err = setup(consul)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	ret := m.Run()
	err = cleanup(consul)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	os.Exit(ret)
}

func TestWatch(t *testing.T) {
	ch := make(chan []*change.Change)
	w, err := New(addr, "", "", 0, NewKeyItemWithPrefix("key4", "consul/folder"), NewKeyItemWithPrefix("key1", ""), NewPrefixItem("prefix"))
	require.NoError(t, err)
	require.NotNil(t, w)
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()
	err = w.Watch(ctx, ch)
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		cc := <-ch
		for _, cng := range cc {
			switch cng.Key() {
			case "prefix1/key2":
				assert.Equal(t, "2", cng.Value())
			case "prefix1/key3":
				assert.Equal(t, "3", cng.Value())
			case "key1":
				assert.Equal(t, "1", cng.Value())
			case "key4":
				assert.Equal(t, "42", cng.Value())
			default:
				assert.Fail(t, "key invalid", cng.Key())
			}
			assert.Positive(t, cng.Version()) //nolint:testifylint
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
	_, err := consul.KV().Put(&api.KVPair{Key: "consul/folder/key4", Value: []byte("42")}, nil)
	if err != nil {
		return err
	}
	_, err = consul.KV().Put(&api.KVPair{Key: "key1", Value: []byte("1")}, nil)
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
