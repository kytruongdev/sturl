package shorturl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/pkg/testutil"
	redisRepo "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/redis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetByShortCode(t *testing.T) {
	tcs := map[string]struct {
		fixture             string
		shortCode           string
		mockCacheKey        string
		mockDataForCache    *model.ShortUrl
		mockGetBytesWantErr error
		setupSetToCacheWant func() *redis.StatusCmd
		want                model.ShortUrl
		wantErr             error
	}{
		"success - found in cache": {
			fixture:             "testdata/accounts.sql",
			shortCode:           "gg123",
			mockCacheKey:        fmt.Sprintf("%s%s", cacheKeyShortURL, "gg123"),
			mockGetBytesWantErr: nil,
			mockDataForCache: &model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "https://google.com",
				Status:      "ACTIVE",
			},
			setupSetToCacheWant: func() *redis.StatusCmd {
				return nil
			},
			want: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "https://google.com",
				Status:      "ACTIVE",
			},
		},
		"success - found in database": {
			fixture:             "testdata/accounts.sql",
			shortCode:           "gg123",
			mockCacheKey:        fmt.Sprintf("%s%s", cacheKeyShortURL, "gg123"),
			mockGetBytesWantErr: errors.New("cache miss"),
			mockDataForCache:    nil,
			setupSetToCacheWant: func() *redis.StatusCmd {
				return nil
			},
			want: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "https://google.com",
				Status:      "ACTIVE",
			},
		},
		"success - even set to cache fails": {
			fixture:             "testdata/accounts.sql",
			shortCode:           "gg123",
			mockCacheKey:        fmt.Sprintf("%s%s", cacheKeyShortURL, "gg123"),
			mockGetBytesWantErr: errors.New("cache miss"),
			mockDataForCache:    nil,
			setupSetToCacheWant: func() *redis.StatusCmd {
				stt := &redis.StatusCmd{}
				stt.SetErr(errors.New("set to cache error"))
				return stt
			},
			want: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "https://google.com",
				Status:      "ACTIVE",
			},
		},
		"not found in both cache and database": {
			fixture:             "testdata/accounts.sql",
			shortCode:           "404",
			mockCacheKey:        fmt.Sprintf("%s%s", cacheKeyShortURL, "404"),
			mockGetBytesWantErr: errors.New("cache miss"),
			setupSetToCacheWant: func() *redis.StatusCmd {
				return nil
			},
			mockDataForCache: nil,
			wantErr:          ErrNotFound,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			testutil.WithTxDB(t, func(tx *sql.Tx) {
				ctx := context.Background()
				testutil.LoadSQLFile(t, tx, tc.fixture)

				getBytesWant := marshalObjectForTesting(t, tc.mockDataForCache)
				setToCacheWant := tc.setupSetToCacheWant()
				redisClient := new(redisRepo.MockRedisClient)
				redisClient.ExpectedCalls = []*mock.Call{
					redisClient.On("GetBytes", mock.Anything, tc.mockCacheKey).Return(getBytesWant, tc.mockGetBytesWantErr),
					redisClient.On("Set", mock.Anything, tc.mockCacheKey, mock.Anything, cacheShortURLTTL).Return(setToCacheWant),
				}

				repo := New(tx, redisClient)
				actual, err := repo.GetByShortCode(ctx, tc.shortCode)
				if tc.wantErr != nil {
					require.ErrorContains(t, err, tc.wantErr.Error())
					return
				}

				require.NoError(t, err)
				require.True(t,
					cmp.Equal(tc.want, actual, cmpopts.IgnoreFields(model.ShortUrl{}, "CreatedAt", "UpdatedAt")),
					"diff: %v",
					cmp.Diff(tc.want, actual, cmpopts.IgnoreFields(model.ShortUrl{}, "CreatedAt", "UpdatedAt")),
				)
			})
		})
	}
}
