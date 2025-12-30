// Package sync handles synchronized read and write access to config values.
package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Bool type with concurrent access support.
type Bool struct {
	Value[bool]
}

// String returns string representation of value.
func (b *Bool) String() string {
	if b.Get() {
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
	Value[int64]
}

// String returns string representation of value.
func (i *Int64) String() string {
	return strconv.FormatInt(i.Get(), 10)
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
	Value[float64]
}

// String returns string representation of value.
func (f *Float64) String() string {
	return fmt.Sprintf("%f", f.Get())
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
	Value[string]
}

// String returns string representation of value.
func (s *String) String() string {
	return s.Get()
}

// SetString parses and sets a value from string type.
func (s *String) SetString(val string) error {
	s.Set(val)
	return nil
}

// TimeDuration is Time.Duration type with concurrent access support.
type TimeDuration struct {
	Value[time.Duration]
}

// String returns string representation of value.
func (s *TimeDuration) String() string {
	return s.Get().String()
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
	Value[string]
}

// MarshalJSON returns the JSON encoding of the value (obfuscated).
func (s *Secret) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
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

// Regexp type with concurrent access support.
type Regexp struct {
	Value[*regexp.Regexp]
}

// MarshalJSON returns the JSON encoding of the value.
func (r *Regexp) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// UnmarshalJSON parses the JSON encoding of the value.
func (r *Regexp) UnmarshalJSON(d []byte) error {
	var str string
	err := json.Unmarshal(d, &str)
	if err != nil {
		return err
	}
	regex, err := regexp.Compile(str)
	if err != nil {
		return err
	}
	r.Set(regex)
	return nil
}

// String returns a string representation of the value.
func (r *Regexp) String() string {
	regex := r.Get()
	if regex == nil {
		return ""
	}
	return regex.String()
}

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
	Value[map[string]string]
}

// String returns a string representation of the value.
func (s *StringMap) String() string {
	m := s.Get()
	b := new(bytes.Buffer)
	firstChar := ""
	for key, value := range m {
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
	Value[[]string]
}

// String returns a string representation of the value.
func (s *StringSlice) String() string {
	return strings.Join(s.Get(), ",")
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
