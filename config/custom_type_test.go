package config_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/beatlabs/harvester/config"
	stdTypes "github.com/beatlabs/harvester/sync"
	"github.com/stretchr/testify/assert"
)

func TestCustomField(t *testing.T) {
	c := &testConfig{}
	cfg, err := config.New(c, nil)
	assert.NoError(t, err)
	err = cfg.Fields[0].Set("expected", 1)
	assert.NoError(t, err)
	err = cfg.Fields[1].Set("bar", 1)
	assert.NoError(t, err)
	assert.Equal(t, "expected", c.CustomValue.Get())
	assert.Equal(t, "bar", c.SomeString.Get())
}

func TestErrorValidationOnCustomField(t *testing.T) {
	c := &testConfig{}
	cfg, err := config.New(c, nil)
	assert.NoError(t, err)
	err = cfg.Fields[0].Set("not_expected", 1)
	assert.Error(t, err)
}

type testConcreteValue struct {
	m     sync.Mutex
	value string
}

func (f *testConcreteValue) Set(value string) {
	f.m.Lock()
	defer f.m.Unlock()
	f.value = value
}

func (f *testConcreteValue) Get() string {
	f.m.Lock()
	defer f.m.Unlock()
	return f.value
}

func (f *testConcreteValue) String() string {
	return f.Get()
}

func (f *testConcreteValue) SetString(value string) error {
	if value != "expected" {
		return fmt.Errorf("unable to store provided value")
	}
	f.Set(value)
	return nil
}

type testConfig struct {
	CustomValue testConcreteValue `seed:"expected"`
	SomeString  stdTypes.String   `seed:"foo"`
}
