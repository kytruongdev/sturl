package shorturl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/id"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/outgoingevent"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/shorturl"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestShorten(t *testing.T) {
	tcs := map[string]struct {
		inp                      ShortenInput
		mockGetByOriginalURLWant model.ShortUrl
		mockGetByOriginalURLErr  error
		mockGenShortCode         func(int) string
		mockInsertShortURLWant   model.ShortUrl
		mockInsertShortURLErr    error
		mockInsertOutboxWant     model.OutgoingEvent
		mockInsertOutboxErr      error
		want                     model.ShortUrl
		wantErr                  error
	}{
		"success - existing url": {
			inp: ShortenInput{OriginalURL: "http://google.com"},
			mockGetByOriginalURLWant: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "http://google.com",
				Status:      model.ShortUrlStatusActive,
			},
			mockGetByOriginalURLErr: nil,
			want: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "http://google.com",
				Status:      model.ShortUrlStatusActive,
			},
		},

		"success - new shortened url": {
			inp:                     ShortenInput{OriginalURL: "http://google.com"},
			mockGetByOriginalURLErr: shorturl.ErrNotFound,
			mockGenShortCode:        func(_ int) string { return "gg123" },
			mockInsertShortURLWant: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "http://google.com",
				Status:      model.ShortUrlStatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},

			mockInsertOutboxWant: model.OutgoingEvent{},
			want: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "http://google.com",
				Status:      model.ShortUrlStatusActive,
			},
		},

		"fail - getByOriginal returns error": {
			inp:                     ShortenInput{OriginalURL: "http://google.com"},
			mockGetByOriginalURLErr: errors.New("db error"),
			wantErr:                 errors.New("db error"),
		},

		"fail - insert shorturl fails": {
			inp:                     ShortenInput{OriginalURL: "http://google.com"},
			mockGetByOriginalURLErr: shorturl.ErrNotFound,
			mockGenShortCode:        func(_ int) string { return "gg123" },
			mockInsertShortURLErr:   errors.New("insert shorturl failed"),
			wantErr:                 errors.New("insert shorturl failed"),
		},

		"fail - insert outbox event fails": {
			inp:                     ShortenInput{OriginalURL: "http://google.com"},
			mockGetByOriginalURLErr: shorturl.ErrNotFound,
			mockGenShortCode:        func(_ int) string { return "gg123" },
			mockInsertShortURLWant: model.ShortUrl{
				ShortCode:   "gg123",
				OriginalURL: "http://google.com",
				Status:      model.ShortUrlStatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},

			mockInsertOutboxErr: errors.New("outbox insert failed"),
			wantErr:             errors.New("outbox insert failed"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			newIDFunc = func() int64 {
				return 123
			}

			defer func() { newIDFunc = id.New }()

			// Mock slug generator
			if tc.mockGenShortCode != nil {
				generateShortCodeFunc = tc.mockGenShortCode
			} else {
				generateShortCodeFunc = func(n int) string { return "ignored" }
			}
			defer func() { generateShortCodeFunc = generateShortCode }()

			// Mock ShortURL repo
			mockShort := new(shorturl.MockRepository)
			mockShort.On("GetByOriginalURL", mock.Anything, tc.inp.OriginalURL).
				Return(tc.mockGetByOriginalURLWant, tc.mockGetByOriginalURLErr)

			if tc.mockGetByOriginalURLErr == shorturl.ErrNotFound {
				mockShort.On("Insert", mock.Anything, mock.Anything).
					Return(tc.mockInsertShortURLWant, tc.mockInsertShortURLErr)
			}

			// Mock Outbox repo
			mockOutbox := new(outgoingevent.MockRepository)
			if tc.mockGetByOriginalURLErr == shorturl.ErrNotFound && tc.mockInsertShortURLErr == nil {
				mockOutbox.On("Insert", mock.Anything, mock.Anything).
					Return(tc.mockInsertOutboxWant, tc.mockInsertOutboxErr)
			}

			// Mock Registry
			mockReg := new(repository.MockRegistry)

			mockReg.On("ShortUrl").Return(mockShort)
			mockReg.On("OutgoingEvent").Return(mockOutbox)

			// Fake DoInTx: simply run fn
			mockReg.On("DoInTx", mock.Anything, mock.Anything, mock.Anything).
				Return(func(ctx context.Context, _ backoff.BackOff, fn func(context.Context, repository.Registry) error) error {
					return fn(ctx, mockReg)
				})

			i := New(mockReg)

			actual, err := i.Shorten(ctx, tc.inp)

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
	}
}
