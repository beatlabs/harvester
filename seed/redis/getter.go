// Package redis handles seeding capabilities with redis.
package redis

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

// Getter definition.
type Getter struct {
	client redis.UniversalClient
}

// New creates a getter.
func New(client redis.UniversalClient) (*Getter, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	return &Getter{client: client}, nil
}

// Get value by key. Returns (nil, 0, nil) when the key does not exist,
// matching the Getter interface contract.
func (g *Getter) Get(key string) (*string, uint64, error) {
	val, err := g.client.Get(context.Background(), key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	return &val, 0, nil
}
