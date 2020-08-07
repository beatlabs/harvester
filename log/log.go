package log

import (
	"errors"
	"io"
	"log"
)

// Func function definition.
type Func func(string, ...interface{})

var (
	debugf = func(format string, v ...interface{}) {
		log.Printf("DEBUG: "+format, v...)
	}
	infof = func(format string, v ...interface{}) {
		log.Printf("INFO: "+format, v...)
	}
	warnf = func(format string, v ...interface{}) {
		log.Printf("WARN: "+format, v...)
	}
	errorf = func(format string, v ...interface{}) {
		log.Printf("ERROR: "+format, v...)
	}
)

// Setup allows for setting up custom loggers.
func Setup(wr io.Writer, inf, waf, erf, dbf Func) error {
	if inf == nil {
		return errors.New("info log function is nil")
	}
	if waf == nil {
		return errors.New("warn log function is nil")
	}
	if erf == nil {
		return errors.New("error log function is nil")
	}
	if dbf == nil {
		return errors.New("debug log function is nil")
	}
	if wr == nil {
		return errors.New("writer is nil")
	}
	infof = inf
	warnf = waf
	errorf = erf
	debugf = dbf
	return nil
}

// Writer returns the loggers writer interface.
func Writer() io.Writer {
	return writer{}
}

// Infof provides log info capabilities.
func Infof(format string, v ...interface{}) {
	infof(format, v...)
}

// Warnf provides log warn capabilities.
func Warnf(format string, v ...interface{}) {
	warnf(format, v...)
}

// Errorf provides log error capabilities.
func Errorf(format string, v ...interface{}) {
	errorf(format, v...)
}

// Debugf provides log debug capabilities.
func Debugf(format string, v ...interface{}) {
	debugf(format, v...)
}

// writer is a wrapper around harvester's logger that implements io.Writer.
// It is available so we can keep using harvester's logger for external dependencies that require io.Writer (like hclog).
type writer struct{}

// Write log using log.Errorf() and will never returns an error.
func (writer) Write(p []byte) (n int, err error) {
	Errorf(string(p))
	return len(p), nil
}
