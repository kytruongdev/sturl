package shorturl

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func TestInsert(t *testing.T) {
	tcs := map[string]struct {
		fixture string
		given   model.ShortUrl
		want    model.ShortUrl
		wantErr error
	}{
		"success": {
			fixture: "testdata/accounts.sql",
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
		"2fail - duplicate pkey": {
			fixture: "testdata/accounts.sql",
			given: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "https://google.com",
				Status:      "ACTIVE",
			},
			wantErr: errors.New("orm: unable to insert into short_urls: pq: duplicate key value violates unique constraint \"short_urls_pkey\""),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			testutil.WithTxDB(t, func(tx *sql.Tx) {
				ctx := context.Background()

				testutil.LoadSQLFile(t, tx, tc.fixture)
				repo := New(tx, nil)
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
