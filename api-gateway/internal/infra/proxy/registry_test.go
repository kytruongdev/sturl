package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	t.Parallel()

	tcs := map[string]struct {
		cfg        ServiceConfig
		expectErr  string
		serviceKey string
	}{
		"ok: valid config": {
			cfg: ServiceConfig{
				Name:             "shortener",
				BaseURL:          "http://example.com",
				ResponseTimeout:  10 * time.Second,
				IdleConnTimeout:  10 * time.Second,
				MaxIdleConns:     10,
				LogServiceName:   true,
				IncludeQueryLogs: false,
			},
			serviceKey: "shortener",
		},
		"err: missing name": {
			cfg: ServiceConfig{
				BaseURL: "http://example.com",
			},
			expectErr: "service name or baseURL missing",
		},
		"err: invalid url": {
			cfg: ServiceConfig{
				Name:    "bad",
				BaseURL: "://invalid",
			},
			expectErr: "invalid",
		},
	}

	for name, tt := range tcs {
		t.Run(name, func(t *testing.T) {
			err := Register(tt.cfg)
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
	t.Parallel()

	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // simulate slow upstream
	}))
	defer slow.Close()

	cfg := ServiceConfig{
		Name:            "slow-service",
		BaseURL:         slow.URL,
		ResponseTimeout: 50 * time.Millisecond, // intentionally shorter to trigger timeout
		IdleConnTimeout: 100 * time.Millisecond,
		MaxIdleConns:    5,
		LogServiceName:  true,
	}

	err := Register(cfg)
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
}
