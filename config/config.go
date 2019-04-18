package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
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

type Field struct {
	Name    string
	Kind    reflect.Kind
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
	ff, err := getFields(tp.Elem())
	if err != nil {
		return nil, err
	}
	return &Config{cfg: reflect.ValueOf(cfg).Elem(), Fields: ff}, nil
}

// Set the value of a property of the provided config.
func (v *Config) Set(name, value string, kind reflect.Kind) error {
	v.Lock()
	defer v.Unlock()
	f := v.cfg.FieldByName(name)
	switch kind {
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		f.SetBool(b)
	case reflect.String:
		f.SetString(value)
	case reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		f.SetInt(v)
	case reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		f.SetFloat(v)
	default:
		return fmt.Errorf("unsupported kind: %v", kind)
	}
	return nil
}

func getFields(tp reflect.Type) ([]*Field, error) {
	dup := make(map[Source]string)
	var ff []*Field
	for i := 0; i < tp.NumField(); i++ {
		fld := tp.Field(i)
		kind := fld.Type.Kind()
		if !isKindSupported(kind) {
			return nil, fmt.Errorf("field %s is not supported(only bool, int64, float64 and string)", fld.Name)
		}
		f := &Field{
			Name:    fld.Name,
			Kind:    kind,
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

func isKindSupported(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool, reflect.Int64, reflect.Float64, reflect.String:
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
