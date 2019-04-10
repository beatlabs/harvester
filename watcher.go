package harvester

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

// Watcher defines methods to watch for configuration changes.
type Watcher interface {
	Watch() error
	Stop() error
}
