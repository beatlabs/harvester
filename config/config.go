package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

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
)

var sourceTags = [...]Source{SourceSeed, SourceEnv, SourceConsul, SourceFlag}

// Field definition of a config value that can change.
type Field struct {
	name    string
	tp      string
	version uint64
	setter  reflect.Value
	printer reflect.Value
	sources map[Source]string
}

// newField constructor.
func newField(prefix string, fld reflect.StructField, val reflect.Value) *Field {
	f := &Field{
		name:    prefix + fld.Name,
		tp:      fld.Type.Name(),
		version: 0,
		setter:  val.Addr().MethodByName("Set"),
		printer: val.Addr().MethodByName("String"),
		sources: make(map[Source]string),
	}

	for _, tag := range sourceTags {
		value, ok := fld.Tag.Lookup(string(tag))
		if ok {
			f.sources[tag] = value
		}
	}

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
	vv := f.printer.Call([]reflect.Value{})
	if len(vv) > 0 {
		return vv[0].String()
	}
	return ""
}

// Set the value of the field.
func (f *Field) Set(value string, version uint64) error {
	if version != 0 && version <= f.version {
		log.Warnf("version %d is older or same as the field's %s", version, f.name)
		return nil
	}
	var arg interface{}
	switch f.tp {
	case "Bool":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		arg = v
	case "String", "Secret":
		arg = value
	case "Int64":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		arg = v
	case "Float64":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		arg = v
	}
	rr := f.setter.Call([]reflect.Value{reflect.ValueOf(arg)})
	if len(rr) > 0 {
		return fmt.Errorf("the set call returned %d values: %v", len(rr), rr)
	}
	f.version = version
	log.Infof("field %s updated with value %v, version: %d", f.name, f, version)
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
