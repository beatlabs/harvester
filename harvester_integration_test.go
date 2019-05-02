package harvester

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taxibeat/harvester/monitor/consul"
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

func Test_harvester_Harvest(t *testing.T) {
	cfg := testConfig{}
	ii := []consul.Item{consul.NewKeyItem("harvester1/name"), consul.NewPrefixItem("harvester")}
	h, err := New(&cfg).
		WithConsulSeed(addr, "", "").
		WithConsulMonitor(addr, "", "", ii...).
		Create()
	require.NoError(t, err)

	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()
	err = h.Harvest(ctx)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, "Mr. Smith", cfg.Name)
	assert.Equal(t, int64(99), cfg.Age)
	assert.Equal(t, 111.1, cfg.Balance)
	assert.Equal(t, false, cfg.HasJob)
}

func cleanup(consul *api.Client) error {
	_, err := consul.KV().Delete("harvester1/name", nil)
	if err != nil {
		return err
	}
	_, err = consul.KV().DeleteTree("harvester", nil)
	if err != nil {
		return err
	}
	return nil
}

func setup(consul *api.Client) error {
	_, err := consul.KV().Put(&api.KVPair{Key: "harvester1/name", Value: []byte("Mr. Smith")}, nil)
	if err != nil {
		return err
	}
	_, err = consul.KV().Put(&api.KVPair{Key: "harvester/age", Value: []byte("99")}, nil)
	if err != nil {
		return err
	}
	_, err = consul.KV().Put(&api.KVPair{Key: "harvester/balance", Value: []byte("111.1")}, nil)
	if err != nil {
		return err
	}
	_, err = consul.KV().Put(&api.KVPair{Key: "harvester/has-job", Value: []byte("false")}, nil)
	if err != nil {
		return err
	}
	return nil
}
