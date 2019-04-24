package change

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChange(t *testing.T) {
	c := New(SourceConsul, "key", "value", 1)
	assert.Equal(t, SourceConsul, c.Source())
	assert.Equal(t, "key", c.Key())
	assert.Equal(t, "value", c.Value())
	assert.Equal(t, uint64(1), c.Version())
}
