package shorturl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/shorturl"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestShorten(t *testing.T) {
	tcs := map[string]struct {
		inp                       ShortenInput
		mockGetByOriginalURLInput string
		mockGetByOriginalURLWant  model.ShortUrl
		mockGetByOriginalURLErr   error
		mockGenShortCode          func(int) string
		mockInsertInput           model.ShortUrl
		mockInsertWant            model.ShortUrl
		mockInsertErr             error
		want                      model.ShortUrl
		wantErr                   error
	}{
		"success - existing url": {
			mockGetByOriginalURLInput: "http://google.com",
			inp: ShortenInput{
				OriginalURL: "http://google.com",
			},
			mockGetByOriginalURLWant: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "http://google.com",
				Status:      model.ShortUrlStatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			want: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "http://google.com",
				Status:      model.ShortUrlStatusActive,
			},
		},
		"success - shortened url": {
			mockGetByOriginalURLInput: "http://google.com",
			inp: ShortenInput{
				OriginalURL: "http://google.com",
			},
			mockGetByOriginalURLErr: shorturl.ErrNotFound,
			mockGenShortCode: func(originalURL int) string {
				return "gg123"
			},
			mockInsertInput: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "http://google.com",
				Status:      model.ShortUrlStatusActive,
			},
			mockInsertWant: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "http://google.com",
				Status:      model.ShortUrlStatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			want: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "http://google.com",
				Status:      model.ShortUrlStatusActive,
			},
		},
		"fail - get by original url return error": {
			mockGetByOriginalURLInput: "http://google.com",
			inp: ShortenInput{
				OriginalURL: "http://google.com",
			},
			mockGetByOriginalURLErr: errors.New("some error"),
			wantErr:                 errors.New("some error"),
		},
		"fail - duplicated short code": {
			mockGetByOriginalURLInput: "http://google.com",
			inp: ShortenInput{
				OriginalURL: "http://google.com",
			},
			mockGetByOriginalURLErr: shorturl.ErrNotFound,
			mockGenShortCode: func(originalURL int) string {
				return "gg123"
			},
			mockInsertInput: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "http://google.com",
				Status:      model.ShortUrlStatusActive,
			},
			mockInsertErr: errors.New("duplicated short code"),
			wantErr:       errors.New("duplicated short code"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			generateShortCodeFunc = tc.mockGenShortCode
			defer func() { generateShortCodeFunc = generateShortCode }()

			mockShortURLRepo := new(shorturl.MockRepository)
			mockShortURLRepo.ExpectedCalls = []*mock.Call{
				mockShortURLRepo.On("GetByOriginalURL", mock.Anything, tc.mockGetByOriginalURLInput).Return(tc.mockGetByOriginalURLWant, tc.mockGetByOriginalURLErr),
				mockShortURLRepo.On("Insert", mock.Anything, tc.mockInsertInput).Return(tc.mockInsertWant, tc.mockInsertErr),
			}

			i := New(mockShortURLRepo)
			actual, err := i.Shorten(ctx, tc.inp)
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
