package proxy

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	t.Parallel()

	tcs := map[string]struct {
		cfg        Config
		expectErr  string
		serviceKey string
	}{
		"ok: valid config": {
			cfg: Config{
				UpstreamServiceName:    "shortener",
				UpstreamServiceBaseURL: "http://example.com",
			},
			serviceKey: "shortener",
		},
		"err: missing name": {
			cfg: Config{
				UpstreamServiceBaseURL: "http://example.com",
			},
			expectErr: "service name or baseURL missing",
		},
		"err: invalid url": {
			cfg: Config{
				UpstreamServiceName:    "bad",
				UpstreamServiceBaseURL: "://invalid",
			},
			expectErr: "invalid",
		},
	}

	for name, tt := range tcs {
		t.Run(name, func(t *testing.T) {
			err := Register(t.Context(), tt.cfg)
			if tt.expectErr == "" {
				require.NoError(t, err, "unexpected error for case %s", name)
				_, ok := get(tt.serviceKey)
				require.True(t, ok, "service %q should be registered", tt.serviceKey)
				return
			}

			require.Error(t, err, "expected error for case %s", name)
			require.Contains(t, err.Error(), tt.expectErr)
		})
	}
}

func TestRegister_ErrorHandlerTimeout(t *testing.T) {
	t.Run("err: timeout", func(t *testing.T) {
		slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond) // simulate slow upstream
		}))

		defer slow.Close()

		initTransportFunc = func() *http.Transport {
			return &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Millisecond,
					KeepAlive: 30 * time.Millisecond,
				}).DialContext,
				TLSHandshakeTimeout:   5 * time.Millisecond,
				ResponseHeaderTimeout: 5 * time.Millisecond,
				IdleConnTimeout:       30 * time.Millisecond,
				MaxIdleConnsPerHost:   50,
			}
		}

		defer func() {
			initTransportFunc = initTransport
		}()

		cfg := Config{
			UpstreamServiceName:    "shortener",
			UpstreamServiceBaseURL: "http://example.com",
		}

		err := Register(context.Background(), cfg)
		require.NoError(t, err, "failed to register slow-service")

		req := httptest.NewRequest(http.MethodGet, "/api/public/test", nil)
		rec := httptest.NewRecorder()

		h := ProxyToService("slow-service")
		h.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		require.Contains(t,
			[]int{http.StatusBadGateway, http.StatusGatewayTimeout},
			res.StatusCode,
			"expected 502 or 504 but got %d", res.StatusCode,
		)
	})
}
