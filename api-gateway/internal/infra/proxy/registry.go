package proxy

import (
	"context"
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/kytruongdev/sturl/api-gateway/internal/infra/common"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/monitoring/logging"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// ServiceConfig defines per-service proxy settings.
type ServiceConfig struct {
	Name             string        // service identifier (e.g. "url-shortener-service")
	BaseURL          string        // upstream base URL
	ResponseTimeout  time.Duration // timeout waiting for upstream response
	IdleConnTimeout  time.Duration // timeout for idle connections
	MaxIdleConns     int           // maximum idle connections per host
	PathPrefix       string        // optional prefix for upstream paths
	Retry            int           // reserved: number of retry attempts (future use)
	LogServiceName   bool          // whether to include "service" field in logs
	IncludeQueryLogs bool          // whether to log query string
}

var registry = map[string]*httputil.ReverseProxy{}

// Register creates and stores a reverse proxy for a given service config.
func Register(cfg ServiceConfig) error {
	if cfg.Name == "" || cfg.BaseURL == "" {
		return errors.New("service name or baseURL missing")
	}

	target, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Preserve default director but add tracing context propagation
	origDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		// Run original director (sets URL scheme/host/path)
		origDirector(req)

		// Always ensure correct upstream host
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host

		// Add prefix if needed
		if cfg.PathPrefix != "" && !strings.HasPrefix(req.URL.Path, cfg.PathPrefix) {
			req.URL.Path = cfg.PathPrefix + req.URL.Path
		}

		// Inject tracing context into downstream headers
		ctx := req.Context()
		// Propagate the trace context (traceparent, tracestate)
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	}

	// Structured error handling
	proxy.ErrorHandler = func(rw http.ResponseWriter, r *http.Request, err error) {
		l := logging.FromContext(r.Context())
		status := http.StatusBadGateway
		msg := "upstream error"

		if errors.Is(err, context.DeadlineExceeded) {
			status = http.StatusGatewayTimeout
			msg = "upstream timeout"
		}
		if errors.Is(err, context.Canceled) {
			l.Info().
				Str("service", cfg.Name).
				Str("target_host", target.Host).
				Str("path", r.URL.Path).
				Str("correlation_id", r.Header.Get(common.HeaderCorrelationID)).
				Str("request_id", r.Header.Get(common.HeaderRequestID)).
				Msg("client canceled")
			return
		}

		l.Error().
			Err(err).
			Str("service", cfg.Name).
			Str("target_host", target.Host).
			Str("path", r.URL.Path).
			Str("correlation_id", r.Header.Get(common.HeaderCorrelationID)).
			Str("request_id", r.Header.Get(common.HeaderRequestID)).
			Int("status", status).
			Msg("proxy error")

		http.Error(rw, msg, status)
	}

	// --- Set transport wrapped with OTel
	baseTransport := &http.Transport{
		ResponseHeaderTimeout: cfg.ResponseTimeout,
		IdleConnTimeout:       cfg.IdleConnTimeout,
		MaxIdleConnsPerHost:   cfg.MaxIdleConns,
	}

	// Wrap with otelhttp transport to auto-create client span + inject headers
	proxy.Transport = otelhttp.NewTransport(baseTransport)

	registry[cfg.Name] = proxy
	return nil
}

// get retrieves a proxy by service name.
func get(serviceName string) (*httputil.ReverseProxy, bool) {
	p, ok := registry[serviceName]
	return p, ok
}
