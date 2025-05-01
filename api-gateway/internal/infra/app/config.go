package app

import (
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/httpserver"
)

type Config struct {
	ServerCfg httpserver.Config
}

// NewConfig returns config
func NewConfig() Config {
	return Config{
		ServerCfg: httpserver.NewConfig(),
	}
}

// Validate validates app config
func (c Config) Validate() error {
	return c.ServerCfg.Validate()
}
