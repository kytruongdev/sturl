package shorturl

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/pkg/testutil"
	redisRepo "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/redis"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdate(t *testing.T) {
	tcs := map[string]struct {
		fixture   string
		shortCode string
		update    model.ShortUrl
		want      model.ShortUrl
		wantErr   bool
	}{
		"success - update status only": {
			fixture:   "testdata/accounts.sql",
			shortCode: "gg123",
			update: model.ShortUrl{
				Status: model.ShortUrlStatusInactive,
			},
			want: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "https://google.com",
				Status:      model.ShortUrlStatusInactive,
			},
			wantErr: false,
		},

		"success - update metadata only": {
			fixture:   "testdata/accounts.sql",
			shortCode: "gg123",
			update: model.ShortUrl{
				Metadata: model.UrlMetadata{
					FinalURL:    "https://google.com",
					Title:       "Google",
					Description: "Search engine",
					Image:       "https://google.com/logo.png",
					Favicon:     "https://google.com/favicon.ico",
				},
			},
			want: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "https://google.com",
				Status:      model.ShortUrlStatusActive,
				Metadata: model.UrlMetadata{
					FinalURL:    "https://google.com",
					Title:       "Google",
					Description: "Search engine",
					Image:       "https://google.com/logo.png",
					Favicon:     "https://google.com/favicon.ico",
				},
			},
			wantErr: false,
		},

		"success - update both status and metadata": {
			fixture:   "testdata/accounts.sql",
			shortCode: "gg123",
			update: model.ShortUrl{
				Status: model.ShortUrlStatusInactive,
				Metadata: model.UrlMetadata{
					FinalURL:    "https://google.com",
					Title:       "Google Search",
					Description: "Search the world",
					Image:       "https://google.com/image.jpg",
					Favicon:     "https://google.com/fav.ico",
				},
			},
			want: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "https://google.com",
				Status:      model.ShortUrlStatusInactive,
				Metadata: model.UrlMetadata{
					FinalURL:    "https://google.com",
					Title:       "Google Search",
					Description: "Search the world",
					Image:       "https://google.com/image.jpg",
					Favicon:     "https://google.com/fav.ico",
				},
			},
			wantErr: false,
		},

		"fail - short code not found": {
			fixture:   "testdata/accounts.sql",
			shortCode: "notfound",
			update: model.ShortUrl{
				Status: model.ShortUrlStatusInactive,
			},
			wantErr: true,
		},

		"success - update with empty status (no status update)": {
			fixture:   "testdata/accounts.sql",
			shortCode: "gg123",
			update: model.ShortUrl{
				Metadata: model.UrlMetadata{
					Title: "Updated Title",
				},
			},
			want: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "https://google.com",
				Status:      model.ShortUrlStatusActive, // Should remain unchanged
				Metadata: model.UrlMetadata{
					Title: "Updated Title",
				},
			},
			wantErr: false,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			testutil.WithTxDB(t, func(tx *sql.Tx) {
				ctx := context.Background()
				testutil.LoadSQLFile(t, tx, tc.fixture)

				// Mock Redis client
				mockRedis := new(redisRepo.MockRedisClient)
				mockRedis.On("GetBytes", mock.Anything, mock.Anything).Return(nil, sql.ErrNoRows)
				mockRedis.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				repo := New(tx, mockRedis)
				err := repo.Update(ctx, tc.update, tc.shortCode)

				if tc.wantErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)

				// Verify the update by fetching the record
				updated, err := repo.GetByShortCode(ctx, tc.shortCode)
				require.NoError(t, err)

				require.True(t,
					cmp.Equal(tc.want, updated, cmpopts.IgnoreFields(model.ShortUrl{}, "CreatedAt", "UpdatedAt")),
					"diff: %v",
					cmp.Diff(tc.want, updated, cmpopts.IgnoreFields(model.ShortUrl{}, "CreatedAt", "UpdatedAt")),
				)
			})
		})
	}
}

