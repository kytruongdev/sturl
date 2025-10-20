package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient provides the specification of the functionality provided by this pkg
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}
