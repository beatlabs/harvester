package sync

import (
	"fmt"
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

	d, err := b.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, "true", string(d))
}

func TestBool_SetString(t *testing.T) {
	var b Bool
	assert.Error(t, b.SetString("wrong"))
	assert.NoError(t, b.SetString("true"))
	assert.True(t, b.Get())
}

func TestBool_UnmarshalJSON(t *testing.T) {
	var b Bool
	err := b.UnmarshalJSON([]byte("wrong"))
	assert.Error(t, err)
	assert.False(t, b.Get())

	err = b.UnmarshalJSON([]byte("true"))
	assert.NoError(t, err)
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

	d, err := i.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, "10", string(d))
}

func TestInt64_SetString(t *testing.T) {
	var i Int64
	assert.Error(t, i.SetString("wrong"))
	assert.NoError(t, i.SetString("10"))
	assert.Equal(t, int64(10), i.Get())
}

func TestInt64_UnmarshalJSON(t *testing.T) {
	var b Int64
	err := b.UnmarshalJSON([]byte("123.544")) // this is wrong
	assert.Error(t, err)
	assert.Equal(t, int64(0), b.Get())

	err = b.UnmarshalJSON([]byte("123"))
	assert.NoError(t, err)
	assert.Equal(t, int64(123), b.Get())
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

	d, err := f.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, "1.23", string(d))
}

func TestFloat64_UnmarshalJSON(t *testing.T) {
	var b Float64
	err := b.UnmarshalJSON([]byte("wrong"))
	assert.Error(t, err)
	assert.Equal(t, float64(0), b.Get())

	err = b.UnmarshalJSON([]byte("123.321"))
	assert.NoError(t, err)
	assert.Equal(t, float64(123.321), b.Get())
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

	d, err := s.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, `"Hello"`, string(d))
}

func TestString_SetString(t *testing.T) {
	var s String
	assert.NoError(t, s.SetString("foo"))
	assert.Equal(t, "foo", s.Get())
}

func TestString_UnmarshalJSON(t *testing.T) {
	var b String
	err := b.UnmarshalJSON([]byte(`foo`))
	assert.Error(t, err)
	assert.Equal(t, "", b.Get())

	err = b.UnmarshalJSON([]byte(`"foo"`))
	assert.NoError(t, err)
	assert.Equal(t, "foo", b.Get())
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

	d, err := s.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, `"***"`, string(d))
}

func TestSecret_SetString(t *testing.T) {
	var s Secret
	assert.NoError(t, s.SetString("foo"))
	assert.Equal(t, "foo", s.Get())
}

func TestSecret_UnmarshalJSON(t *testing.T) {
	var b String
	err := b.UnmarshalJSON([]byte(`"foo"`))
	assert.NoError(t, err)
	assert.Equal(t, "foo", b.Get())
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

	d, err := f.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%d", testTime.Nanoseconds()), string(d))
}

func TestTimeDuration_SetString(t *testing.T) {
	var f TimeDuration
	assert.Error(t, f.SetString("kuku"))
	assert.NoError(t, f.SetString("3s"))
	assert.Equal(t, 3*time.Second, f.Get())
}

func TestTimeDuration_UnmarshalJSON(t *testing.T) {
	var b TimeDuration
	err := b.UnmarshalJSON([]byte(`foo`))
	assert.Error(t, err)
	assert.Equal(t, time.Duration(0), b.Get())

	err = b.UnmarshalJSON([]byte(`1`))
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(1), b.Get())
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

	d, err := sm.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, `{"key":"value"}`, string(d))
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

func TestStringMap_UnmarshalJSON(t *testing.T) {
	var b StringMap
	err := b.UnmarshalJSON([]byte(`wrong`))
	assert.Error(t, err)
	assert.Equal(t, map[string]string(nil), b.Get())

	err = b.UnmarshalJSON([]byte(`{ "a": "b" }`))
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "b"}, b.Get())
}

func TestStringSlice(t *testing.T) {
	var sl StringSlice
	ch := make(chan struct{})
	go func() {
		sl.Set([]string{"value1", "value2"})
		ch <- struct{}{}
	}()
	<-ch
	assert.Equal(t, []string{"value1", "value2"}, sl.Get())
	assert.Equal(t, "value1,value2", sl.String())

	d, err := sl.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, `["value1","value2"]`, string(d))
}

func TestStringSlice_SetString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		result      []string
		throwsError bool
	}{
		{"empty", "", []string{}, false},
		{"empty with spaces", "   ", []string{}, false},
		{"single item", "value", []string{"value"}, false},
		{"multiple items", "value1,value2", []string{"value1", "value2"}, false},
		{"multiple items with spaces", "  value1 ,  value2 ", []string{"value1", "value2"}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sm := StringSlice{}

			err := sm.SetString(test.input)
			if test.throwsError {
				assert.Error(t, err)
			}

			assert.Equal(t, test.result, sm.Get())
		})
	}
}

func TestStringSlice_UnmarshalJSON(t *testing.T) {
	var b StringSlice
	err := b.UnmarshalJSON([]byte(`wrong`))
	assert.Error(t, err)
	assert.Equal(t, []string(nil), b.Get())

	err = b.UnmarshalJSON([]byte(`["a", "b"]`))
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, b.Get())
}
