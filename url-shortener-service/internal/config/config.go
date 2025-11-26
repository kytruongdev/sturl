package config

import (
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/app"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/db/pg"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/kafka"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/transportmeta"
)

// GlobalConfig represents the aggregated configuration for the URL shortener service.
// It combines all infrastructure component configurations (application, database, server, monitoring, transport)
// into a single structure for centralized management and validation.
// This package serves as the infrastructure configuration aggregator, making it easy to extend
// with new infrastructure components (e.g., Kafka, message queues) in the future.
type GlobalConfig struct {
	AppCfg           app.Config           // Application-level configuration (service name, environment)
	PGCfg            pg.Config            // PostgreSQL database connection configuration
	ServerCfg        httpserver.Config    // HTTP server configuration (address, Redis address)
	MonitoringCfg    monitoring.Config    // Observability configuration (logging, tracing, metrics)
	TransportMetaCfg transportmeta.Config // Request metadata propagation configuration
	KafkaCfg         kafka.Config
}

// NewGlobalConfig creates and loads a new GlobalConfig instance from environment variables.
// It initializes all infrastructure component configurations by calling their respective NewConfig functions.
// Each component config is loaded independently, allowing for modular configuration management.
func NewGlobalConfig() GlobalConfig {
	return GlobalConfig{
		AppCfg:           app.NewConfig(),
		PGCfg:            pg.NewConfig(),
		ServerCfg:        httpserver.NewConfig(),
		MonitoringCfg:    monitoring.NewConfig(),
		TransportMetaCfg: transportmeta.NewConfig(),
		KafkaCfg:         kafka.NewConfig(),
	}
}

// Validate performs validation on all infrastructure component configurations.
// It validates each component config in sequence and returns the first error encountered.
// This ensures all configurations are valid before the application starts.
func (c GlobalConfig) Validate() error {
	if err := c.AppCfg.Validate(); err != nil {
		return err
	}
	if err := c.PGCfg.Validate(); err != nil {
		return err
	}
	if err := c.ServerCfg.Validate(); err != nil {
		return err
	}
	if err := c.MonitoringCfg.Validate(); err != nil {
		return err
	}
	if err := c.TransportMetaCfg.Validate(); err != nil {
		return err
	}
	if err := c.KafkaCfg.Validate(); err != nil {
		return err
	}

	return nil
}
