package pg

import (
	"errors"
	"os"
)

// Config represents config of PG database
type Config struct {
	PGUrl string
}

// NewConfig returns config
func NewConfig() Config {
	return Config{
		PGUrl: os.Getenv("PG_URL"),
	}
}

// Validate validates app config
func (c Config) Validate() error {
	if c.PGUrl == "" {
		return errors.New("required env variable 'PG_URL' not found")
	}

	return nil
}
