package proxy

import (
	"net/http"
)

// ProxyToService forwards the request to the upstream service registered in proxy registry.
func ProxyToService(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if p, ok := Get(serviceName); ok {
			p.ServeHTTP(w, r)
			return
		}
		http.Error(w, "unknown service", http.StatusBadGateway)
	}
}
