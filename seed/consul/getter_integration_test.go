//go:build integration
// +build integration

package consul

import (
	"log"
	"os"
	"testing"

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

func TestGetter_Get(t *testing.T) {
	one := "1"
	type args struct {
		key  string
		addr string
	}
	tests := map[string]struct {
		args    args
		want    *string
		wantErr bool
	}{
		"success":       {args: args{addr: addr, key: "get_key1"}, want: &one, wantErr: false},
		"missing key":   {args: args{addr: addr, key: "get_key2"}, want: nil, wantErr: false},
		"wrong address": {args: args{addr: "xxx", key: "get_key1"}, want: nil, wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gtr, err := New(tt.args.addr, "", "", 0)
			require.NoError(t, err)
			got, version, err := gtr.Get(tt.args.key)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.True(t, version >= uint64(0))
			}
		})
	}
}

func cleanup(consul *api.Client) error {
	_, err := consul.KV().Delete("get_key1", nil)
	if err != nil {
		return err
	}
	return nil
}

func setup(consul *api.Client) error {
	_, err := consul.KV().Put(&api.KVPair{Key: "get_key1", Value: []byte("1")}, nil)
	if err != nil {
		return err
	}
	return nil
}
