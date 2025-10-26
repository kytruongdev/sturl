package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/kytruongdev/sturl/api-gateway/internal/infra/env"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/logger"
)

func ProxyToURLService(w http.ResponseWriter, r *http.Request) {
	target, _ := url.Parse(env.GetAndValidateF("URL_SHORTENER_URL"))
	proxy := httputil.NewSingleHostReverseProxy(target)

	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
	}

	proxy.ErrorHandler = func(rw http.ResponseWriter, r *http.Request, err error) {
		zlog := logger.FromContext(r.Context())
		zlog.Error().Err(err).Str("target_host", target.Host).Str("target_path", target.Path).Msg("proxy error")
		http.Error(rw, "upstream error", http.StatusBadGateway)
	}

	proxy.ServeHTTP(w, r)
}
