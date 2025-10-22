package redis

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInt(t *testing.T) {
	rdb := initRedisClientForTestingPurpose()
	ctx := context.Background()
	repo := &impl{redis: rdb}

	// Prepare test data
	rdb.Set(ctx, "int_key", "123", 0)
	rdb.Set(ctx, "bad_key", "abc", 0)
	rdb.Del(ctx, "missing_key")

	type testCase struct {
		name         string
		key          string
		expectedVal  int
		expectErrMsg string
	}

	testCases := []testCase{
		{
			name:        "valid int value",
			key:         "int_key",
			expectedVal: 123,
		},
		{
			name:        "missing key",
			key:         "missing_key",
			expectedVal: 0,
		},
		{
			name:         "invalid int value",
			key:          "bad_key",
			expectErrMsg: "invalid int",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := repo.GetInt(ctx, tc.key)

			if tc.expectErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectErrMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedVal, val)
			}
		})
	}
}

func TestGetInt64(t *testing.T) {
	rdb := initRedisClientForTestingPurpose()
	ctx := context.Background()
	repo := &impl{redis: rdb}

	// Setup test data
	rdb.Set(ctx, "int64_key", "922337203685477580", 0)
	rdb.Set(ctx, "bad_int64_key", "not-a-number", 0)
	rdb.Del(ctx, "missing_key")

	type testCase struct {
		name         string
		key          string
		expectedVal  int64
		expectErrMsg string
	}

	testCases := []testCase{
		{
			name:        "valid int64 value",
			key:         "int64_key",
			expectedVal: 922337203685477580,
		},
		{
			name:        "missing key returns zero",
			key:         "missing_key",
			expectedVal: 0,
		},
		{
			name:         "invalid int64 value",
			key:          "bad_int64_key",
			expectErrMsg: "invalid int64",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := repo.GetInt64(ctx, tc.key)

			if tc.expectErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectErrMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedVal, val)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	rdb := initRedisClientForTestingPurpose()
	ctx := context.Background()
	repo := &impl{redis: rdb}

	// Setup test data
	rdb.Set(ctx, "true_key", "true", 0)
	rdb.Set(ctx, "false_key", "false", 0)
	rdb.Set(ctx, "one_key", "1", 0)
	rdb.Set(ctx, "zero_key", "0", 0)
	rdb.Set(ctx, "bad_bool_key", "maybe", 0)
	rdb.Del(ctx, "missing_key")

	type testCase struct {
		name         string
		key          string
		expectedVal  bool
		expectErrMsg string
	}

	testCases := []testCase{
		{
			name:        "true string value",
			key:         "true_key",
			expectedVal: true,
		},
		{
			name:        "false string value",
			key:         "false_key",
			expectedVal: false,
		},
		{
			name:        "numeric 1 as true",
			key:         "one_key",
			expectedVal: true,
		},
		{
			name:        "numeric 0 as false",
			key:         "zero_key",
			expectedVal: false,
		},
		{
			name:        "missing key returns false",
			key:         "missing_key",
			expectedVal: false,
		},
		{
			name:         "invalid bool value",
			key:          "bad_bool_key",
			expectErrMsg: "invalid bool",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := repo.GetBool(ctx, tc.key)

			if tc.expectErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectErrMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedVal, val)
			}
		})
	}
}

func TestGetString(t *testing.T) {
	rdb := initRedisClientForTestingPurpose()
	ctx := context.Background()
	repo := &impl{redis: rdb}

	// Setup test data
	rdb.Set(ctx, "string_key", "hello world", 0)
	rdb.Set(ctx, "empty_string_key", "", 0)
	rdb.Del(ctx, "missing_key")

	type testCase struct {
		name         string
		key          string
		expectedVal  string
		expectErrMsg string
	}

	testCases := []testCase{
		{
			name:        "normal string value",
			key:         "string_key",
			expectedVal: "hello world",
		},
		{
			name:        "empty string value",
			key:         "empty_string_key",
			expectedVal: "",
		},
		{
			name:        "missing key returns empty",
			key:         "missing_key",
			expectedVal: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := repo.GetString(ctx, tc.key)

			if tc.expectErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectErrMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedVal, val)
			}
		})
	}
}

func TestGetBytes(t *testing.T) {
	rdb := initRedisClientForTestingPurpose()
	ctx := context.Background()
	repo := &impl{redis: rdb}

	// Setup test data
	rdb.Set(ctx, "bytes_key", "hello", 0)
	rdb.Set(ctx, "empty_bytes_key", "", 0)
	rdb.Del(ctx, "missing_key")

	type testCase struct {
		name         string
		key          string
		expectedVal  []byte
		expectErr    bool
		expectErrMsg string
	}

	testCases := []testCase{
		{
			name:        "normal string as bytes",
			key:         "bytes_key",
			expectedVal: []byte("hello"),
		},
		{
			name:        "empty string as bytes",
			key:         "empty_bytes_key",
			expectedVal: []byte(""),
		},
		{
			name:        "missing key returns nil",
			key:         "missing_key",
			expectedVal: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := repo.GetBytes(ctx, tc.key)

			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectErrMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedVal, val)
			}
		})
	}
}

func initRedisClientForTestingPurpose() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	return rdb
}
