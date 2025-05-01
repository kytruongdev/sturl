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

	r.URL.Host = target.Host
	r.URL.Scheme = target.Scheme
	r.Host = target.Host

	proxy.ServeHTTP(w, r)
}
