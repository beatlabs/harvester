package harvester

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupLogging(t *testing.T) {
	stubLogf := func(string, ...interface{}) {}
	type args struct {
		infof  LogFunc
		warnf  LogFunc
		errorf LogFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "success", args: args{infof: stubLogf, warnf: stubLogf, errorf: stubLogf}, wantErr: false},
		{name: "missing info", args: args{infof: nil, warnf: stubLogf, errorf: stubLogf}, wantErr: true},
		{name: "missing warn", args: args{infof: stubLogf, warnf: nil, errorf: stubLogf}, wantErr: true},
		{name: "missing error", args: args{infof: stubLogf, warnf: stubLogf, errorf: nil}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetupLogging(tt.args.infof, tt.args.warnf, tt.args.errorf)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
