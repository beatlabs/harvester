package config

import (
	"errors"
	"fmt"
	"reflect"
)

type parser struct {
	dups map[Source]string
}

func newParser() *parser {
	return &parser{}
}

func (p *parser) GetFields(cfg interface{}) ([]*Field, error) {
	p.dups = make(map[Source]string)

	tp := reflect.TypeOf(cfg)
	if tp.Kind() != reflect.Ptr {
		return nil, errors.New("configuration should be a pointer type")
	}
	val := reflect.ValueOf(cfg).Elem()

	return p.getFields("", tp.Elem(), &val)
}

func (p *parser) getFields(prefix string, tp reflect.Type, val *reflect.Value) ([]*Field, error) {
	var ff []*Field
	for i := 0; i < tp.NumField(); i++ {
		f := tp.Field(i)
		fld, err := NewField(prefix, &f, val)
		if err != nil {
			if !p.isNestedTypeSupported(f.Type) {
				return nil, err
			}
			nested, err := p.getNestedFields(prefix, f, val.Field(i))
			if err != nil {
				return nil, err
			}
			ff = append(ff, nested...)
			continue
		}
		value, ok := fld.Sources()[SourceConsul]
		if ok {
			if p.isKeyValueDuplicate(SourceConsul, value) {
				return nil, fmt.Errorf("duplicate value %v for source %s", fld, SourceConsul)
			}
		}
		ff = append(ff, fld)
	}
	return ff, nil
}

func (p *parser) getNestedFields(prefix string, sf reflect.StructField, val reflect.Value) ([]*Field, error) {
	if val.Type().Kind() == reflect.Ptr && val.IsNil() {
		return make([]*Field, 0), nil
	}

	typ := val.Type()
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	return p.getFields(prefix+sf.Name, typ, &val)
}

func (p *parser) isNestedTypeSupported(t reflect.Type) bool {
	if t.Kind() == reflect.Struct {
		return true
	}
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		return true
	}
	return false
}

func (p *parser) isKeyValueDuplicate(src Source, value string) bool {
	v, ok := p.dups[src]
	if ok {
		if value == v {
			return true
		}
	}
	p.dups[src] = value
	return false
}
