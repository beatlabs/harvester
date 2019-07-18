package vault

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	testCases := []struct {
		desc        string
		address     string
		token       string
		timeout     time.Duration
		expectedErr error
	}{
		{
			desc:        "Empty address",
			address:     "",
			expectedErr: errors.New("address is empty"),
		},
		{
			desc:        "Proper creation without timeout",
			address:     "something",
			token:       "token",
			timeout:     0,
			expectedErr: nil,
		},
		{
			desc:        "Proper creation with a timeout",
			address:     "someting",
			token:       "token",
			timeout:     5 * time.Second,
			expectedErr: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			getter, err := New(tC.address, tC.token, tC.timeout)

			if tC.expectedErr != nil {
				assert.EqualError(t, err, tC.expectedErr.Error())
				assert.Nil(t, getter)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, &Getter{}, getter)
			}
		})
	}
}

func Test_Getter_Split_Key(t *testing.T) {
	testCases := []struct {
		desc         string
		inputKey     string
		expectedPath string
		expectedKey  string
		expectedErr  error
	}{
		{
			desc:        "Malformed key",
			inputKey:    "something",
			expectedErr: errors.New("malformed key: something"),
		},
		{
			desc:        "Malformed key with leading slash",
			inputKey:    "/something",
			expectedErr: errors.New("malformed key: /something"),
		},
		{
			desc:        "Malformed key with trailing slash",
			inputKey:    "something/",
			expectedErr: errors.New("malformed key: something/"),
		},
		{
			desc:         "Proper key (short)",
			inputKey:     "path-to-the-key/the-key",
			expectedPath: "path-to-the-key",
			expectedKey:  "the-key",
			expectedErr:  nil,
		},
		{
			desc:         "Proper key (long)",
			inputKey:     "path/to/the/key/the-key",
			expectedPath: "path/to/the/key",
			expectedKey:  "the-key",
			expectedErr:  nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			getter := &Getter{}
			path, key, err := getter.splitKey(tC.inputKey)
			if tC.expectedErr != nil {
				assert.Equal(t, "", path)
				assert.Equal(t, "", key)
				assert.EqualError(t, err, tC.expectedErr.Error())
			} else {
				assert.Equal(t, tC.expectedPath, path)
				assert.Equal(t, tC.expectedKey, key)
				assert.NoError(t, err)
			}
		})
	}
}
