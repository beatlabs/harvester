// +build integration

package consul

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/taxibeat/harvester"
)

func TestPayload(t *testing.T) {

	params := make(map[string]interface{})
	params["datacenter"] = ""
	params["token"] = ""
	params["key"] = "test/test"
	params["type"] = "key"

	ch := make(chan *harvester.Change, 0)
	chErr := make(chan error, 0)

	w, err := New("127.0.0.1:8500", params, ch, chErr, false)
	require.NoError(t, err)
	require.NotNil(t, w)

}
