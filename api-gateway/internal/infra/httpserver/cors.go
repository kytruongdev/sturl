package httpserver

import (
	"net/http"

	"github.com/kytruongdev/sturl/api-gateway/internal/infra/common"
)

// CORSConfig holds the configuration options for Cross-Origin Resource Sharing (CORS)
type CORSConfig struct {
	allowedOrigins   []string
	allowedMethods   []string
	allowedHeaders   []string
	exposedHeaders   []string
	allowCredentials bool
	maxAge           int
}

// NewCORSConfig returns a CORS middleware instance using the provided configuration
func NewCORSConfig(origins []string, opts ...CORSOption) CORSConfig {
	cfg := CORSConfig{
		allowedOrigins: origins,
		allowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		allowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
		},
		exposedHeaders:   []string{"Link", common.HeaderCorrelationID, common.HeaderRequestID},
		allowCredentials: true,
		maxAge:           300,
	}

	for _, o := range opts {
		o(&cfg)
	}

	return cfg
}

// CORSOption enables tweaking the CORSConfig
type CORSOption func(*CORSConfig)
