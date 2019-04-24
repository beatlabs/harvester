package change

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

// Change contains all the information of a change.
type Change struct {
	src     Source
	key     string
	value   string
	version uint64
}

// New constructor.
func New(src Source, key string, value string, version uint64) *Change {
	return &Change{src: src, key: key, value: value, version: version}
}

// Source of the change.
func (c Change) Source() Source {
	return c.src
}

// Key of the change.
func (c Change) Key() string {
	return c.key
}

// Value fo the change.
func (c Change) Value() string {
	return c.value
}

// Version of the change.
func (c Change) Version() uint64 {
	return c.version
}
