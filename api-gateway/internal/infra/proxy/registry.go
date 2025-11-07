package proxy

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/kytruongdev/sturl/api-gateway/internal/infra/common"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/monitoring/logging"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/transportmeta"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Config defines the configuration of an upstream service
type Config struct {
	Name             string // service identifier (e.g. "url-shortener-service")
	BaseURL          string // upstream base URL
	PathPrefix       string // optional prefix for upstream paths
	Retry            int    // reserved: number of retry attempts (future use)
	LogServiceName   bool   // whether to include "service" field in logs
	IncludeQueryLogs bool   // whether to log query string
}

var (
	registry = make(map[string]*httputil.ReverseProxy)
	regMu    sync.RWMutex
)

// Register adds a new ReverseProxy instance to the internal registry for later lookup
func Register(ctx context.Context, cfg Config) error {
	proxy, err := buildReverseProxy(cfg)
	if err != nil {
		return err
	}

	regMu.Lock()
	registry[cfg.Name] = proxy
	regMu.Unlock()

	monitoring.FromContext(ctx).LoggerWithSpan(ctx).Info().
		Str("service", cfg.Name).
		Str("target", cfg.BaseURL).
		Msg("proxy registered successfully")

	return nil
}

func buildReverseProxy(cfg Config) (*httputil.ReverseProxy, error) {
	if cfg.Name == "" || cfg.BaseURL == "" {
		return nil, errors.New("service name or baseURL missing")
	}

	target, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, err
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
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		IdleConnTimeout:       30 * time.Second,
		MaxIdleConnsPerHost:   50,
	}

	// Wrap with otelhttp transport to auto-create client span + inject headers
	wrapped := transportmeta.WrapTransport(baseTransport)
	proxy.Transport = otelhttp.NewTransport(wrapped)

	return proxy, nil
}

// get retrieves the reverse proxy for the given service name.
// Returns (nil, false) if not found.
func get(name string) (*httputil.ReverseProxy, bool) {
	regMu.RLock()
	p, ok := registry[name]
	regMu.RUnlock()
	return p, ok
}

// clearRegistry resets the global proxy registry (for testing only).
func clearRegistry() {
	regMu.Lock()
	defer regMu.Unlock()
	registry = make(map[string]*httputil.ReverseProxy)
}
