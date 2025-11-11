package shorturl

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/shorturl"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRetrieve(t *testing.T) {
	tcs := map[string]struct {
		shortCode                string
		mockGetByShortCodeResult model.ShortUrl
		mockGetByShortCodeErr    error
		want                     model.ShortUrl
		wantErr                  error
	}{
		"success": {
			shortCode: "abc",
			mockGetByShortCodeResult: model.ShortUrl{
				ShortCode:   "abc",
				OriginalURL: "https://abc.com/123",
				Status:      model.ShortUrlStatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockGetByShortCodeErr: nil,
			want: model.ShortUrl{
				ShortCode:   "abc",
				OriginalURL: "https://abc.com/123",
				Status:      model.ShortUrlStatusActive,
			},
		},
		"fail - URL is inactive": {
			shortCode: "abc",
			mockGetByShortCodeResult: model.ShortUrl{
				ShortCode:   "abc",
				OriginalURL: "https://abc.com/123",
				Status:      model.ShortUrlStatusInactive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockGetByShortCodeErr: nil,
			wantErr:               ErrInactiveURL,
		},
		"not found": {
			shortCode:             "abc",
			mockGetByShortCodeErr: shorturl.ErrNotFound,
			wantErr:               ErrURLNotfound,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockShortURLRepo := new(shorturl.MockRepository)
			mockShortURLRepo.ExpectedCalls = []*mock.Call{
				mockShortURLRepo.On("GetByShortCode", mock.Anything, tc.shortCode).Return(tc.mockGetByShortCodeResult, tc.mockGetByShortCodeErr),
			}

			i := New(mockShortURLRepo)
			actual, err := i.Retrieve(ctx, tc.shortCode)
			if tc.wantErr != nil {
				require.EqualError(t, err, tc.wantErr.Error())
			} else {
				require.NoError(t, err)
				require.True(t,
					cmp.Equal(tc.want, actual, cmpopts.IgnoreFields(model.ShortUrl{}, "CreatedAt", "UpdatedAt")),
					"diff: %v",
					cmp.Diff(tc.want, actual, cmpopts.IgnoreFields(model.ShortUrl{}, "CreatedAt", "UpdatedAt")),
				)
			}
		})
	}
}
