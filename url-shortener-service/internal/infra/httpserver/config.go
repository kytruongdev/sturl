package httpserver

import (
	"errors"
	"os"
)

// Config defines options for the HTTP server such as address and timeouts
type Config struct {
	ServerAddr  string
	RedisAddr   string
	LogLevel    string
	ServiceName string
	AppEnv      string
}

// NewConfig creates a new HTTP server configuration from environment variables
func NewConfig() Config {
	return Config{
		ServerAddr:  os.Getenv("SERVER_ADDR"),
		RedisAddr:   os.Getenv("REDIS_ADDR"),
		LogLevel:    os.Getenv("LOG_LEVEL"),
		ServiceName: os.Getenv("SERVICE_NAME"),
		AppEnv:      os.Getenv("APP_ENV"),
	}
}

// Validate ensures the HTTP server configuration is valid
func (c Config) Validate() error {
	if c.ServerAddr == "" {
		return errors.New("required env variable 'SERVER_ADDR' not found")
	}

	if c.RedisAddr == "" {
		return errors.New("required env variable 'REDIS_ADDR' not found")
	}

	if c.LogLevel == "" {
		return errors.New("required env variable 'LOG_LEVEL' not found")
	}

	if c.ServiceName == "" {
		return errors.New("required env variable 'SERVICE_NAME' not found")
	}

	if c.AppEnv == "" {
		return errors.New("required env variable 'APP_ENV' not found")
	}

	return nil
}
