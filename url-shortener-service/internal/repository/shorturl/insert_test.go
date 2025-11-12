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
	"github.com/stretchr/testify/require"
)

func TestInsert(t *testing.T) {
	tcs := map[string]struct {
		fixture      string
		mockCacheKey string
		given        model.ShortUrl
		want         model.ShortUrl
		wantErr      error
	}{
		"success": {
			fixture:      "testdata/accounts.sql",
			mockCacheKey: fmt.Sprintf("%s%s", cacheKeyShortURL, "fb123"),
			given: model.ShortUrl{
				ShortCode:   "fb123",
				OriginalURL: "https://facebook123.com",
				Status:      "ACTIVE",
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
			wantErr: errors.New("orm: unable to insert into short_urls: ERROR: duplicate key value violates unique constraint \"short_urls_pkey\" (SQLSTATE 23505)"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			testutil.WithTxDB(t, func(tx *sql.Tx) {
				ctx := context.Background()
				testutil.LoadSQLFile(t, tx, tc.fixture)
				repo := New(tx, new(redisRepo.MockRedisClient))
				actual, err := repo.Insert(ctx, tc.given)
				if tc.wantErr != nil {
					require.EqualError(t, err, tc.wantErr.Error())
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
