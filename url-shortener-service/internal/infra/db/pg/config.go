package pg

import (
	"errors"
	"os"
)

// Config represents the PostgreSQL database configuration.
type Config struct {
	PGUrl string
}

// NewConfig creates a new PostgreSQL configuration from environment variables.
func NewConfig() Config {
	return Config{
		PGUrl: os.Getenv("PG_URL"),
	}
}

// Validate checks that the PostgreSQL configuration contains a valid connection URL.
func (c Config) Validate() error {
	if c.PGUrl == "" {
		return errors.New("[pg.Config] required env variable 'PG_URL' not found")
	}

	return nil
}
