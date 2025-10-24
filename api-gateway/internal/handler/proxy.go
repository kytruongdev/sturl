package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/kytruongdev/sturl/api-gateway/internal/infra/env"
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

	proxy.ServeHTTP(w, r)
}
