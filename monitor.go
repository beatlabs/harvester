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

// GetValueFunc function definition for getting a value for a key from a source.
type GetValueFunc func(key string) (string, error)

// Monitor defines a monitoring interface.
type Monitor interface {
	Monitor()
}
