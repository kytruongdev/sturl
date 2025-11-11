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

	"github.com/kytruongdev/sturl/api-gateway/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/transportmeta"
	pkgerrors "github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Config represents the configuration for registering an upstream service proxy.
// It defines the configuration of an upstream service including its name, base URL, and optional path prefix.
type Config struct {
	UpstreamServiceName    string // Name identifier for the upstream service
	UpstreamServiceBaseURL string // Base URL where the upstream service is hosted
	PathPrefix             string // Optional path prefix to prepend to requests
}

var (
	// registry stores registered reverse proxy instances keyed by service name.
	// It uses sync.Map for concurrent-safe access.
	registry sync.Map
)

// Register adds a new ReverseProxy instance to the internal registry for later lookup
func Register(ctx context.Context, cfg Config) error {
	proxy, err := buildReverseProxy(cfg)
	if err != nil {
		return pkgerrors.Wrap(err, "error building proxy for "+cfg.UpstreamServiceName)
	}

	registry.Store(cfg.UpstreamServiceName, proxy)

	// svcName, _ := config.Get(ctx, config.CtxKeyServiceName)

	l := monitoring.Log(ctx)

	l.Info().
		Str("service", "api-gateway").
		Str("target", cfg.UpstreamServiceBaseURL).
		Msg("proxy registered successfully")

	return nil
}

func buildReverseProxy(cfg Config) (*httputil.ReverseProxy, error) {
	if cfg.UpstreamServiceName == "" || cfg.UpstreamServiceBaseURL == "" {
		return nil, errors.New("service name or baseURL missing")
	}

	target, err := url.Parse(cfg.UpstreamServiceBaseURL)
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
		l := monitoring.Log(r.Context())
		status := http.StatusBadGateway
		msg := "upstream error"

		if errors.Is(err, context.DeadlineExceeded) {
			status = http.StatusGatewayTimeout
			msg = "upstream timeout"
		}

		if errors.Is(err, context.Canceled) {
			l.Info().
				Str("service", cfg.UpstreamServiceName).
				Str("target_host", target.Host).
				Str("path", r.URL.Path).
				Str("correlation_id", r.Header.Get("X-Correlation-ID")).
				Str("request_id", r.Header.Get("X-Request-ID")).
				Msg("client canceled")
			return
		}

		l.Error().
			Err(err).
			Str("service", cfg.UpstreamServiceName).
			Str("target_host", target.Host).
			Str("path", r.URL.Path).
			Str("correlation_id", r.Header.Get("X-Correlation-ID")).
			Str("request_id", r.Header.Get("X-Request-ID")).
			Int("status", status).
			Msg("proxy error")

		http.Error(rw, msg, status)
	}

	// --- Set transport wrapped with OTel
	wrapped := transportmeta.WrapTransport(initTransportFunc())
	proxy.Transport = otelhttp.NewTransport(wrapped)

	return proxy, nil
}

// get retrieves the reverse proxy for the given service name.
// Returns (nil, false) if not found.
func get(name string) (*httputil.ReverseProxy, bool) {
	v, ok := registry.Load(name)
	if !ok {
		return nil, false
	}
	return v.(*httputil.ReverseProxy), true
}

// clearRegistry resets the global proxy registry (for testing only).
func clearRegistry() {
	registry.Range(func(k, _ any) bool {
		registry.Delete(k)
		return true
	})
}

const (
	dialerKeepAlive       = 30 * time.Second
	dialerTimeout         = 5 * time.Second
	tlsHandshakeTimeout   = 5 * time.Second
	responseHeaderTimeout = 5 * time.Second
	idleConnTimeout       = 30 * time.Second
	maxIdleConnsPerHost   = 50
)

var (
	initTransportFunc = initTransport
)

func initTransport() *http.Transport {
	return &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   dialerTimeout,
			KeepAlive: dialerKeepAlive,
		}).DialContext,
		TLSHandshakeTimeout:   tlsHandshakeTimeout,
		ResponseHeaderTimeout: responseHeaderTimeout,
		IdleConnTimeout:       idleConnTimeout,
		MaxIdleConnsPerHost:   maxIdleConnsPerHost,
	}
}
