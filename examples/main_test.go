package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmail(t *testing.T) {
	var email Email

	err := email.SetString("invalid")
	require.Error(t, err)
	assert.Empty(t, email.Get())

	err = email.SetString("foo@example.com")
	require.NoError(t, err)
	assert.Equal(t, "foo@example.com", email.Get())
	assert.Equal(t, "foo", email.GetName())
	assert.Equal(t, "example.com", email.GetDomain())
	assert.Equal(t, email.Get(), email.String())
}

func TestConfigString(t *testing.T) {
	var cfg config
	require.NoError(t, cfg.IndexName.SetString("customers-v2"))
	require.NoError(t, cfg.CacheRetention.SetString("86400"))
	require.NoError(t, cfg.LogLevel.SetString("INFO"))
	require.NoError(t, cfg.OpeningBalance.SetString("123.45"))
	require.NoError(t, cfg.AccessToken.SetString("secret"))
	require.NoError(t, cfg.Email.SetString("bar@example.com"))

	got := cfg.String()
	assert.Contains(t, got, "IndexName: customers-v2")
	assert.Contains(t, got, "CacheRetention: 86400")
	assert.Contains(t, got, "LogLevel: INFO")
	assert.Contains(t, got, "OpeningBalance: 123.450000")
	assert.Contains(t, got, "AccessToken: secret")
	assert.Contains(t, got, "Email: bar@example.com")
	assert.True(t, strings.HasPrefix(got, "config:"))
}

func TestCreateRedisClient(t *testing.T) {
	client := createRedisClient()
	require.NotNil(t, client)
	t.Cleanup(func() {
		assert.NoError(t, client.Close())
	})
	assert.NotNil(t, client.Options())
}
