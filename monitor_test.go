package harvester

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	Name   string `seed:"John Doe" env:"ENV_NAME" consul:"/config/name"`
	Age    int    `seed:"18" env:"ENV_AGE" consul:"/config/age"`
	HasJob bool   `seed:"true" env:"ENV_HAS_JOB" consul:"/config/has-job"`
}

func TestTags(t *testing.T) {
	cfg := &TestConfig{}
	r := reflect.TypeOf(cfg)
	v := reflect.ValueOf(cfg)
	f := v.Elem().Field(0)
	f.SetString("Test")
	assert.Equal(t, "Test", cfg.Name)
	assert.Equal(t, "Name", r.Elem().Field(0).Name)
}

func TestTags1(t *testing.T) {
	cfg := &TestConfig{}
	r := reflect.TypeOf(cfg)
	f := r.Elem().Field(0)
	assert.Equal(t, "Name", f.Name)
	assert.Equal(t, "John Doe", f.Tag.Get("seed"))
}

func Extract(tag reflect.StructTag) []reflect.StructTag {
	tags := strings.Split(string(tag), " ")
	if len(tags) == 0 {
		return nil
	}
	tt := make([]reflect.StructTag, len(tags))
	for i := 0; i < len(tags); i++ {
		tt[i] = reflect.StructTag(tags[i])
	}
	return tt
}
