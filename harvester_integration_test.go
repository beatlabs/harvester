// +build integration

package harvester

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taxibeat/harvester/monitor/consul"
)

var (
	csl *api.KV
)

func TestMain(m *testing.M) {
	config := api.DefaultConfig()
	config.Address = addr
	c, err := api.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	csl = c.KV()
	err = cleanup()
	if err != nil {
		log.Fatal(err)
	}
	err = setup()
	if err != nil {
		log.Fatal(err)
	}
	ret := m.Run()
	err = cleanup()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(ret)
}

func Test_harvester_Harvest(t *testing.T) {
	cfg := testConfig{}
	ii := []consul.Item{consul.NewKeyItem("harvester1/name"), consul.NewPrefixItem("harvester")}
	h, err := New(&cfg).
		WithConsulSeed(addr, "", "", 0).
		WithConsulMonitor(addr, "", "", 0, ii...).
		Create()
	require.NoError(t, err)

	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()
	err = h.Harvest(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "Mr. Smith", cfg.Name)
	assert.Equal(t, int64(99), cfg.Age)
	assert.Equal(t, 111.1, cfg.Balance)
	assert.Equal(t, false, cfg.HasJob)
}

func cleanup() error {
	_, err := csl.Delete("harvester1/name", nil)
	if err != nil {
		return err
	}
	_, err = csl.DeleteTree("harvester", nil)
	if err != nil {
		return err
	}
	return nil
}

func setup() error {
	_, err := csl.Put(&api.KVPair{Key: "harvester1/name", Value: []byte("Mr. Smith")}, nil)
	if err != nil {
		return err
	}
	_, err = csl.Put(&api.KVPair{Key: "harvester/age", Value: []byte("99")}, nil)
	if err != nil {
		return err
	}
	_, err = csl.Put(&api.KVPair{Key: "harvester/balance", Value: []byte("111.1")}, nil)
	if err != nil {
		return err
	}
	_, err = csl.Put(&api.KVPair{Key: "harvester/has-job", Value: []byte("false")}, nil)
	if err != nil {
		return err
	}
	return nil
}
