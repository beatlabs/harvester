// Package redis handles seeding capabilities with redis.
package redis

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
)

// Getter definition.
type Getter struct {
	client *redis.Client
}

// New creates a getter.
func New(client *redis.Client) (*Getter, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	return &Getter{client: client}, nil
}

// Get value by key.
func (g Getter) Get(key string) (*string, uint64, error) {
	val, err := g.client.Get(context.Background(), key).Result()
	if err != nil {
		return nil, 0, err
	}
	return &val, 0, nil
}
