package app

import (
	"errors"
	"os"
)

// Config represents application-level configuration for business logic.
// It contains service identification and environment settings.
type Config struct {
	ServiceName string
	AppEnv      string
}

// NewConfig creates a new app configuration from environment variables.
func NewConfig() Config {
	return Config{
		ServiceName: os.Getenv("SERVICE_NAME"),
		AppEnv:      os.Getenv("APP_ENV"),
	}
}

// Validate checks that the app configuration contains a valid fields.
func (cfg Config) Validate() error {
	if cfg.ServiceName == "" {
		return errors.New("[app.Config] required env variable 'SERVICE_NAME' not found")
	}

	if cfg.AppEnv == "" {
		return errors.New("[app.Config] required env variable 'APP_ENV' not found")
	}

	return nil
}
