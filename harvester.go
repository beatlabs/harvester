package harvester

import (
	"context"
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

// Source definition.
type Source string

const (
	// SourceSeed defines a seed value.
	SourceSeed Source = "seed"
	// SourceEnv defines a value from environment variables.
	SourceEnv Source = "env"
	// SourceConsul defines a value from consul.
	SourceConsul Source = "consul"
)

// Change contains all the information that
type Change struct {
	Src     Source
	Key     string
	Value   string
	Version uint64
}

// GetValueFunc function definition for getting a value for a key from a source.
type GetValueFunc func(key string) (string, error)

// Monitorer defines a monitoring interface.
type Monitorer interface {
	Monitor(ctx context.Context)
}

// WatchItem definition.
type WatchItem struct {
	Type string
	Key  string
}

// NewKeyWatchItem creates a new key watch item for the watcher.
func NewKeyWatchItem(key string) WatchItem {
	return WatchItem{Type: "key", Key: key}
}

// NewPrefixWatchItem creates a prefix key watch item for the watcher.
func NewPrefixWatchItem(key string) WatchItem {
	return WatchItem{Type: "keyprefix", Key: key}
}

// Watcher defines methods to watch for configuration changes.
type Watcher interface {
	Watch(ww ...WatchItem) error
	Stop() error
}

// Harvester interface.
type Harvester interface {
	Harvest(ctx context.Context, cfg interface{}) error
}
