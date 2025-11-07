package app

import (
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/db/pg"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
)

// Config holds service-level settings for the API Gateway (service name, port, environment)
type Config struct {
	PGCfg     pg.Config
	ServerCfg httpserver.Config
}

// NewConfig loads Config from environment variables and sensible defaults
func NewConfig() Config {
	return Config{
		PGCfg:     pg.NewConfig(),
		ServerCfg: httpserver.NewConfig(),
	}
}

// Validate checks that the Config contains valid values (e.g., port within range)
func (c Config) Validate() error {
	if err := c.PGCfg.Validate(); err != nil {
		return err
	}

	return c.ServerCfg.Validate()
}
