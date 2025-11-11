package proxy

import (
	"net/http"
)

// ProxyToService forwards the current HTTP request to a registered upstream service
func ProxyToService(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if p, ok := get(serviceName); ok {
			p.ServeHTTP(w, r)
			return
		}

		http.Error(w, "unknown service", http.StatusBadGateway)
	}
}
