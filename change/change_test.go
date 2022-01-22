package change

import (
	"testing"

	"github.com/beatlabs/harvester/config"
	"github.com/stretchr/testify/assert"
)

func TestChange(t *testing.T) {
	c := New(config.SourceConsul, "key", "value", 1)
	assert.Equal(t, config.SourceConsul, c.Source())
	assert.Equal(t, "key", c.Key())
	assert.Equal(t, "value", c.Value())
	assert.Equal(t, uint64(1), c.Version())
}

func BenchmarkChangeValueSlice200Bytes(b *testing.B) {
	benchChangeSliceAndChannelSend(200, b)
}
func BenchmarkChangePointerSlice200Bytes(b *testing.B) {
	benchChangePointerSliceAndChannelSend(200, b)
}
func BenchmarkChangeValueSlice1Kb(b *testing.B) {
	benchChangeSliceAndChannelSend(1000, b)
}
func BenchmarkChangePointerSlice1Kb(b *testing.B) {
	benchChangePointerSliceAndChannelSend(1000, b)
}
func BenchmarkChangeValueSlice10Kb(b *testing.B) {
	benchChangeSliceAndChannelSend(10000, b)
}
func BenchmarkChangePointerSlice10Kb(b *testing.B) {
	benchChangePointerSliceAndChannelSend(10000, b)
}
func BenchmarkChangeValueSlice100Kb(b *testing.B) {
	benchChangeSliceAndChannelSend(100000, b)
}
func BenchmarkChangePointerSlice100Kb(b *testing.B) {
	benchChangePointerSliceAndChannelSend(100000, b)
}

func benchChangeSliceAndChannelSend(sizeInBytes uint32, b *testing.B) {
	c := make(chan []Change)
	for n := 0; n < b.N; n++ {
		changeSlice := make([]Change, 100)
		for i := 0; i < 100; i++ {
			changeSlice[i] = Change{
				src:     "",
				key:     "key",
				value:   string(make([]byte, sizeInBytes)),
				version: 0,
			}
		}
		go func() {
			c <- changeSlice
		}()
		<-c
	}
	close(c)
}

func benchChangePointerSliceAndChannelSend(sizeInBytes uint32, b *testing.B) {
	c := make(chan []*Change)
	for n := 0; n < b.N; n++ {
		changeSlice := make([]*Change, 100)
		for i := 0; i < 100; i++ {
			changeSlice[i] = &Change{
				src:     "",
				key:     "key",
				value:   string(make([]byte, sizeInBytes)),
				version: 0,
			}
		}
		go func() {
			c <- changeSlice
		}()
		<-c
	}
	close(c)
}
