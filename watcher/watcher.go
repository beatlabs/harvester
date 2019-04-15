package watcher

// Watcher defines methods to watch for configuration changes.
type Watcher interface {
	Watch() error
	Stop() error
}
