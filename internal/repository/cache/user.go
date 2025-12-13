package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"connectify/internal/domain"

	"github.com/redis/go-redis/v9"
)

var ErrKeyNotExist = redis.Nil

// UserCache is the user cache
type UserCache struct {
	client redis.Cmdable
	expire time.Duration
}

// NewUserCache creates a user cache instance
func NewUserCache(client redis.Cmdable) *UserCache {
	return &UserCache{
		client: client,
		expire: time.Minute * 10,
	}
}

// Get gets user information from the cache
func (c *UserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := c.key(id)
	data, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return domain.User{}, ErrKeyNotExist
	}
	if err != nil {
		return domain.User{}, err
	}

	var user domain.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return domain.User{}, err
	}

	return user, nil
}

// Set writes user information to the cache
func (c *UserCache) Set(ctx context.Context, user domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(user.Id), data, c.expire).Err()
}

// key generates the Redis key for the user cache
func (c *UserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
