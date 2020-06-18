package log

import (
	"bytes"
	"io"
	"log"
	"os"
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
	assert.Equal(t, os.Stdout, Writer())
}

func TestSetupLogging(t *testing.T) {
	stubLogf := func(string, ...interface{}) {}
	type args struct {
		writer io.Writer
		infof  Func
		warnf  Func
		errorf Func
		debugf Func
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "success", args: args{writer: os.Stdin, infof: stubLogf, warnf: stubLogf, errorf: stubLogf, debugf: stubLogf}, wantErr: false},
		{name: "missing writer", args: args{writer: nil, infof: stubLogf, warnf: stubLogf, errorf: stubLogf, debugf: stubLogf}, wantErr: true},
		{name: "missing info", args: args{writer: os.Stdin, infof: nil, warnf: stubLogf, errorf: stubLogf, debugf: stubLogf}, wantErr: true},
		{name: "missing warn", args: args{writer: os.Stdin, infof: stubLogf, warnf: nil, errorf: stubLogf, debugf: stubLogf}, wantErr: true},
		{name: "missing error", args: args{writer: os.Stdin, infof: stubLogf, warnf: stubLogf, errorf: nil, debugf: stubLogf}, wantErr: true},
		{name: "missing debug", args: args{writer: os.Stdin, infof: stubLogf, warnf: stubLogf, errorf: stubLogf}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Setup(tt.args.writer, tt.args.infof, tt.args.warnf, tt.args.errorf, tt.args.debugf)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
