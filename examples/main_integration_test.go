//go:build integration
// +build integration

package main

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExampleHelpersIntegration(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, os.Unsetenv("ENV_CACHE_RETENTION_SECONDS"))
	})

	setEnvVarCacheRetention()
	assert.Equal(t, "86400", os.Getenv("ENV_CACHE_RETENTION_SECONDS"))

	seedConsulAccessToken("integration-token")
	cl, err := api.NewClient(api.DefaultConfig())
	require.NoError(t, err)
	pair, _, err := cl.KV().Get("harvester/example/accesstoken", nil)
	require.NoError(t, err)
	require.NotNil(t, pair)
	assert.Equal(t, "integration-token", string(pair.Value))

	err = setRedisOpeningBalance(context.Background(), "1234.56")
	require.NoError(t, err)
	client := createRedisClient()
	t.Cleanup(func() {
		assert.NoError(t, client.Close())
	})
	got, err := client.Get(context.Background(), "opening-balance").Result()
	require.NoError(t, err)
	assert.Equal(t, "1234.56", got)
}

func TestMainIntegration(t *testing.T) {
	main()
}
