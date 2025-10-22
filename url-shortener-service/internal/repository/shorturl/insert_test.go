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

func TestInsert(t *testing.T) {
	tcs := map[string]struct {
		fixture             string
		mockCacheKey        string
		setupSetToCacheWant func() *redis.StatusCmd
		given               model.ShortUrl
		want                model.ShortUrl
		wantErr             error
	}{
		"success": {
			fixture:      "testdata/accounts.sql",
			mockCacheKey: fmt.Sprintf("%s%s", cacheKeyShortURL, "fb123"),
			given: model.ShortUrl{
				ShortCode:   "fb123",
				OriginalURL: "https://facebook123.com",
				Status:      "ACTIVE",
			},
			setupSetToCacheWant: func() *redis.StatusCmd {
				return nil
			},
			want: model.ShortUrl{
				ShortCode:   "fb123",
				OriginalURL: "https://facebook123.com",
				Status:      "ACTIVE",
			},
		},
		"success - even set to cache fails": {
			fixture:      "testdata/accounts.sql",
			mockCacheKey: fmt.Sprintf("%s%s", cacheKeyShortURL, "fb123"),
			given: model.ShortUrl{
				ShortCode:   "fb123",
				OriginalURL: "https://facebook123.com",
				Status:      "ACTIVE",
			},
			setupSetToCacheWant: func() *redis.StatusCmd {
				return nil
			},
			want: model.ShortUrl{
				ShortCode:   "fb123",
				OriginalURL: "https://facebook123.com",
				Status:      "ACTIVE",
			},
		},
		"fail - duplicate pkey": {
			fixture:      "testdata/accounts.sql",
			mockCacheKey: fmt.Sprintf("%s%s", cacheKeyShortURL, "fb123"),
			given: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "https://google.com",
				Status:      "ACTIVE",
			},
			setupSetToCacheWant: func() *redis.StatusCmd {
				return nil
			},
			wantErr: errors.New("orm: unable to insert into short_urls: pq: duplicate key value violates unique constraint \"short_urls_pkey\""),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			testutil.WithTxDB(t, func(tx *sql.Tx) {
				ctx := context.Background()
				testutil.LoadSQLFile(t, tx, tc.fixture)

				setToCacheWant := tc.setupSetToCacheWant()
				redisClient := new(redisRepo.MockRedisClient)
				redisClient.ExpectedCalls = []*mock.Call{
					redisClient.On("Set", mock.Anything, tc.mockCacheKey, mock.Anything, cacheShortURLTTL).Return(setToCacheWant),
				}

				repo := New(tx, redisClient)
				actual, err := repo.Insert(ctx, tc.given)
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
