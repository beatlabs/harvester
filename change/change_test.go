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
	ch := make(chan []Change)
	benchChangeSliceAndChannelSend(200, ch, b)
	close(ch)
}
func BenchmarkChangePointerSlice200Bytes(b *testing.B) {
	ch := make(chan []*Change)
	benchChangePointerSliceAndChannelSend(200, ch, b)
	close(ch)
}
func BenchmarkChangeValueSlice1Kb(b *testing.B) {
	ch := make(chan []Change)
	benchChangeSliceAndChannelSend(1000, ch, b)
	close(ch)
}
func BenchmarkChangePointerSlice1Kb(b *testing.B) {
	ch := make(chan []*Change)
	benchChangePointerSliceAndChannelSend(1000, ch, b)
	close(ch)
}
func BenchmarkChangeValueSlice10Kb(b *testing.B) {
	ch := make(chan []Change)
	benchChangeSliceAndChannelSend(10000, ch, b)
	close(ch)
}
func BenchmarkChangePointerSlice10Kb(b *testing.B) {
	ch := make(chan []*Change)
	benchChangePointerSliceAndChannelSend(10000, ch, b)
	close(ch)
}
func BenchmarkChangeValueSlice100Kb(b *testing.B) {
	ch := make(chan []Change)
	benchChangeSliceAndChannelSend(100000, ch, b)
	close(ch)
}
func BenchmarkChangePointerSlice100Kb(b *testing.B) {
	ch := make(chan []*Change)
	benchChangePointerSliceAndChannelSend(100000, ch, b)
	close(ch)
}

func benchChangeSliceAndChannelSend(sizeInBytes uint32, c chan []Change, b *testing.B) {
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
}

func benchChangePointerSliceAndChannelSend(sizeInBytes uint32, c chan []*Change, b *testing.B) {
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
}
