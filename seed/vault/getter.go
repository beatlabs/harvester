package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
)

// Getter implementation of the getter interface.
type Getter struct {
	client *api.Client
}

// New constructor. Timeout is set to 60s when 0 is provided
func New(addr, token string, timeout time.Duration) (*Getter, error) {
	client, err := newClient(addr, token, timeout)
	if err != nil {
		return nil, err
	}
	return &Getter{client: client}, nil
}

func newClient(addr, token string, timeout time.Duration) (*api.Client, error) {
	if addr == "" {
		return nil, errors.New("address is empty")
	}
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	config := api.DefaultConfig()
	config.Address = addr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	client.SetToken(token)

	return client, nil
}

// Get the specific key value from Vault.
func (g *Getter) Get(key string) (*string, uint64, error) {
	path, secretKey, err := g.splitKey(key)
	if err != nil {
		return nil, 0, err
	}

	secret, err := g.client.Logical().Read(path)
	if err != nil {
		return nil, 0, err
	}
	if secret == nil {
		return nil, 0, nil
	}

	data, ok := secret.Data["data"]
	if !ok {
		return nil, 0, fmt.Errorf("could not fetch Vault secret data at path: %s", path)
	}

	secrets, ok := data.(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("invalid data stored in Vault at path: %s", path)
	}

	secretValue, ok := secrets[secretKey]
	if !ok {
		return nil, 0, fmt.Errorf("no Vault secret could be found at path %s with key %s", path, secretKey)
	}

	var secretValueString string

	switch value := secretValue.(type) {
	case json.Number:
		secretValueString = value.String()
	case string:
		secretValueString = value
	case bool:
		if value {
			secretValueString = "true"
		} else {
			secretValueString = "false"
		}
	default:
		return nil, 0, fmt.Errorf("unsupported data type stored in Vault at path %s with key %s (%T found)", path, secretKey, value)
	}

	return &secretValueString, 0, nil
}

// splitKey splits a given key and returns the Vault path plus the key inside the path.
// Example: splitKey("/path/to/the/key/the-key") will return ("path/to/the/key", "the-key", nil)
// Returns an error if the key cannot be split.
func (g *Getter) splitKey(key string) (string, string, error) {
	parts := strings.Split(
		strings.Trim(key, "/"),
		"/",
	)

	if len(parts) <= 1 {
		return "", "", fmt.Errorf("malformed key: %s", key)
	}

	return strings.Join(parts[:len(parts)-1], "/"), parts[len(parts)-1], nil
}
