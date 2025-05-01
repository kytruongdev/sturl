package app

import (
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/db/pg"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
)

type Config struct {
	PGCfg     pg.Config
	ServerCfg httpserver.Config
}

// NewConfig returns config
func NewConfig() Config {
	return Config{
		PGCfg:     pg.NewConfig(),
		ServerCfg: httpserver.NewConfig(),
	}
}

// Validate validates app config
func (c Config) Validate() error {
	if err := c.PGCfg.Validate(); err != nil {
		return err
	}

	return c.ServerCfg.Validate()
}
