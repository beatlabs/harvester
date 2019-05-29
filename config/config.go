package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"

	"github.com/taxibeat/harvester/log"
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
)

// Field definition of a config value that can change.
type Field struct {
	Name    string
	Type    string
	Method  reflect.Value
	Version uint64
	Sources map[Source]string
}

// Config manages configuration and handles updates on the values.
type Config struct {
	Fields []*Field
	sync.Mutex
	cfg reflect.Value
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
	return &Config{cfg: val, Fields: ff}, nil
}

// Set the value of a property of the provided config.
func (v *Config) Set(name, value string, version uint64) error {
	fld := v.getField(name)
	if fld == nil {
		return fmt.Errorf("field %s not found", name)
	}
	if version <= fld.Version {
		log.Warnf("version %d is older or same as field's %s version %d", version, fld.Name, fld.Version)
		return nil
	}
	f := v.cfg.FieldByName(name)
	switch fld.Type {
	case "Bool":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		val := reflect.New(reflect.TypeOf(v))
		callSetter(&f, &val)
	case "String":
		val := reflect.New(reflect.TypeOf(value))
		callSetter(&f, &val)
	case "Int64":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		val := reflect.New(reflect.TypeOf(v))
		callSetter(&f, &val)
	case "Float64":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		val := reflect.New(reflect.TypeOf(v))
		callSetter(&f, &val)
	}
	return nil
}

func (v *Config) getField(name string) *Field {
	for _, f := range v.Fields {
		if f.Name == name {
			return f
		}
	}
	return nil
}

func callSetter(f *reflect.Value, arg *reflect.Value) {
	method := f.MethodByName("Set")
	rr := method.Call([]reflect.Value{*arg})
	if len(rr) > 0 {
		log.Warnf("the set call returned %d values: %v", len(rr), rr)
	}
}

func getFields(tp reflect.Type, val *reflect.Value) ([]*Field, error) {
	dup := make(map[Source]string)
	var ff []*Field
	for i := 0; i < tp.NumField(); i++ {
		fld := tp.Field(i)
		if !isTypeSupported(fld.Type) {
			return nil, fmt.Errorf("field %s is not supported(only bool, int64, float64 and string from the sync package of harvester)", fld.Name)
		}
		str := val.FieldByIndex([]int{i})
		cnt := str.NumMethod()
		_ = cnt
---
		f := &Field{
			Name:    fld.Name,
			Type:    fld.Type.Name(),
			Method:  str.MethodByName("Set"),
			Version: 0,
			Sources: make(map[Source]string),
		}
		value, ok := fld.Tag.Lookup(string(SourceSeed))
		if ok {
			f.Sources[SourceSeed] = value
		}
		value, ok = fld.Tag.Lookup(string(SourceEnv))
		if ok {
			f.Sources[SourceEnv] = value
		}
		value, ok = fld.Tag.Lookup(string(SourceConsul))
		if ok {
			if isKeyValueDuplicate(dup, SourceConsul, value) {
				return nil, fmt.Errorf("duplicate value %s for source %s", value, SourceConsul)
			}
			f.Sources[SourceConsul] = value
		}
		ff = append(ff, f)
	}
	return ff, nil
}

func isTypeSupported(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	if t.PkgPath() != "github.com/taxibeat/harvester/sync" {
		return false
	}
	switch t.Name() {
	case "Bool", "Int64", "Float64", "String":
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
