package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	var b Bool
	ch := make(chan struct{})
	go func() {
		b.Set(true)
		ch <- struct{}{}
	}()
	<-ch
	assert.True(t, b.Get())
	assert.Equal(t, "true", b.String())
}

func TestBool_SetString(t *testing.T) {
	var b Bool
	assert.Error(t, b.SetString("wrong"))
	assert.NoError(t, b.SetString("true"))
	assert.True(t, b.Get())
}

func TestInt64(t *testing.T) {
	var i Int64
	ch := make(chan struct{})
	go func() {
		i.Set(10)
		ch <- struct{}{}
	}()
	<-ch
	assert.Equal(t, int64(10), i.Get())
	assert.Equal(t, "10", i.String())
}

func TestInt64_SetString(t *testing.T) {
	var i Int64
	assert.Error(t, i.SetString("wrong"))
	assert.NoError(t, i.SetString("10"))
	assert.Equal(t, int64(10), i.Get())
}

func TestFloat64(t *testing.T) {
	var f Float64
	ch := make(chan struct{})
	go func() {
		f.Set(1.23)
		ch <- struct{}{}
	}()
	<-ch
	assert.Equal(t, float64(1.23), f.Get())
	assert.Equal(t, "1.230000", f.String())
}

func TestFloat64_SetString(t *testing.T) {
	var f Float64
	assert.Error(t, f.SetString("wrong"))
	assert.NoError(t, f.SetString("1.230000"))
	assert.Equal(t, 1.23, f.Get())
}

func TestString(t *testing.T) {
	var s String
	ch := make(chan struct{})
	go func() {
		s.Set("Hello")
		ch <- struct{}{}
	}()
	<-ch
	assert.Equal(t, "Hello", s.Get())
	assert.Equal(t, "Hello", s.String())
}

func TestString_SetString(t *testing.T) {
	var s String
	assert.NoError(t, s.SetString("foo"))
	assert.Equal(t, "foo", s.Get())
}

func TestSecret(t *testing.T) {
	var s Secret
	ch := make(chan struct{})
	go func() {
		s.Set("Hello")
		ch <- struct{}{}
	}()
	<-ch
	assert.Equal(t, "Hello", s.Get())
	assert.Equal(t, "***", s.String())
}

func TestSecret_SetString(t *testing.T) {
	var s Secret
	assert.NoError(t, s.SetString("foo"))
	assert.Equal(t, "foo", s.Get())
}
