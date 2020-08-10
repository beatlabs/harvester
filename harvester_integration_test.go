// +build integration

package harvester

import (
	"bytes"
	"context"
	"log"
	"testing"
	"time"

	"github.com/beatlabs/harvester/sync"
	"github.com/beatlabs/harvester/tests"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	secretLog = []string{
		`INFO: automatically monitoring consul key "harvester1/name"`,
		`INFO: automatically monitoring consul key "harvester/age"`,
		`INFO: automatically monitoring consul key "harvester/balance"`,
		`INFO: automatically monitoring consul key "harvester/has-job"`,
		`INFO: automatically monitoring consul key "harvester/foo/bar"`,
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
		`INFO: field "FooBar" updated with value "123", version: `,
		`INFO: plan for key harvester1/name created`,
		`INFO: plan for key harvester/age created`,
		`INFO: plan for key harvester/balance created`,
		`INFO: plan for key harvester/has-job created`,
		`INFO: plan for key harvester/foo/bar created`,
	}
)

type harvesterIntegrationSuite struct {
	suite.Suite
	consulRuntime *tests.ConsulRuntime
	consulKV      *api.KV
}

func TestHarvesterIntegrationSuite(t *testing.T) {
	suite.Run(t, new(harvesterIntegrationSuite))
}

func (s *harvesterIntegrationSuite) SetupSuite() {
	var err error
	s.consulRuntime, err = tests.NewConsulRuntime(tests.ConsulVersion)
	s.Require().NoError(err)
	s.Require().NoError(s.consulRuntime.StartUp())

	client, err := s.consulRuntime.GetClient()
	s.Require().NoError(err)
	s.consulKV = client.KV()
	s.setup()
}

func (s *harvesterIntegrationSuite) TearDownSuite() {
	s.NoError(s.consulRuntime.TearDown())
}

type testConfigWithSecret struct {
	Name    sync.Secret  `seed:"John Doe" consul:"harvester1/name"`
	Age     sync.Int64   `seed:"18" consul:"harvester/age"`
	Balance sync.Float64 `seed:"99.9" consul:"harvester/balance"`
	HasJob  sync.Bool    `seed:"true" consul:"harvester/has-job"`
	Foo     fooStruct
}

type fooStruct struct {
	Bar sync.Int64 `seed:"123" consul:"harvester/foo/bar"`
}

func (s *harvesterIntegrationSuite) Test_harvester_Harvest() {
	buf := bytes.NewBuffer(make([]byte, 0))
	log.SetOutput(buf)
	cfg := testConfigWithSecret{}
	h, err := New(&cfg).
		WithConsulSeed(s.consulRuntime.GetAddress(), "", "", 0).
		WithConsulMonitor(s.consulRuntime.GetAddress(), "", "", 0).
		Create()
	s.Require().NoError(err)

	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()
	err = h.Harvest(ctx)

	testLogOutput(buf, s.T())
	s.NoError(err)
	s.Equal("Mr. Smith", cfg.Name.Get())
	s.Equal(int64(99), cfg.Age.Get())
	s.Equal(111.1, cfg.Balance.Get())
	s.Equal(false, cfg.HasJob.Get())
	s.Equal(int64(123), cfg.Foo.Bar.Get())
	_, err = s.consulKV.Put(&api.KVPair{Key: "harvester1/name", Value: []byte("Mr. Anderson")}, nil)
	s.Require().NoError(err)
	time.Sleep(1000 * time.Millisecond)
	s.Equal("Mr. Anderson", cfg.Name.Get())

	_, err = s.consulKV.Put(&api.KVPair{Key: "harvester/foo/bar", Value: []byte("42")}, nil)
	s.Require().NoError(err)
	time.Sleep(1000 * time.Millisecond)
	s.Equal(int64(42), cfg.Foo.Bar.Get())
}

func testLogOutput(buf *bytes.Buffer, t *testing.T) {
	log := buf.String()
	for _, logLine := range secretLog {
		assert.Contains(t, log, logLine)
	}
}

func (s *harvesterIntegrationSuite) setup() {
	_, err := s.consulKV.Put(&api.KVPair{Key: "harvester1/name", Value: []byte("Mr. Smith")}, nil)
	s.NoError(err)
	_, err = s.consulKV.Put(&api.KVPair{Key: "harvester/age", Value: []byte("99")}, nil)
	s.NoError(err)
	_, err = s.consulKV.Put(&api.KVPair{Key: "harvester/balance", Value: []byte("111.1")}, nil)
	s.NoError(err)
	_, err = s.consulKV.Put(&api.KVPair{Key: "harvester/has-job", Value: []byte("false")}, nil)
	s.NoError(err)
}
