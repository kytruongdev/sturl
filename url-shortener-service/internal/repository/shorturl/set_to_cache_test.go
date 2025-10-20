package shorturl

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	redisRepo "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/redis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSetToCache(t *testing.T) {
	type testCase struct {
		name        string
		input       model.ShortUrl
		setupMock   func(m *redisRepo.MockRedisClient, jsonVal []byte)
		expectedErr bool
	}

	tcs := []testCase{
		{
			name: "success",
			input: model.ShortUrl{
				ShortCode:   "abc123",
				OriginalURL: "https://example.com",
			},
			setupMock: func(m *redisRepo.MockRedisClient, jsonVal []byte) {
				key := "short_url::abc123"
				statusCmd := redis.NewStatusResult("OK", nil)
				m.On("Set", mock.Anything, key, jsonVal, 24*time.Hour).Return(statusCmd)
			},
			expectedErr: false,
		},
		{
			name: "redis set fail",
			input: model.ShortUrl{
				ShortCode:   "abc456",
				OriginalURL: "https://fail.com",
			},
			setupMock: func(m *redisRepo.MockRedisClient, jsonVal []byte) {
				key := "short_url::abc456"
				statusCmd := redis.NewStatusResult("", fmt.Errorf("connection lost"))
				m.On("Set", mock.Anything, key, jsonVal, 24*time.Hour).Return(statusCmd)
			},
			expectedErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			mockRedis := new(redisRepo.MockRedisClient)
			repo := impl{redisClient: mockRedis}

			jsonVal, _ := json.Marshal(tc.input)
			tc.setupMock(mockRedis, jsonVal)

			err := repo.SetToCache(context.Background(), tc.input)

			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockRedis.AssertExpectations(t)
		})
	}
}
