package consul

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taxibeat/harvester"
)

func TestNew(t *testing.T) {
	type args struct {
		address string
		params  map[string]interface{}
		ch      chan<- *harvester.Change
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				address: "addr",
				params:  map[string]interface{}{"key": "value"},
				ch:      make(chan *harvester.Change),
			},
			wantErr: false,
		},
		{
			name: "missing address",
			args: args{
				address: "",
				params:  map[string]interface{}{"key": "value"},
				ch:      make(chan *harvester.Change),
			},
			wantErr: true,
		},
		{
			name: "missing params",
			args: args{
				address: "addr",
				params:  nil,
				ch:      make(chan *harvester.Change),
			},
			wantErr: true,
		},
		{
			name: "missing channel",
			args: args{
				address: "addr",
				params:  map[string]interface{}{"key": "value"},
				ch:      nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.address, tt.args.params, tt.args.ch, false)
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
