// +build integration

package consul

import (
	"testing"

	"github.com/beatlabs/harvester/tests"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/suite"
)

type getterTestSuite struct {
	suite.Suite
	consulRuntime *tests.ConsulRuntime
}

func TestGetterTestSuite(t *testing.T) {
	suite.Run(t, new(getterTestSuite))
}

func (s *getterTestSuite) SetupSuite() {
	var err error
	s.consulRuntime, err = tests.NewConsulRuntime(tests.ConsulVersion)
	s.NoError(err)
	s.NoError(s.consulRuntime.StartUp())

	consul, err := s.consulRuntime.GetClient()
	s.NoError(err)
	s.NoError(setup(consul))
}

func (s *getterTestSuite) TearDownSuite() {
	s.NoError(s.consulRuntime.TearDown())
}

func (s *getterTestSuite) TestGetter_Get() {
	one := "1"
	type args struct {
		key  string
		addr string
	}
	testList := []struct {
		name    string
		args    args
		want    *string
		wantErr bool
	}{
		{name: "success", args: args{addr: s.consulRuntime.GetAddress(), key: "get_key1"}, want: &one, wantErr: false},
		{name: "missing key", args: args{addr: s.consulRuntime.GetAddress(), key: "get_key2"}, want: nil, wantErr: false},
		{name: "wrong address", args: args{addr: "xxx", key: "get_key1"}, want: nil, wantErr: true},
	}
	for _, tt := range testList {
		s.Run(tt.name, func() {
			gtr, err := New(tt.args.addr, "", "", 0)
			s.Require().NoError(err)
			got, version, err := gtr.Get(tt.args.key)
			if tt.wantErr {
				s.Error(err)
				s.Empty(got)
			} else {
				s.NoError(err)
				s.Equal(tt.want, got)
				s.True(version >= uint64(0))
			}
		})
	}
}

func setup(consul *api.Client) error {
	_, err := consul.KV().Put(&api.KVPair{Key: "get_key1", Value: []byte("1")}, nil)
	return err
}
