package public

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestShorten(t *testing.T) {
	type mockCtrl struct {
		inp    shorturl.ShortenInput
		output model.ShortUrl
		err    error
	}

	type testCase struct {
		requestBody string
		mockCtrl    mockCtrl
		wantCode    int
		wantResp    string
		wantErr     *httpserver.Error
	}

	tcs := map[string]testCase{
		"success": {
			requestBody: `{"original_url": "https://google.com"}`,
			mockCtrl: mockCtrl{
				inp: shorturl.ShortenInput{
					OriginalURL: "https://google.com",
				},
				output: model.ShortUrl{
					ShortCode:   "abc123",
					OriginalURL: "https://google.com",
					Status:      model.ShortUrlStatusActive,
					CreatedAt:   time.Date(2025, 10, 20, 0, 0, 0, 0, time.UTC),
					UpdatedAt:   time.Date(2025, 10, 20, 0, 0, 0, 0, time.UTC),
				},
			},
			wantCode: http.StatusOK,
			wantResp: `{"short_code":"abc123","original_url":"https://google.com","status":"ACTIVE","created_at":"2025-10-20T00:00:00Z","updated_at":"2025-10-20T00:00:00Z"}`,
		},
		"empty original_url": {
			requestBody: `{}`,
			mockCtrl:    mockCtrl{},
			wantCode:    http.StatusBadRequest,
			wantErr:     WebErrEmptyOriginalURL,
		},
		"invalid URL format": {
			requestBody: `{"original_url": "not-a-url"}`,
			mockCtrl:    mockCtrl{},
			wantCode:    http.StatusBadRequest,
			wantErr:     WebErrInvalidOriginalURL,
		},
		"controller returns error": {
			requestBody: `{"original_url": "https://google.com"}`,
			mockCtrl: mockCtrl{
				inp: shorturl.ShortenInput{
					OriginalURL: "https://google.com",
				},
				output: model.ShortUrl{}, err: errors.New("error")},
			wantCode: http.StatusInternalServerError,
			wantErr:  httpserver.ErrDefaultInternal,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/public/v1/shorten", strings.NewReader(tc.requestBody))
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, chi.NewRouteContext())
			rec := httptest.NewRecorder()
			req = req.WithContext(ctx)

			ctrl := new(shorturl.MockController)
			ctrl.ExpectedCalls = []*mock.Call{
				ctrl.On("Shorten", mock.Anything, tc.mockCtrl.inp).Return(tc.mockCtrl.output, tc.mockCtrl.err),
			}

			handler := Handler{shortUrlCtrl: ctrl}
			handler.Shorten().ServeHTTP(rec, req)
			require.Equal(t, tc.wantCode, rec.Code)

			var actErr httpserver.Error
			err := httpserver.ParseJSON(rec.Result().Body, &actErr)
			if tc.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.wantErr.Code, actErr.Code)
				require.Equal(t, tc.wantErr.Desc, actErr.Desc)
			} else {
				require.Nil(t, err)
				require.Empty(t, actErr.Code)
				require.Empty(t, actErr.Desc)
				require.Equal(t, tc.wantResp, rec.Body.String())
			}
		})
	}
}
