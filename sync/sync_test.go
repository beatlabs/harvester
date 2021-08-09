package sync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	var b Bool
	ch := make(chan struct{})
	go func() {
		b.Set(true)
		ch <- struct{}{}
	}()
	<-ch
	assert.True(t, b.Get())
	assert.Equal(t, "true", b.String())
}

func TestBool_SetString(t *testing.T) {
	var b Bool
	assert.Error(t, b.SetString("wrong"))
	assert.NoError(t, b.SetString("true"))
	assert.True(t, b.Get())
}

func TestInt64(t *testing.T) {
	var i Int64
	ch := make(chan struct{})
	go func() {
		i.Set(10)
		ch <- struct{}{}
	}()
	<-ch
	assert.Equal(t, int64(10), i.Get())
	assert.Equal(t, "10", i.String())
}

func TestInt64_SetString(t *testing.T) {
	var i Int64
	assert.Error(t, i.SetString("wrong"))
	assert.NoError(t, i.SetString("10"))
	assert.Equal(t, int64(10), i.Get())
}

func TestFloat64(t *testing.T) {
	var f Float64
	ch := make(chan struct{})
	go func() {
		f.Set(1.23)
		ch <- struct{}{}
	}()
	<-ch
	assert.Equal(t, 1.23, f.Get())
	assert.Equal(t, "1.230000", f.String())
}

func TestFloat64_SetString(t *testing.T) {
	var f Float64
	assert.Error(t, f.SetString("wrong"))
	assert.NoError(t, f.SetString("1.230000"))
	assert.Equal(t, 1.23, f.Get())
}

func TestString(t *testing.T) {
	var s String
	ch := make(chan struct{})
	go func() {
		s.Set("Hello")
		ch <- struct{}{}
	}()
	<-ch
	assert.Equal(t, "Hello", s.Get())
	assert.Equal(t, "Hello", s.String())
}

func TestString_SetString(t *testing.T) {
	var s String
	assert.NoError(t, s.SetString("foo"))
	assert.Equal(t, "foo", s.Get())
}

func TestSecret(t *testing.T) {
	var s Secret
	ch := make(chan struct{})
	go func() {
		s.Set("Hello")
		ch <- struct{}{}
	}()
	<-ch
	assert.Equal(t, "Hello", s.Get())
	assert.Equal(t, "***", s.String())
}

func TestSecret_SetString(t *testing.T) {
	var s Secret
	assert.NoError(t, s.SetString("foo"))
	assert.Equal(t, "foo", s.Get())
}

func TestTimeDuration(t *testing.T) {
	var f TimeDuration
	testTime := 3 * time.Second
	ch := make(chan struct{})
	go func() {
		f.Set(testTime)
		ch <- struct{}{}
	}()
	<-ch
	assert.Equal(t, testTime, f.Get())
	assert.Equal(t, testTime.String(), f.String())
}

func TestTimeDuration_SetString(t *testing.T) {
	var f TimeDuration
	assert.Error(t, f.SetString("kuku"))
	assert.NoError(t, f.SetString("3s"))
	assert.Equal(t, 3*time.Second, f.Get())
}

func TestStringMap(t *testing.T) {
	var sm StringMap
	ch := make(chan struct{})
	go func() {
		sm.Set(map[string]string{"key": "value"})
		ch <- struct{}{}
	}()
	<-ch
	assert.Equal(t, map[string]string{"key": "value"}, sm.Get())
	assert.Equal(t, "key=\"value\"", sm.String())
}

func TestStringMap_SetString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		result      map[string]string
		throwsError bool
	}{
		{"empty", "", map[string]string{}, false},
		{"empty with spaces", "   ", map[string]string{}, false},
		{"single item", "key:value", map[string]string{"key": "value"}, false},
		{"single item with route as val", "key:http://thing", map[string]string{"key": "http://thing"}, false},
		{"key without value", "key", nil, true},
		{"multiple items", "key1:value,key2:value", map[string]string{"key1": "value", "key2": "value"}, false},
		{"multiple items with spaces", " key1 : value , key2 :value ", map[string]string{"key1": "value", "key2": "value"}, false},
		{"multiple urls", "key1:http://one,key2:https://two", map[string]string{"key1": "http://one", "key2": "https://two"}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sm := StringMap{}

			err := sm.SetString(test.input)
			if test.throwsError {
				assert.Error(t, err)
			}

			assert.Equal(t, test.result, sm.Get())
		})
	}
}

func TestStringMap_SetString_DoesntOverrideValueIfError(t *testing.T) {
	sm := StringMap{}

	assert.NoError(t, sm.SetString("k1:v1"))
	assert.Equal(t, map[string]string{"k1": "v1"}, sm.Get())

	assert.Error(t, sm.SetString("k1:v1,k2:v2,k3"))
	assert.Equal(t, map[string]string{"k1": "v1"}, sm.Get())
}
