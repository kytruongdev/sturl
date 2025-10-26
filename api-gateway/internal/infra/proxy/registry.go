package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/kytruongdev/sturl/api-gateway/internal/infra/logger"
)

// ProxyRegistry holds reusable reverse proxies by service name
type ProxyRegistry struct {
	proxies map[string]*httputil.ReverseProxy
}

// NewRegistry creates a new ProxyRegistry
func NewRegistry() *ProxyRegistry {
	return &ProxyRegistry{proxies: make(map[string]*httputil.ReverseProxy)}
}

// Register creates and stores a proxy for a given service
func (r *ProxyRegistry) Register(serviceName string, targetURL string) error {
	target, err := url.Parse(targetURL)
	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
	}

	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		zlog := logger.FromContext(req.Context())
		zlog.Error().
			Err(err).
			Str("service", serviceName).
			Str("target_host", target.Host).
			Str("path", req.URL.Path).
			Msg("proxy error")

		http.Error(rw, "upstream service unavailable", http.StatusBadGateway)
	}

	proxy.Transport = &http.Transport{
		ResponseHeaderTimeout: 5 * time.Second,
		IdleConnTimeout:       10 * time.Second,
	}

	r.proxies[serviceName] = proxy
	return nil
}

// Get returns a proxy by service name
func (r *ProxyRegistry) Get(serviceName string) (*httputil.ReverseProxy, bool) {
	p, ok := r.proxies[serviceName]
	return p, ok
}
