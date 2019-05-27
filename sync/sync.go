package sync

import (
	"sync"
)

// Bool type with concurrent access support.
type Bool struct {
	sync.RWMutex
	value bool
}

// Get returns the internal value.
func (b *Bool) Get() bool {
	b.RLock()
	defer b.RUnlock()
	return b.value
}

// Set a value.
func (b *Bool) Set(value bool) {
	b.Lock()
	defer b.Unlock()
	b.value = value
}

// Int64 type with concurrent access support.
type Int64 struct {
	sync.RWMutex
	value int64
}

// Get returns the internal value.
func (i *Int64) Get() int64 {
	i.RLock()
	defer i.RUnlock()
	return i.value
}

// Set a value.
func (i *Int64) Set(value int64) {
	i.Lock()
	defer i.Unlock()
	i.value = value
}

// Float64 type with concurrent access support.
type Float64 struct {
	sync.RWMutex
	value float64
}

// Get returns the internal value.
func (f *Float64) Get() float64 {
	f.RLock()
	defer f.RUnlock()
	return f.value
}

// Set a value.
func (f *Float64) Set(value float64) {
	f.Lock()
	defer f.Unlock()
	f.value = value
}

// String type with concurrent access support.
type String struct {
	sync.RWMutex
	value string
}

// Get returns the internal value.
func (s *String) Get() string {
	s.RLock()
	defer s.RUnlock()
	return s.value
}

// Set a value.
func (s *String) Set(value string) {
	s.Lock()
	defer s.Unlock()
	s.value = value
}
