package harvester

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taxibeat/harvester/monitor/consul"
)

func TestCreate(t *testing.T) {
	ii := []consul.Item{consul.NewKeyItem("harvester1/name"), consul.NewPrefixItem("harvester")}
	type args struct {
		cfg   interface{}
		addr  string
		items []consul.Item
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "invalid cfg", args: args{cfg: "test", addr: addr, items: ii}, wantErr: true},
		{name: "invalid address", args: args{cfg: &testConfig{}, addr: "", items: ii}, wantErr: true},
		{name: "missing items", args: args{cfg: &testConfig{}, addr: addr, items: []consul.Item{}}, wantErr: true},
		{name: "success", args: args{cfg: &testConfig{}, addr: addr, items: ii}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.cfg).
				WithConsulSeed(tt.args.addr, "", "").
				WithConsulMonitor(tt.args.addr, "", "", tt.args.items...).
				Create()
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

type testConfig struct {
	Name    string  `seed:"John Doe" consul:"harvester1/name"`
	Age     int64   `seed:"18"  consul:"harvester/age"`
	Balance float64 `seed:"99.9"  consul:"harvester/balance"`
	HasJob  bool    `seed:"true"  consul:"harvester/has-job"`
}
