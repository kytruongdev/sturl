package public

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRedirect(t *testing.T) {
	type mockCtrl struct {
		inp    string
		output model.ShortUrl
		err    error
	}

	tcs := map[string]struct {
		mockCtrl  mockCtrl
		shortCode string
		wantCode  int
		wantErr   *httpserver.Error
	}{
		"success": {
			shortCode: "gg",
			mockCtrl: mockCtrl{
				inp: "gg",
				output: model.ShortUrl{
					ShortCode:   "gg",
					OriginalURL: "https://google.com",
					Status:      model.ShortUrlStatusActive,
					CreatedAt:   time.Date(2025, 10, 20, 0, 0, 0, 0, time.UTC),
					UpdatedAt:   time.Date(2025, 10, 20, 0, 0, 0, 0, time.UTC),
				},
			},
			wantCode: http.StatusMovedPermanently,
		},
		"fail - empty short code": {
			shortCode: "",
			mockCtrl:  mockCtrl{},
			wantCode:  http.StatusBadRequest,
			wantErr:   WebErrEmptyShortCode,
		},
		"fail - inactive url": {
			shortCode: "gg",
			mockCtrl: mockCtrl{
				inp: "gg",
				err: shorturl.ErrInactiveURL,
			},
			wantCode: http.StatusBadRequest,
			wantErr:  WebErrInactiveOriginalURL,
		},
		"fail - internal server error": {
			shortCode: "gg",
			mockCtrl: mockCtrl{
				inp: "gg",
				err: errors.New("some error"),
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  httpserver.ErrDefaultInternal,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/public/v1/redirect", nil)
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("shortcode", tc.shortCode)

			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
			rec := httptest.NewRecorder()
			req = req.WithContext(ctx)

			ctrl := new(shorturl.MockController)
			ctrl.ExpectedCalls = []*mock.Call{
				ctrl.On("Retrieve", mock.Anything, tc.shortCode).Return(tc.mockCtrl.output, tc.mockCtrl.err),
			}

			handler := Handler{shortUrlCtrl: ctrl}
			handler.Redirect().ServeHTTP(rec, req)
			require.Equal(t, tc.wantCode, rec.Code)

			var actErr httpserver.Error
			err := httpserver.ParseJSON(rec.Result().Body, &actErr)
			if tc.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.wantErr.Code, actErr.Code)
				require.Equal(t, tc.wantErr.Desc, actErr.Desc)
			} else {
				// require.Nil(t, err)
				require.Empty(t, actErr.Code)
				require.Empty(t, actErr.Desc)
			}
		})
	}
}
