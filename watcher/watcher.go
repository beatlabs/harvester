package watcher

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

// Item definition.
type Item struct {
	Type string
	Key  string
}

// NewKeyItem creates a new key watch item for the watcher.
func NewKeyItem(key string) Item {
	return Item{Type: "key", Key: key}
}

// NewPrefixItem creates a prefix key watch item for the watcher.
func NewPrefixItem(key string) Item {
	return Item{Type: "keyprefix", Key: key}
}

// Watcher defines methods to watch for configuration changes.
type Watcher interface {
	Watch(ww ...Item) error
	Stop() error
}
