package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/beatlabs/harvester/log"
)

// Source definition.
type Source string

const (
	// SourceSeed defines a seed value.
	SourceSeed Source = "seed"
	// SourceEnv defines a value from environment variables.
	SourceEnv Source = "env"
	// SourceConsul defines a value from consul.
	SourceConsul Source = "consul"
	// SourceFlag defines a value from CLI flag.
	SourceFlag Source = "flag"
	// SourceSecret defines that the value is a secret, can should not be logged in it's entirety.
	SourceSecret Source = "secret"
)

var sourceTags = [...]Source{SourceSeed, SourceEnv, SourceConsul, SourceFlag, SourceSecret}

// CfgType represents an interface which any config field type must implement.
type CfgType interface {
	fmt.Stringer
	SetString(string) error
}

// Field definition of a config value that can change.
type Field struct {
	name        string
	tp          string
	version     uint64
	structField CfgType
	sources     map[Source]string
	secret      bool
}

// newField constructor.
func newField(prefix string, fld reflect.StructField, val reflect.Value) *Field {
	f := &Field{
		name:        prefix + fld.Name,
		tp:          fld.Type.Name(),
		version:     0,
		structField: val.Addr().Interface().(CfgType),
		sources:     make(map[Source]string),
	}

	for _, tag := range sourceTags {
		value, ok := fld.Tag.Lookup(string(tag))
		if ok {
			f.sources[tag] = value
		}
	}

	_, f.secret = f.sources[SourceSecret]

	return f
}

// Name getter.
func (f *Field) Name() string {
	return f.name
}

// Type getter.
func (f *Field) Type() string {
	return f.tp
}

// Sources getter.
func (f *Field) Sources() map[Source]string {
	return f.sources
}

// String returns string representation of field's value.
func (f *Field) String() string {
	return f.structField.String()
}

// LogValue returns string representation of field's value, but masks the value with asterisks when it's a secret.
func (f *Field) LogValue() string {
	value := f.structField.String()
	if f.secret && len(value) > 3 {
		return fmt.Sprintf("%s%s", value[:3], strings.Repeat("*", len(value)-3))
	}
	return value
}

// Set the value of the field.
func (f *Field) Set(value string, version uint64) error {
	if version != 0 && version <= f.version {
		log.Warnf("version %d is older or same as the field's %s", version, f.name)
		return nil
	}

	if err := f.structField.SetString(value); err != nil {
		return err
	}

	f.version = version
	log.Infof("field %s updated with value %v, version: %d", f.name, f.LogValue(), version)
	return nil
}

// Config manages configuration and handles updates on the values.
type Config struct {
	Fields []*Field
}

// New creates a new monitor.
func New(cfg interface{}) (*Config, error) {
	if cfg == nil {
		return nil, errors.New("configuration is nil")
	}

	ff, err := newParser().ParseCfg(cfg)
	if err != nil {
		return nil, err
	}

	return &Config{Fields: ff}, nil
}
