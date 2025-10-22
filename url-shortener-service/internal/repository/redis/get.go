package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	pkgerrors "github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// GetInt gets a int value for the key from Redis
func (i impl) GetInt(ctx context.Context, key string) (int, error) {
	return getValue[int](ctx, i.redis, key)
}

// GetInt64 gets a int64 value for the key from Redis
func (i impl) GetInt64(ctx context.Context, key string) (int64, error) {
	return getValue[int64](ctx, i.redis, key)
}

// GetBool gets a bool value for the key from Redis
func (i impl) GetBool(ctx context.Context, key string) (bool, error) {
	return getValue[bool](ctx, i.redis, key)
}

// GetString gets a string value for the key from Redis
func (i impl) GetString(ctx context.Context, key string) (string, error) {
	return getValue[string](ctx, i.redis, key)
}

// GetBytes gets an array of byte value for the key from Redis
func (i impl) GetBytes(ctx context.Context, key string) ([]byte, error) {
	return getValue[[]byte](ctx, i.redis, key)
}

func getValue[T any](ctx context.Context, c *redis.Client, key string) (T, error) {
	var zero T

	val, err := c.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return zero, nil
		}
		return zero, pkgerrors.WithStack(err)
	}

	var result any
	switch any(zero).(type) {
	case []byte:
		result = []byte(val)
	case string:
		result = val
	case int:
		i, err := strconv.Atoi(val)
		if err != nil {
			return zero, fmt.Errorf("invalid int: %w", err)
		}
		result = i
	case int64:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return zero, fmt.Errorf("invalid int64: %w", err)
		}
		result = i
	case float64:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return zero, fmt.Errorf("invalid float64: %w", err)
		}
		result = f
	case bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return zero, fmt.Errorf("invalid bool: %w", err)
		}
		result = b
	default:
		if err := json.Unmarshal([]byte(val), &zero); err != nil {
			return zero, fmt.Errorf("unsupported type or invalid JSON: %w", err)
		}
		return zero, nil
	}

	casted, ok := result.(T)
	if !ok {
		return zero, fmt.Errorf("type assertion failed: expected %T but got %T", zero, result)
	}

	return casted, nil
}
