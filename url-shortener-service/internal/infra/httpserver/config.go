package httpserver

import (
	"errors"
	"os"
)

// Config defines options for the HTTP server such as address and timeouts
type Config struct {
	ServerAddr string
	RedisAddr  string
}

// NewConfig creates a new HTTP server configuration from environment variables
func NewConfig() Config {
	return Config{
		ServerAddr: os.Getenv("SERVER_ADDR"),
		RedisAddr:  os.Getenv("REDIS_ADDR"),
	}
}

// Validate ensures the HTTP server configuration is valid
func (c Config) Validate() error {
	if c.ServerAddr == "" {
		return errors.New("[httpserver.Config] required env variable 'SERVER_ADDR' not found")
	}

	if c.RedisAddr == "" {
		return errors.New("[httpserver.Config] required env variable 'REDIS_ADDR' not found")
	}

	return nil
}
