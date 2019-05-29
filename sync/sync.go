package sync

import (
	"sync"
)

// Bool type with concurrent access support.
type Bool struct {
	mu    sync.RWMutex
	value bool
}

// Get returns the internal value.
func (b *Bool) Get() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.value
}

// Set a value.
func (b *Bool) Set(value bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.value = value
}

// Int64 type with concurrent access support.
type Int64 struct {
	mu    sync.RWMutex
	value int64
}

// Get returns the internal value.
func (i *Int64) Get() int64 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.value
}

// Set a value.
func (i *Int64) Set(value int64) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.value = value
}

// Float64 type with concurrent access support.
type Float64 struct {
	mu    sync.RWMutex
	value float64
}

// Get returns the internal value.
func (f *Float64) Get() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.value
}

// Set a value.
func (f *Float64) Set(value float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.value = value
}

// String type with concurrent access support.
type String struct {
	mu    sync.RWMutex
	value string
}

// Get returns the internal value.
func (s *String) Get() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// Set a value.
func (s *String) Set(value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.value = value
}
