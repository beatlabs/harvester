package log

import (
	"io"
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
)

var consulLog = &consul{}

// ConsulLogger return a consul compatible logger.
func ConsulLogger() hclog.Logger {
	return consulLog
}

type consul struct{}

func (l consul) Log(level hclog.Level, msg string, args ...interface{}) {
	switch level {
	case hclog.NoLevel:
	case hclog.Trace:
		debugf(msg, args)
	case hclog.Debug:
		debugf(msg, args)
	case hclog.Info:
		infof(msg, args)
	case hclog.Warn:
		warnf(msg, args)
	case hclog.Error:
		errorf(msg, args)
	case hclog.Off:
	}
}

func (l consul) Trace(msg string, args ...interface{}) {
	debugf(msg, args)
}

func (l consul) Debug(msg string, args ...interface{}) {
	debugf(msg, args)
}

func (l consul) Info(msg string, args ...interface{}) {
	infof(msg, args)
}

func (l consul) Warn(msg string, args ...interface{}) {
	warnf(msg, args)
}

func (l consul) Error(msg string, args ...interface{}) {
	errorf(msg, args)
}

func (l consul) IsTrace() bool {
	return true
}

func (l consul) IsDebug() bool {
	return true
}

func (l consul) IsInfo() bool {
	return true
}

func (l consul) IsWarn() bool {
	return true
}

func (l consul) IsError() bool {
	return true
}

func (l consul) ImpliedArgs() []interface{} {
	return []interface{}{}
}

func (l consul) With(_ ...interface{}) hclog.Logger {
	return consulLog
}

func (l consul) Name() string {
	return "consul"
}

func (l consul) Named(_ string) hclog.Logger {
	return consulLog
}

func (l consul) ResetNamed(_ string) hclog.Logger {
	return consulLog
}

func (l consul) SetLevel(_ hclog.Level) {
}

func (l consul) StandardLogger(_ *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(os.Stderr, "", log.LstdFlags)
}

func (l consul) StandardWriter(_ *hclog.StandardLoggerOptions) io.Writer {
	return os.Stderr
}
