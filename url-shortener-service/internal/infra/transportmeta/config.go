package transportmeta

import (
	"os"
	"strings"
)

// Config represents the configuration for request metadata propagation.
type Config struct {
	Required []string // List of required metadata header keys, e.g. ["X-Correlation-ID","X-Request-ID"]
}

// NewConfig creates a new transport metadata configuration from environment variables.
func NewConfig() Config {
	raw := os.Getenv("REQUIRES_METADATA")
	parts := strings.Split(raw, ",")
	var req []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			req = append(req, p)
		}
	}

	return Config{Required: req}
}

// Validate performs validation on the transport metadata configuration.
// Currently, no validation is required as all metadata headers are optional.
func (cfg Config) Validate() error {
	return nil
}
