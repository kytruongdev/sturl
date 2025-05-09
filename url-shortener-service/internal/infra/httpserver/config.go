package httpserver

import (
	"errors"
	"os"
)

type Config struct {
	ServerAddr string
	RedisAddr  string
}

// NewConfig returns config
func NewConfig() Config {
	return Config{
		ServerAddr: os.Getenv("SERVER_ADDR"),
		RedisAddr:  os.Getenv("REDIS_ADDR"),
	}
}

// Validate validates app config
func (c Config) Validate() error {
	if c.ServerAddr == "" {
		return errors.New("required env variable 'SERVER_ADDR' not found")
	}

	if c.RedisAddr == "" {
		return errors.New("required env variable 'REDIS_ADDR' not found")
	}

	return nil
}
