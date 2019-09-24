package sync

import (
	"fmt"
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

// Print returns string representation of value.
func (b *Bool) Print() string {
	if b.value {
		return "true"
	}
	return "false"
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

// Print returns string representation of value.
func (i *Int64) Print() string {
	return fmt.Sprintf("%d", i.value)
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

// Print returns string representation of value.
func (f *Float64) Print() string {
	return fmt.Sprintf("%f", f.value)
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

// Print returns string representation of value.
func (s *String) Print() string {
	return s.value
}

// SecretBool bool type for secrets with concurrent access support.
type SecretBool struct{ Bool }

// Print returns obfuscated string representation of value.
func (s *SecretBool) Print() string {
	return "***"
}

// SecretInt64 int64 type for secrets with concurrent access support.
type SecretInt64 struct{ Int64 }

// Print returns obfuscated string representation of value.
func (s *SecretInt64) Print() string {
	return "***"
}

// SecretFloat64 float64 type for secrets with concurrent access support.
type SecretFloat64 struct{ Float64 }

// Print returns obfuscated string representation of value.
func (s *SecretFloat64) Print() string {
	return "***"
}

// SecretString string type for secrets with concurrent access support.
type SecretString struct{ String }

// Print returns obfuscated string representation of value.
func (s *SecretString) Print() string {
	return "***"
}
