package consul

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	type args struct {
		addr    string
		timeout time.Duration
	}
	tests := map[string]struct {
		args    args
		wantErr bool
	}{
		"success":                  {args: args{addr: "addr", timeout: 0}, wantErr: false},
		"success explicit timeout": {args: args{addr: "addr", timeout: 30 * time.Second}, wantErr: false},
		"missing address":          {args: args{addr: ""}, wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := New(tt.args.addr, "dc", "token", 0)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

func TestGetter_GetContext_NilContext(t *testing.T) {
	g := &Getter{}
	val, version, err := g.GetContext(nil, "any-key") //nolint:staticcheck // asserting the nil-context guard
	require.EqualError(t, err, "context is nil")
	assert.Nil(t, val)
	assert.Equal(t, uint64(0), version)
}

func TestGetter_Get_Unreachable(t *testing.T) {
	g, err := New("127.0.0.1:1", "dc", "token", 0)
	require.NoError(t, err)

	val, version, err := g.Get("any-key")
	require.Error(t, err)
	assert.Nil(t, val)
	assert.Equal(t, uint64(0), version)
}
