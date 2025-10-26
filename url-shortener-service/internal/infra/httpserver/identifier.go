package httpserver

import (
	"net/http"

	"github.com/rs/xid"
)

// IdentifierConfig defines configuration for the Identifier middleware.
//
// This middleware ensures that every incoming HTTP request has unique
// identifiers such as X-Correlation-ID and/or X-Request-ID.
// These IDs are used to trace requests across distributed services.
type IdentifierConfig struct {
	EnableXCorrelationID bool
	EnableXRequestID     bool
	XIDs                 map[string]string
}

// Identifier represents a middleware that attaches unique
// identifiers to each HTTP request.
//
// It ensures every request entering the system is traceable
// across multiple services by having consistent header values.
type Identifier struct {
	config IdentifierConfig
}

// NewIdentifier creates a new Identifier middleware instance
// using the provided configuration.
func NewIdentifier(config IdentifierConfig) Identifier {
	return Identifier{
		config: config,
	}
}

// Middleware returns an http.Handler middleware that ensures
// each incoming request has X-Correlation-ID and/or X-Request-ID headers.
//
// If the headers are missing, it generates new ones using xid.
// The middleware also reflects these IDs back to the response headers
// so that clients and upstream services can trace requests easily.
func (req Identifier) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if req.config.EnableXCorrelationID {
			corrID := r.Header.Get("X-Correlation-ID")
			if corrID == "" {
				corrID = xid.New().String()
				r.Header.Set("X-Correlation-ID", corrID)
			}

			w.Header().Set("X-Correlation-ID", corrID)
		}

		if req.config.EnableXRequestID {
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = xid.New().String()
				r.Header.Set("X-Request-ID", reqID)
			}

			w.Header().Set("X-Request-ID", reqID)
		}

		for k, v := range req.config.XIDs {
			if k == "X-Correlation-ID" || k == "X-Request-ID" {
				continue
			}

			r.Header.Set(k, v)
			w.Header().Set(k, v)
		}

		next.ServeHTTP(w, r)
	})
}
