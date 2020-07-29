// +build integration

package harvester

import (
	"bytes"
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/beatlabs/harvester/sync"

	"github.com/beatlabs/harvester/monitor/consul"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	csl       *api.KV
	secretLog = []string{
		`INFO: field "Name" updated with value "***", version: `,
		`INFO: seed value *** applied on field Name`,
		`INFO: field "Name" updated with value "***", version: `,
		`INFO: consul value *** applied on field Name`,
		`INFO: field "Age" updated with value "18", version: `,
		`INFO: seed value 18 applied on field Age`,
		`INFO: field "Age" updated with value "99", version: `,
		`INFO: consul value 99 applied on field Age`,
		`INFO: field "Balance" updated with value "99.900000", version: `,
		`INFO: seed value 99.900000 applied on field Balance`,
		`INFO: field "Balance" updated with value "111.100000", version: `,
		`INFO: consul value 111.100000 applied on field Balance`,
		`INFO: field "HasJob" updated with value "true", version: `,
		`INFO: seed value true applied on field HasJob`,
		`INFO: field "HasJob" updated with value "false", version: `,
		`INFO: consul value false applied on field HasJob`,
		`INFO: plan for key harvester1/name created`,
		`INFO: plan for keyprefix harvester created`,
	}
)

type testConfigWithSecret struct {
	Name    sync.Secret  `seed:"John Doe" consul:"harvester1/name"`
	Age     sync.Int64   `seed:"18"  consul:"harvester/age"`
	Balance sync.Float64 `seed:"99.9"  consul:"harvester/balance"`
	HasJob  sync.Bool    `seed:"true"  consul:"harvester/has-job"`
}

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
	buf := bytes.NewBuffer(make([]byte, 0))
	log.SetOutput(buf)
	cfg := testConfigWithSecret{}
	ii := []consul.Item{consul.NewKeyItem("harvester1/name"), consul.NewPrefixItem("harvester")}
	h, err := New(&cfg).
		WithConsulSeed(addr, "", "", 0).
		WithConsulMonitor(addr, "", "", 0, ii...).
		Create()
	require.NoError(t, err)

	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()
	err = h.Harvest(ctx)
	testLogOutput(buf, t)
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

func testLogOutput(buf *bytes.Buffer, t *testing.T) {
	log := buf.String()
	for _, logLine := range secretLog {
		assert.Contains(t, log, logLine)
	}
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
