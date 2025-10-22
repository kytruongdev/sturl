package redis

import (
	"context"
	"time"

	pkgerrors "github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// RedisClient provides the specification of the functionality provided by this pkg
type RedisClient interface {
	GetInt(ctx context.Context, key string) (int, error)
	GetInt64(ctx context.Context, key string) (int64, error)
	GetBool(ctx context.Context, key string) (bool, error)
	GetString(ctx context.Context, key string) (string, error)
	GetBytes(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd
}
type impl struct {
	redis *redis.Client
}

func NewRedisClient(ctx context.Context, cfg *redis.Options) (RedisClient, error) {
	rbd := redis.NewClient(cfg)

	if err := rbd.Ping(context.Background()).Err(); err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	return impl{
		redis: rbd,
	}, nil
}
