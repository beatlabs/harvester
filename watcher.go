package harvester

// Change contains all the information that
type Change struct {
	Key   string
	Value string
}

// Watcher defines methods to watch for configuration changes.
type Watcher interface {
	Watch() error
	Stop() error
}
