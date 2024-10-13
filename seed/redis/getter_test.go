package redis

import (
	"testing"

	"github.com/go-redis/redis/v8"
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
