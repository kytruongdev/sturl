package monitoring

import (
	"context"

	"github.com/kytruongdev/sturl/api-gateway/internal/infra/monitoring/logging"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/monitoring/tracing"
)

// Config holds the monitoring configuration (e.g., log level, service name)
type Config struct {
	Logging logging.Config
	Tracing tracing.Config
}

// ConfigFromEnv loads monitoring configuration from environment variables
func ConfigFromEnv() Config {
	return Config{
		Logging: logging.FromEnv(),
		Tracing: tracing.FromEnv(),
	}
}

// Init sets up logging and tracing for the service using the provided configuration
func Init(ctx context.Context, cfg Config) (*Monitor, func(context.Context) error, error) {
	base := logging.New(cfg.Logging)
	ctx = logging.ToContext(ctx, base)

	shutdown, err := tracing.Init(ctx, cfg.Tracing)
	if err != nil {
		noop := func(context.Context) error { return nil }
		return &Monitor{base: base}, noop, err
	}
	return &Monitor{base: base}, shutdown, nil
}
