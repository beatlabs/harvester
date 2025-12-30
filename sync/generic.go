// Package sync handles synchronized read and write access to config values.
package sync

import (
	"encoding/json"
	"sync"
)

// Value is a generic type with concurrent access support.
type Value[T any] struct {
	rw    sync.RWMutex
	value T
}

// Get returns the internal value.
func (v *Value[T]) Get() T {
	v.rw.RLock()
	defer v.rw.RUnlock()
	return v.value
}

// Set a value.
func (v *Value[T]) Set(value T) {
	v.rw.Lock()
	defer v.rw.Unlock()
	v.value = value
}

// MarshalJSON returns the JSON encoding of the value.
func (v *Value[T]) MarshalJSON() ([]byte, error) {
	v.rw.RLock()
	defer v.rw.RUnlock()
	return json.Marshal(v.value)
}

// UnmarshalJSON returns the JSON encoding of the value.
func (v *Value[T]) UnmarshalJSON(d []byte) error {
	v.rw.Lock()
	defer v.rw.Unlock()
	return json.Unmarshal(d, &v.value)
}
