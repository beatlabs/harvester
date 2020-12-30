package log

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	Infof("Test %s", "logging")
	assert.Contains(t, buf.String(), "INFO: Test logging")
	buf.Reset()
	Warnf("Test %s", "logging")
	assert.Contains(t, buf.String(), "WARN: Test logging")
	buf.Reset()
	Debugf("Test %s", "logging")
	assert.Contains(t, buf.String(), "DEBUG: Test logging")
	buf.Reset()
	Errorf("Test %s", "logging")
	assert.Contains(t, buf.String(), "ERROR: Test logging")
}

func TestSetupLogging(t *testing.T) {
	stubLogf := func(string, ...interface{}) {}
	type args struct {
		infof  Func
		warnf  Func
		errorf Func
		debugf Func
	}
	tests := map[string]struct {
		args    args
		wantErr bool
	}{
		"success":       {args: args{infof: stubLogf, warnf: stubLogf, errorf: stubLogf, debugf: stubLogf}, wantErr: false},
		"missing info":  {args: args{infof: nil, warnf: stubLogf, errorf: stubLogf, debugf: stubLogf}, wantErr: true},
		"missing warn":  {args: args{infof: stubLogf, warnf: nil, errorf: stubLogf, debugf: stubLogf}, wantErr: true},
		"missing error": {args: args{infof: stubLogf, warnf: stubLogf, errorf: nil, debugf: stubLogf}, wantErr: true},
		"missing debug": {args: args{infof: stubLogf, warnf: stubLogf, errorf: stubLogf}, wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := Setup(tt.args.infof, tt.args.warnf, tt.args.errorf, tt.args.debugf)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
