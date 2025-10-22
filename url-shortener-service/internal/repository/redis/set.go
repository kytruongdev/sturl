package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Set sets a key/value pair which expires in Redis. If key already exists, it will override the value.
func (i impl) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd {
	return i.redis.Set(ctx, key, value, ttl)
}
