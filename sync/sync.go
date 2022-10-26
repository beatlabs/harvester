// Package sync handles synchronized read and write access to config values.
package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
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

// MarshalJSON returns the JSON encoding of the value.
func (b *Bool) MarshalJSON() ([]byte, error) {
	b.rw.RLock()
	defer b.rw.RUnlock()
	return json.Marshal(b.value)
}

// MarshalJSON returns the JSON encoding of the value.
func (b *Bool) UnmarshalJSON(d []byte) error {
	b.rw.RLock()
	defer b.rw.RUnlock()
	return json.Unmarshal(d, &b.value)
}

// String returns string representation of value.
func (b *Bool) String() string {
	b.rw.Lock()
	defer b.rw.Unlock()
	if b.value {
		return "true"
	}
	return "false"
}

// SetString parses and sets a value from string type.
func (b *Bool) SetString(val string) error {
	v, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}
	b.Set(v)
	return nil
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

// MarshalJSON returns the JSON encoding of the value.
func (i *Int64) MarshalJSON() ([]byte, error) {
	i.rw.RLock()
	defer i.rw.RUnlock()
	return json.Marshal(i.value)
}

// MarshalJSON returns the JSON encoding of the value.
func (i *Int64) UnmarshalJSON(d []byte) error {
	i.rw.RLock()
	defer i.rw.RUnlock()
	return json.Unmarshal(d, &i.value)
}

// String returns string representation of value.
func (i *Int64) String() string {
	i.rw.RLock()
	defer i.rw.RUnlock()
	return fmt.Sprintf("%d", i.value)
}

// SetString parses and sets a value from string type.
func (i *Int64) SetString(val string) error {
	v, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return err
	}
	i.Set(v)
	return nil
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

// MarshalJSON returns the JSON encoding of the value.
func (f *Float64) MarshalJSON() ([]byte, error) {
	f.rw.RLock()
	defer f.rw.RUnlock()
	return json.Marshal(f.value)
}

// MarshalJSON returns the JSON encoding of the value.
func (f *Float64) UnmarshalJSON(d []byte) error {
	f.rw.RLock()
	defer f.rw.RUnlock()
	return json.Unmarshal(d, &f.value)
}

// String returns string representation of value.
func (f *Float64) String() string {
	f.rw.RLock()
	defer f.rw.RUnlock()
	return fmt.Sprintf("%f", f.value)
}

// SetString parses and sets a value from string type.
func (f *Float64) SetString(val string) error {
	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return err
	}
	f.Set(v)
	return nil
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

// MarshalJSON returns the JSON encoding of the value.
func (s *String) MarshalJSON() ([]byte, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return json.Marshal(s.value)
}

// MarshalJSON returns the JSON encoding of the value.
func (s *String) UnmarshalJSON(d []byte) error {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return json.Unmarshal(d, &s.value)
}

// String returns string representation of value.
func (s *String) String() string {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return s.value
}

// SetString parses and sets a value from string type.
func (s *String) SetString(val string) error {
	s.Set(val)
	return nil
}

// TimeDuration is Time.Duration type with concurrent access support.
type TimeDuration struct {
	rw    sync.RWMutex
	value time.Duration
}

// Get returns the internal value.
func (s *TimeDuration) Get() time.Duration {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return s.value
}

// Set a value.
func (s *TimeDuration) Set(value time.Duration) {
	s.rw.Lock()
	defer s.rw.Unlock()
	s.value = value
}

// MarshalJSON returns the JSON encoding of the value.
func (s *TimeDuration) MarshalJSON() ([]byte, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return json.Marshal(s.value)
}

// MarshalJSON returns the JSON encoding of the value.
func (s *TimeDuration) UnmarshalJSON(d []byte) error {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return json.Unmarshal(d, &s.value)
}

// String returns string representation of value.
func (s *TimeDuration) String() string {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return s.value.String()
}

// SetString parses and sets a value from string type.
func (s *TimeDuration) SetString(val string) error {
	value, err := time.ParseDuration(val)
	if err != nil {
		return err
	}
	s.Set(value)
	return nil
}

// Secret string type for secrets with concurrent access support.
type Secret struct {
	rw    sync.RWMutex
	value string
}

// Get returns the internal value.
func (s *Secret) Get() string {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return s.value
}

// Set a value.
func (s *Secret) Set(value string) {
	s.rw.Lock()
	defer s.rw.Unlock()
	s.value = value
}

// MarshalJSON returns the JSON encoding of the value.
func (s *Secret) MarshalJSON() (out []byte, err error) {
	return json.Marshal(s.String())
}

// MarshalJSON returns the JSON encoding of the value.
func (s *Secret) UnmarshalJSON(d []byte) error {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return json.Unmarshal(d, &s.value)
}

// String returns obfuscated string representation of value.
func (s *Secret) String() string {
	return "***"
}

// SetString parses and sets a value from string type.
func (s *Secret) SetString(val string) error {
	s.Set(val)
	return nil
}

type Regexp struct {
	rw    sync.RWMutex
	value *regexp.Regexp
}

// Get returns the internal value.
func (r *Regexp) Get() *regexp.Regexp {
	r.rw.RLock()
	defer r.rw.RUnlock()
	return r.value
}

// Set a value.
func (r *Regexp) Set(value *regexp.Regexp) {
	r.rw.Lock()
	defer r.rw.Unlock()
	r.value = value
}

// MarshalJSON returns the JSON encoding of the value.
func (r *Regexp) MarshalJSON() ([]byte, error) {
	r.rw.RLock()
	defer r.rw.RUnlock()
	return json.Marshal(r.value.String())
}

// UnmarshalJSON returns the JSON encoding of the value.
func (r *Regexp) UnmarshalJSON(d []byte) error {
	var str string
	err := json.Unmarshal(d, &str)
	if err != nil {
		fmt.Println("json unmarshal")
		return err
	}
	regex, err := regexp.Compile(str)
	if err != nil {
		fmt.Println("regex compile")
		return err
	}
	r.Set(regex)
	return nil
}

// String returns a string representation of the value.
func (r *Regexp) String() string {
	r.rw.RLock()
	defer r.rw.RUnlock()
	return r.value.String()
}

//
// SetString parses and sets a value from string type.
func (r *Regexp) SetString(val string) error {
	compiled, err := regexp.Compile(val)
	if err != nil {
		return err
	}
	r.Set(compiled)
	return nil
}

// StringMap is a map[string]string type with concurrent access support.
type StringMap struct {
	rw    sync.RWMutex
	value map[string]string
}

// Get returns the internal value.
func (s *StringMap) Get() map[string]string {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return s.value
}

// Set a value.
func (s *StringMap) Set(value map[string]string) {
	s.rw.Lock()
	defer s.rw.Unlock()
	s.value = value
}

// MarshalJSON returns the JSON encoding of the value.
func (s *StringMap) MarshalJSON() ([]byte, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return json.Marshal(s.value)
}

// UnmarshalJSON returns the JSON encoding of the value.
func (s *StringMap) UnmarshalJSON(d []byte) error {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return json.Unmarshal(d, &s.value)
}

// String returns a string representation of the value.
func (s *StringMap) String() string {
	s.rw.RLock()
	defer s.rw.RUnlock()
	b := new(bytes.Buffer)
	firstChar := ""
	for key, value := range s.value {
		_, _ = fmt.Fprintf(b, "%s%s=%q", firstChar, key, value)
		firstChar = ","
	}
	return b.String()
}

// SetString parses and sets a value from string type.
func (s *StringMap) SetString(val string) error {
	dict := make(map[string]string)
	if val == "" || strings.TrimSpace(val) == "" {
		s.Set(dict)
		return nil
	}
	for _, pair := range strings.Split(val, ",") {
		items := strings.SplitN(pair, ":", 2)
		if len(items) != 2 {
			return fmt.Errorf("map must be formatted as `key:value`, got %q", pair)
		}
		key, value := strings.TrimSpace(items[0]), strings.TrimSpace(items[1])
		dict[key] = value
	}
	s.Set(dict)
	return nil
}

// StringSlice is a []string type with concurrent access support.
type StringSlice struct {
	rw    sync.RWMutex
	value []string
}

// Get returns the internal value.
func (s *StringSlice) Get() []string {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return s.value
}

// Set a value.
func (s *StringSlice) Set(value []string) {
	s.rw.Lock()
	defer s.rw.Unlock()
	s.value = value
}

// MarshalJSON returns the JSON encoding of the value.
func (s *StringSlice) MarshalJSON() ([]byte, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return json.Marshal(s.value)
}

// UnmarshalJSON returns the JSON encoding of the value.
func (s *StringSlice) UnmarshalJSON(d []byte) error {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return json.Unmarshal(d, &s.value)
}

// String returns a string representation of the value.
func (s *StringSlice) String() string {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return strings.Join(s.value, ",")
}

// SetString parses and sets a value from string type.
func (s *StringSlice) SetString(val string) error {
	slice := make([]string, 0)
	if val == "" || strings.TrimSpace(val) == "" {
		s.Set(slice)
		return nil
	}
	for _, item := range strings.Split(val, ",") {
		slice = append(slice, strings.TrimSpace(item))
	}
	s.Set(slice)
	return nil
}
