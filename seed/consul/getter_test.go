package consul

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}
