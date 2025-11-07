package app

import (
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/httpserver"
)

// Config holds service-level settings for the API Gateway (service name, port, environment)
type Config struct {
	ServerCfg httpserver.Config
}

// NewConfig loads Config from environment variables and sensible defaults
func NewConfig() Config {
	return Config{
		ServerCfg: httpserver.NewConfig(),
	}
}

// Validate checks that the Config contains valid values (e.g., port within range)
func (c Config) Validate() error {
	return c.ServerCfg.Validate()
}
