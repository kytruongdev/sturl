package redis

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	rdb := initRedisClientForTestingPurpose()
	ctx := context.Background()
	repo := &impl{redis: rdb}

	type testCase struct {
		key         string
		value       interface{}
		ttl         time.Duration
		expectValue string
	}

	tcs := map[string]testCase{
		"set string value": {
			key:         "set:string",
			value:       "hello world",
			ttl:         5 * time.Second,
			expectValue: "hello world",
		},
		"set int value (auto stringify)": {
			key:         "set:int",
			value:       42,
			ttl:         5 * time.Second,
			expectValue: "42",
		},
		"set JSON struct": {
			key: "set:json",
			value: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{"Alice", 30},
			ttl:         5 * time.Second,
			expectValue: `{"name":"Alice","age":30}`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			val := tc.value
			if v := reflect.ValueOf(val); v.Kind() == reflect.Struct {
				bs, err := json.Marshal(val)
				require.NoError(t, err)
				val = bs
			}

			cmd := repo.Set(ctx, tc.key, val, tc.ttl)
			require.NoError(t, cmd.Err())

			got, err := rdb.Get(ctx, tc.key).Result()
			require.NoError(t, err)
			assert.Equal(t, tc.expectValue, got)
		})
	}
}
