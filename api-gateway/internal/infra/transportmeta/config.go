package transportmeta

import (
	"os"
	"strings"
)

// Config defines the behavior for propagating request metadata such as request IDs
type Config struct {
	Required []string // e.g. ["X-Correlation-ID","X-Request-ID"]
}

// LoadConfigFromEnv loads metadata configuration from the REQUIRES_METADATA
// environment variable, which is a comma-separated list of required headers.
func LoadConfigFromEnv() Config {
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
