package monitoring

import (
	"errors"
	"os"
)

// Config defines setup parameters for monitoring initialization.
type Config struct {
	ServiceName     string
	Env             string
	OTLPEndpointURL string
	LogLevel        string
	LogPretty       bool
}

// NewConfig creates a new monitoring configuration from environment variables.
func NewConfig() Config {
	return Config{
		LogPretty:       true,
		OTLPEndpointURL: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		LogLevel:        os.Getenv("LOG_LEVEL"),
		ServiceName:     os.Getenv("SERVICE_NAME"),
		Env:             os.Getenv("APP_ENV"),
	}
}

// Validate checks that the monitoring configuration contains all required fields.
func (cfg Config) Validate() error {
	if cfg.ServiceName == "" {
		return errors.New("[monitoring.Config] required env variable 'SERVICE_NAME' not found")
	}

	if cfg.Env == "" {
		return errors.New("[monitoring.Config] required env variable 'APP_ENV' not found")
	}

	if cfg.OTLPEndpointURL == "" {
		return errors.New("[monitoring.Config] required env variable 'OTEL_EXPORTER_OTLP_ENDPOINT' not found")
	}

	if cfg.LogLevel == "" {
		return errors.New("required env variable 'LOG_LEVEL' not found")
	}

	return nil
}
