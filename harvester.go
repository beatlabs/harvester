package harvester

import (
	"errors"
	"log"
)

// LogFunc function definition.
type LogFunc func(string, ...interface{})

var (
	logWarnf = func(format string, v ...interface{}) {
		log.Printf("WARN: "+format, v...)
	}
	logErrorf = func(format string, v ...interface{}) {
		log.Printf("ERROR: "+format, v...)
	}
)

// SetupLogging allows for setting up custom loggers.
func SetupLogging(warnf, errorf LogFunc) error {
	if warnf == nil {
		return errors.New("warn log function is nil")
	}
	if errorf == nil {
		return errors.New("error log function is nil")
	}
	logWarnf = warnf
	logErrorf = errorf
	return nil
}
