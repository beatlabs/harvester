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
	assert.Equal(t, "true", b.Print())
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
	assert.Equal(t, "10", i.Print())
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
	assert.Equal(t, "1.230000", f.Print())
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
	assert.Equal(t, "Hello", s.Print())
}

func TestSecretBool(t *testing.T) {
	var sb SecretBool
	ch := make(chan struct{})
	go func() {
		sb.Set(true)
		ch <- struct{}{}
	}()
	<-ch
	assert.True(t, sb.Get())
	assert.Equal(t, "***", sb.Print())
}

func TestSecretInt64(t *testing.T) {
	var si SecretInt64
	ch := make(chan struct{})
	go func() {
		si.Set(10)
		ch <- struct{}{}
	}()
	<-ch
	assert.Equal(t, int64(10), si.Get())
	assert.Equal(t, "***", si.Print())
}

func TestSecretFloat64(t *testing.T) {
	var sf SecretFloat64
	ch := make(chan struct{})
	go func() {
		sf.Set(1.23)
		ch <- struct{}{}
	}()
	<-ch
	assert.Equal(t, float64(1.23), sf.Get())
	assert.Equal(t, "***", sf.Print())
}

func TestSecretString(t *testing.T) {
	var ss SecretString
	ch := make(chan struct{})
	go func() {
		ss.Set("Hello")
		ch <- struct{}{}
	}()
	<-ch
	assert.Equal(t, "Hello", ss.Get())
	assert.Equal(t, "***", ss.Print())
}
