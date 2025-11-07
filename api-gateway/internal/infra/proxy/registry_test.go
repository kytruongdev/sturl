package proxy

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	t.Parallel()
	t.Cleanup(clearRegistry)

	tcs := map[string]struct {
		cfg        Config
		expectErr  string
		serviceKey string
	}{
		"ok: valid config": {
			cfg: Config{
				Name:             "shortener",
				BaseURL:          "http://example.com",
				LogServiceName:   true,
				IncludeQueryLogs: false,
			},
			serviceKey: "shortener",
		},
		"err: missing name": {
			cfg: Config{
				BaseURL: "http://example.com",
			},
			expectErr: "service name or baseURL missing",
		},
		"err: invalid url": {
			cfg: Config{
				Name:    "bad",
				BaseURL: "://invalid",
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
