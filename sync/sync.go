package sync

import (
	"sync"
)

// Bool type with concurrent access support.
type Bool struct {
	rw    sync.RWMutex
	value bool
}

// Get returns the internal value.
func (b *Bool) Get() bool {
	b.rw.RLock()
	defer b.rw.RUnlock()
	return b.value
}

// Set a value.
func (b *Bool) Set(value bool) {
	b.rw.Lock()
	defer b.rw.Unlock()
	b.value = value
}

// Int64 type with concurrent access support.
type Int64 struct {
	rw    sync.RWMutex
	value int64
}

// Get returns the internal value.
func (i *Int64) Get() int64 {
	i.rw.RLock()
	defer i.rw.RUnlock()
	return i.value
}

// Set a value.
func (i *Int64) Set(value int64) {
	i.rw.Lock()
	defer i.rw.Unlock()
	i.value = value
}

// Float64 type with concurrent access support.
type Float64 struct {
	rw    sync.RWMutex
	value float64
}

// Get returns the internal value.
func (f *Float64) Get() float64 {
	f.rw.RLock()
	defer f.rw.RUnlock()
	return f.value
}

// Set a value.
func (f *Float64) Set(value float64) {
	f.rw.Lock()
	defer f.rw.Unlock()
	f.value = value
}

// String type with concurrent access support.
type String struct {
	rw    sync.RWMutex
	value string
}

// Get returns the internal value.
func (s *String) Get() string {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return s.value
}

// Set a value.
func (s *String) Set(value string) {
	s.rw.Lock()
	defer s.rw.Unlock()
	s.value = value
}
