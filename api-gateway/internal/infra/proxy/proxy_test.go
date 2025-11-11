package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProxyToService(t *testing.T) {
	t.Parallel()

	okUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok!")
	}))
	defer okUpstream.Close()

	setupProxyForTesting(t, "ok-service", okUpstream)

	tcs := map[string]struct {
		serviceName string
		method      string
		path        string
		expectCode  int
		expectBody  string
	}{
		"ok: known service proxied": {
			serviceName: "ok-service",
			method:      http.MethodGet,
			path:        "/api/public/v1/redirect/123",
			expectCode:  http.StatusOK,
			expectBody:  "ok!",
		},
		"err: unknown service returns 502": {
			serviceName: "unknown-service",
			method:      http.MethodGet,
			path:        "/whatever",
			expectCode:  http.StatusBadGateway,
			expectBody:  "unknown service",
		},
	}

	for name, tt := range tcs {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			h := ProxyToService(tt.serviceName)
			h.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)

			require.Equal(t, tt.expectCode, res.StatusCode,
				"unexpected status code for case %s", name)

			if tt.expectBody != "" {
				require.Contains(t, string(body), tt.expectBody,
					"unexpected body for case %s", name)
			}
		})
	}
}

// setupProxyForTesting registers fake service in proxy registry
func setupProxyForTesting(t *testing.T, name string, upstream *httptest.Server) {
	t.Helper()
	cfg := Config{
		UpstreamServiceName:    name,
		UpstreamServiceBaseURL: upstream.URL,
	}
	err := Register(t.Context(), cfg)
	require.NoError(t, err, "failed to register proxy for %s", name)
}
