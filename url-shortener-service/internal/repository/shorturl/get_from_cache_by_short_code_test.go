package shorturl

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	redisRepo "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/redis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetByShortCodeFromCache(t *testing.T) {
	type testCase struct {
		name        string
		shortCode   string
		setupMock   func(m *redisRepo.MockRedisClient, key string)
		expected    model.ShortUrl
		expectedErr bool
	}

	tests := []testCase{
		{
			name:      "success - cache hit and valid JSON",
			shortCode: "abc123",
			setupMock: func(m *redisRepo.MockRedisClient, key string) {
				mockData := model.ShortUrl{
					ShortCode:   "abc123",
					OriginalURL: "https://example.com",
				}
				jsonVal, _ := json.Marshal(mockData)
				cmd := redis.NewStringResult(string(jsonVal), nil)
				m.On("Get", mock.Anything, key).Return(cmd)
			},
			expected: model.ShortUrl{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
			},
			expectedErr: false,
		},
		{
			name:      "fail - redis.Nil (cache miss)",
			shortCode: "missing",
			setupMock: func(m *redisRepo.MockRedisClient, key string) {
				cmd := redis.NewStringResult("", redis.Nil)
				m.On("Get", mock.Anything, key).Return(cmd)
			},
			expectedErr: true,
		},
		{
			name:      "fail - redis connection error",
			shortCode: "err001",
			setupMock: func(m *redisRepo.MockRedisClient, key string) {
				cmd := redis.NewStringResult("", fmt.Errorf("connection lost"))
				m.On("Get", mock.Anything, key).Return(cmd)
			},
			expectedErr: true,
		},
		{
			name:      "fail - invalid JSON in cache",
			shortCode: "badjson",
			setupMock: func(m *redisRepo.MockRedisClient, key string) {
				cmd := redis.NewStringResult("not a json", nil)
				m.On("Get", mock.Anything, key).Return(cmd)
			},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRedis := new(redisRepo.MockRedisClient)
			repo := impl{redisClient: mockRedis}

			key := fmt.Sprintf("%s:%s", cacheKeyShortURL, tc.shortCode)
			tc.setupMock(mockRedis, key)

			result, err := repo.GetByShortCodeFromCache(context.Background(), tc.shortCode)

			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, result)
			}

			mockRedis.AssertExpectations(t)
		})
	}
}
