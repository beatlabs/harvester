package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	type args struct {
		client redis.UniversalClient
	}
	tests := map[string]struct {
		args        args
		expectedErr string
	}{
		"success":        {args: args{client: &redis.Client{}}},
		"missing client": {args: args{client: nil}, expectedErr: "client is nil"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := New(tt.args.client)
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

type stubRedisClient struct {
	redis.UniversalClient
	result string
	err    error
}

func (s *stubRedisClient) Get(_ context.Context, _ string) *redis.StringCmd {
	return redis.NewStringResult(s.result, s.err)
}

func TestGetter_Get_Unit(t *testing.T) {
	sentinel := errors.New("connection refused")
	tests := map[string]struct {
		stub    *stubRedisClient
		wantVal *string
		wantErr bool
	}{
		"key found": {
			stub:    &stubRedisClient{result: "hello", err: nil},
			wantVal: strPtr("hello"),
		},
		"key not found (redis.Nil)": {
			stub:    &stubRedisClient{result: "", err: redis.Nil},
			wantVal: nil,
		},
		"connection error": {
			stub:    &stubRedisClient{result: "", err: sentinel},
			wantVal: nil,
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			g := &Getter{client: tt.stub}
			val, version, err := g.Get("any-key")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantVal, val)
			assert.Equal(t, uint64(0), version)
		})
	}
}

func strPtr(s string) *string { return &s }
