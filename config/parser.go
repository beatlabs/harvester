package config

import (
	"errors"
	"fmt"
	"reflect"
)

type structFieldType uint

const (
	typeInvalid structFieldType = iota
	typeField
	typeStruct
)

type parser struct {
	dups map[Source]map[string]bool
}

func newParser() *parser {
	return &parser{}
}

func (p *parser) ParseCfg(cfg interface{}, chNotify chan<- ChangeNotification) ([]*Field, error) {
	p.dups = make(map[Source]map[string]bool)

	value := reflect.ValueOf(cfg)
	if !value.IsValid() {
		return nil, errors.New("configuration is nil")
	}

	tp := value.Type()
	if tp.Kind() != reflect.Ptr {
		return nil, errors.New("configuration should be a pointer type")
	}
	if value.IsNil() {
		return nil, errors.New("configuration pointer is nil")
	}
	if tp.Elem().Kind() != reflect.Struct {
		return nil, errors.New("configuration pointer should point to a struct")
	}

	return p.getFields("", tp.Elem(), value.Elem(), chNotify)
}

func (p *parser) getFields(prefix string, tp reflect.Type, val reflect.Value, chNotify chan<- ChangeNotification) ([]*Field, error) {
	var ff []*Field

	for i := 0; i < tp.NumField(); i++ {
		f := tp.Field(i)
		fieldValue := val.Field(i)

		typ, err := p.getStructFieldType(f, fieldValue)
		if err != nil {
			return nil, err
		}

		switch typ {
		case typeField:
			fld, err := p.createField(prefix, f, fieldValue, chNotify)
			if err != nil {
				return nil, err
			}
			ff = append(ff, fld)
		case typeStruct:
			nested, err := p.getFields(prefix+f.Name, f.Type, fieldValue, chNotify)
			if err != nil {
				return nil, err
			}
			ff = append(ff, nested...)
		case typeInvalid:
		}
	}
	return ff, nil
}

func (p *parser) createField(prefix string, f reflect.StructField, val reflect.Value, chNotify chan<- ChangeNotification) (*Field, error) {
	fld, err := newField(prefix, f, val, chNotify)
	if err != nil {
		return nil, err
	}

	// Duplicate key detection is intentionally limited to Consul and Redis.
	// For env, flag, and file tags, the Go compiler enforces struct field name
	// uniqueness, which makes duplicate tag values harmless in practice. For
	// Consul and Redis, multiple fields could share the same remote key string,
	// causing silent overwrites at runtime — so we reject duplicates eagerly here.
	value, ok := fld.Sources()[SourceConsul]
	if ok {
		if p.isKeyValueDuplicate(SourceConsul, value) {
			return nil, fmt.Errorf("duplicate value %v for source %s", fld, SourceConsul)
		}
	}
	value, ok = fld.Sources()[SourceRedis]
	if ok {
		if p.isKeyValueDuplicate(SourceRedis, value) {
			return nil, fmt.Errorf("duplicate value %v for source %s", fld, SourceRedis)
		}
	}
	value, ok = fld.Sources()[SourceFlag]
	if ok {
		if p.isKeyValueDuplicate(SourceFlag, value) {
			return nil, fmt.Errorf("duplicate value %v for source %s", fld, SourceFlag)
		}
	}
	return fld, nil
}

func (p *parser) isKeyValueDuplicate(src Source, value string) bool {
	if p.dups[src] == nil {
		p.dups[src] = make(map[string]bool)
	}
	if p.dups[src][value] {
		return true
	}
	p.dups[src][value] = true
	return false
}

func (p *parser) getStructFieldType(f reflect.StructField, val reflect.Value) (structFieldType, error) {
	t := f.Type
	if t.Kind() != reflect.Struct {
		return typeInvalid, fmt.Errorf("only struct type supported for %s", f.Name)
	}

	cfgType := reflect.TypeOf((*CfgType)(nil)).Elem()

	for _, tag := range sourceTags {
		if _, ok := f.Tag.Lookup(string(tag)); ok {
			if !val.Addr().CanInterface() {
				return typeInvalid, fmt.Errorf("field %s must be exported", f.Name)
			}
			if !val.Addr().Type().Implements(cfgType) {
				return typeInvalid, fmt.Errorf("field %s must implement CfgType interface", f.Name)
			}
			return typeField, nil
		}
	}

	return typeStruct, nil
}
