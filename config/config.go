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

// Field definition of a config value that can change.
type Field struct {
	name    string
	tp      string
	version uint64
	setter  reflect.Value
	printer reflect.Value
	sources map[Source]string
}

// NewField constructor.
func NewField(fld *reflect.StructField, val *reflect.Value) (*Field, error) {
	if !isTypeSupported(fld.Type) {
		return nil, fmt.Errorf("field %s is not supported (only types from the sync package of harvester)", fld.Name)
	}
	f := &Field{
		name:    fld.Name,
		tp:      fld.Type.Name(),
		version: 0,
		setter:  val.FieldByName(fld.Name).Addr().MethodByName("Set"),
		printer: val.FieldByName(fld.Name).Addr().MethodByName("String"),
		sources: make(map[Source]string),
	}
	value, ok := fld.Tag.Lookup(string(SourceSeed))
	if ok {
		f.sources[SourceSeed] = value
	}
	value, ok = fld.Tag.Lookup(string(SourceEnv))
	if ok {
		f.sources[SourceEnv] = value
	}
	value, ok = fld.Tag.Lookup(string(SourceConsul))
	if ok {
		f.sources[SourceConsul] = value
	}
	value, ok = fld.Tag.Lookup(string(SourceFlag))
	if ok {
		f.sources[SourceFlag] = value
	}
	return f, nil
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
	tp := reflect.TypeOf(cfg)
	if tp.Kind() != reflect.Ptr {
		return nil, errors.New("configuration should be a pointer type")
	}
	val := reflect.ValueOf(cfg).Elem()
	ff, err := getFields(tp.Elem(), &val)
	if err != nil {
		return nil, err
	}
	return &Config{Fields: ff}, nil
}

func getFields(tp reflect.Type, val *reflect.Value) ([]*Field, error) {
	dup := make(map[Source]string)
	var ff []*Field
	for i := 0; i < tp.NumField(); i++ {
		f := tp.Field(i)
		fld, err := NewField(&f, val)
		if err != nil {
			return nil, err
		}
		value, ok := fld.Sources()[SourceConsul]
		if ok {
			if isKeyValueDuplicate(dup, SourceConsul, value) {
				return nil, fmt.Errorf("duplicate value %v for source %s", fld, SourceConsul)
			}
		}
		ff = append(ff, fld)
	}
	return ff, nil
}

func isTypeSupported(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	if t.PkgPath() != "github.com/beatlabs/harvester/sync" {
		return false
	}
	switch t.Name() {
	case "Bool", "Int64", "Float64", "String", "Secret":
		return true
	default:
		return false
	}
}

func isKeyValueDuplicate(dup map[Source]string, src Source, value string) bool {
	v, ok := dup[src]
	if ok {
		if value == v {
			return true
		}
	}
	dup[src] = value
	return false
}
