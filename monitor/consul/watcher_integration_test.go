// +build integration

package consul

import (
	"context"
	"testing"

	"github.com/beatlabs/harvester/change"
	"github.com/beatlabs/harvester/tests"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/suite"
)

type watcherTestSuite struct {
	suite.Suite
	consulRuntime *tests.ConsulRuntime
}

func TestWatcherTestSuite(t *testing.T) {
	suite.Run(t, new(watcherTestSuite))
}

func (s *watcherTestSuite) SetupSuite() {
	var err error
	s.consulRuntime, err = tests.NewConsulRuntime(tests.ConsulVersion)
	s.NoError(err)
	s.NoError(s.consulRuntime.StartUp())

	consul, err := s.consulRuntime.GetClient()
	s.NoError(err)
	s.NoError(setup(consul))
}

func (s *watcherTestSuite) TearDownSuite() {
	s.NoError(s.consulRuntime.TearDown())
}

func (s *watcherTestSuite) TestWatch() {
	ch := make(chan []*change.Change)
	w, err := New(s.consulRuntime.GetAddress(), "", "", 0, NewKeyItem("key1"), NewPrefixItem("prefix1"))
	s.Require().NoError(err)
	s.Require().NotNil(w)
	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()
	err = w.Watch(ctx, ch)
	s.Require().NoError(err)

	for i := 0; i < 2; i++ {
		cc := <-ch
		for _, cng := range cc {
			switch cng.Key() {
			case "prefix1/key2":
				s.Equal("2", cng.Value())
			case "prefix1/key3":
				s.Equal("3", cng.Value())
			case "key1":
				s.Equal("1", cng.Value())
			default:
				s.Fail("key invalid", cng.Key())
			}
			s.True(cng.Version() > 0)
		}
	}
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
