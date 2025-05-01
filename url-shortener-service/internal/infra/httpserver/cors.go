package httpserver

import (
	"net/http"
)

// CORSConfig holds the CORS configuration
type CORSConfig struct {
	allowedOrigins   []string
	allowedMethods   []string
	allowedHeaders   []string
	exposedHeaders   []string
	allowCredentials bool
	maxAge           int
}

// NewCORSConfig initializes and returns a CORSConfig
func NewCORSConfig(origins []string, opts ...CORSOption) CORSConfig {
	cfg := CORSConfig{
		allowedOrigins: origins,
		allowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		allowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
		},
		exposedHeaders:   []string{"Link"},
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
