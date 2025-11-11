package proxy

import (
	"testing"

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
