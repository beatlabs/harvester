package change

import "github.com/beatlabs/harvester/config"

// Change contains all the information of a change.
type Change struct {
	src     config.Source
	key     string
	value   string
	version uint64
}

// New constructor.
func New(src config.Source, key string, value string, version uint64) *Change {
	return &Change{src: src, key: key, value: value, version: version}
}

// Source of the change.
func (c Change) Source() config.Source {
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
