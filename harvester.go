package harvester

import (
	"errors"
	"log"
)

// LogFunc function definition.
type LogFunc func(string, ...interface{})

var (
	logInfof = func(format string, v ...interface{}) {
		log.Printf("INFO: "+format, v...)
	}
	logWarnf = func(format string, v ...interface{}) {
		log.Printf("WARN: "+format, v...)
	}
	logErrorf = func(format string, v ...interface{}) {
		log.Printf("ERROR: "+format, v...)
	}
)

// SetupLogging allows for setting up custom loggers.
func SetupLogging(infof, warnf, errorf LogFunc) error {
	if infof == nil {
		return errors.New("info log function is nil")
	}
	if warnf == nil {
		return errors.New("warn log function is nil")
	}
	if errorf == nil {
		return errors.New("error log function is nil")
	}
	logInfof = infof
	logWarnf = warnf
	logErrorf = errorf
	return nil
}

// LogInfof provides log info capabilities.
func LogInfof(format string, v ...interface{}) {
	logInfof(format, v...)
}

// LogWarnf provides log warn capabilities.
func LogWarnf(format string, v ...interface{}) {
	logWarnf(format, v...)
}

// LogErrorf provides log error capabilities.
func LogErrorf(format string, v ...interface{}) {
	logErrorf(format, v...)
}

// Harvester interface.
type Harvester interface {
	Harvest(cfg interface{}) error
}
