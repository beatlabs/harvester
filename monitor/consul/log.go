package consul

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/hashicorp/go-hclog"
)

var logger = &clog{}

type clog struct{}

var _ hclog.Logger = logger

func (l clog) Log(level hclog.Level, msg string, args ...interface{}) {
	switch level {
	case hclog.NoLevel:
	case hclog.Trace:
		l.Trace(msg, args...)
	case hclog.Debug:
		l.Debug(msg, args...)
	case hclog.Info:
		l.Info(msg, args...)
	case hclog.Warn:
		l.Warn(msg, args...)
	case hclog.Error:
		l.Error(msg, args...)

	case hclog.Off:
	}
}

func (l clog) Trace(msg string, args ...interface{}) {
	slog.Debug(fmt.Sprintf(msg, args...))
}

func (l clog) Debug(msg string, args ...interface{}) {
	slog.Debug(fmt.Sprintf(msg, args...))
}

func (l clog) Info(msg string, args ...interface{}) {
	slog.Info(fmt.Sprintf(msg, args...))
}

func (l clog) Warn(msg string, args ...interface{}) {
	slog.Warn(fmt.Sprintf(msg, args...))
}

func (l clog) Error(msg string, args ...interface{}) {
	slog.Error(fmt.Sprintf(msg, args...))
}

func (l clog) IsTrace() bool {
	return true
}

func (l clog) IsDebug() bool {
	return true
}

func (l clog) IsInfo() bool {
	return true
}

func (l clog) IsWarn() bool {
	return true
}

func (l clog) IsError() bool {
	return true
}

func (l clog) ImpliedArgs() []interface{} {
	return []interface{}{}
}

func (l clog) With(_ ...interface{}) hclog.Logger {
	return logger
}

func (l clog) Name() string {
	return "consul"
}

func (l clog) Named(_ string) hclog.Logger {
	return logger
}

func (l clog) ResetNamed(_ string) hclog.Logger {
	return logger
}

func (l clog) SetLevel(_ hclog.Level) {
}

func (l clog) StandardLogger(_ *hclog.StandardLoggerOptions) *log.Logger {
	return slog.NewLogLogger(slog.Default().Handler(), slog.LevelInfo)
}

func (l clog) StandardWriter(_ *hclog.StandardLoggerOptions) io.Writer {
	return os.Stderr
}

func (l clog) GetLevel() hclog.Level {
	return hclog.NoLevel
}
