// +build integration

package harvester

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/beatlabs/harvester/monitor/consul"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, "Mr. Smith", cfg.Name.Get())
	assert.Equal(t, int64(99), cfg.Age.Get())
	assert.Equal(t, 111.1, cfg.Balance.Get())
	assert.Equal(t, false, cfg.HasJob.Get())
	_, err = csl.Put(&api.KVPair{Key: "harvester1/name", Value: []byte("Mr. Anderson")}, nil)
	require.NoError(t, err)
	time.Sleep(1000 * time.Millisecond)
	assert.Equal(t, "Mr. Anderson", cfg.Name.Get())
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
