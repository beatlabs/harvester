// +build integration

package vault

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	addr    = "http://0.0.0.0:8200"
	token   = "root"
	timeout = 5 * time.Second
)

func Test_Get_From_Vault(t *testing.T) {
	client, err := newClient(addr, token, timeout)
	require.NoError(t, err)

	basePath := "secret/data/harvester/test/getter"
	deleteSecretsAtPath(t, client, basePath)

	getter := newGetter(t)

	writeSecret(t, client, basePath, map[string]interface{}{
		"string":        "foo",
		"integer":       42,
		"float":         13.37,
		"boolean-true":  true,
		"boolean-false": false,
		"slice":         []interface{}{"foo", "bar"},
	})

	testCases := []struct {
		desc          string
		key           string
		expectedValue *string
		expectedErr   error
	}{
		{
			desc:        "Malformed key",
			key:         "/something",
			expectedErr: errors.New("malformed key: /something"),
		},
		{
			desc:          "Unexisting key",
			key:           "something/that/does/not/exist",
			expectedValue: nil,
			expectedErr:   nil,
		},
		{
			desc:          "Unexisting secret within key",
			key:           fmt.Sprintf("%s/something", basePath),
			expectedValue: nil,
			expectedErr:   fmt.Errorf("no Vault secret could be found at path %s with key something", basePath),
		},
		{
			desc:          "Unsupported data type",
			key:           fmt.Sprintf("%s/slice", basePath),
			expectedValue: nil,
			expectedErr:   fmt.Errorf("unsupported data type stored in Vault at path %s with key slice ([]interface {} found)", basePath),
		},
		{
			desc:          "String value",
			key:           fmt.Sprintf("%s/string", basePath),
			expectedValue: pointerToString("foo"),
		},
		{
			desc:          "Integer value",
			key:           fmt.Sprintf("%s/integer", basePath),
			expectedValue: pointerToString("42"),
		},
		{
			desc:          "Float value",
			key:           fmt.Sprintf("%s/float", basePath),
			expectedValue: pointerToString("13.37"),
		},
		{
			desc:          "Boolean value (true)",
			key:           fmt.Sprintf("%s/boolean-true", basePath),
			expectedValue: pointerToString("true"),
		},
		{
			desc:          "Boolean value (false)",
			key:           fmt.Sprintf("%s/boolean-false", basePath),
			expectedValue: pointerToString("false"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			foundValue, version, err := getter.Get(tC.key)

			if tC.expectedErr != nil {
				assert.Nil(t, foundValue)
				assert.Equal(t, uint64(0), version)
				assert.EqualError(t, err, tC.expectedErr.Error())
			} else {
				assert.Equal(t, tC.expectedValue, foundValue)
				assert.Equal(t, uint64(0), version)
				assert.NoError(t, err)
			}
		})
	}
}

func newGetter(t *testing.T) *Getter {
	// TODO: this should move to env-based vars...
	getter, err := New(addr, token, timeout)
	require.NoError(t, err)
	return getter
}

func deleteSecretsAtPath(t *testing.T, client *api.Client, path string) {
	_, err := client.Logical().Delete(path)
	require.NoError(t, err)
}

func writeSecret(t *testing.T, client *api.Client, path string, data map[string]interface{}) {
	_, err := client.Logical().Write(
		path,
		map[string]interface{}{
			"data": data,
		},
	)
	require.NoError(t, err)
}

func pointerToString(s string) *string {
	return &s
}
