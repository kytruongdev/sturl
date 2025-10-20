package shorturl

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func TestGetByOriginalURL(t *testing.T) {
	tcs := map[string]struct {
		fixture     string
		originalURL string
		want        model.ShortUrl
		wantErr     error
	}{
		"success": {
			fixture:     "testdata/accounts.sql",
			originalURL: "https://google.com",
			want: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "https://google.com",
				Status:      "ACTIVE",
			},
		},
		"not found": {
			fixture:     "testdata/accounts.sql",
			originalURL: "https://404.notfound",
			wantErr:     ErrNotFound,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			testutil.WithTxDB(t, func(tx *sql.Tx) {
				ctx := context.Background()

				testutil.LoadSQLFile(t, tx, tc.fixture)
				repo := New(tx, nil)
				actual, err := repo.GetByOriginalURL(ctx, tc.originalURL)
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
