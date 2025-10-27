package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/env"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/proxy"
)

type Router struct {
	CorsOrigins []string
}

func (rtr Router) Routes(r chi.Router) {
	r.Group(rtr.public)
}

func (rtr Router) public(r chi.Router) {
	const prefix = "/api/public"
	r.Group(func(r chi.Router) {
		urlShortenerSvcName := env.GetAndValidateF("URL_SHORTENER_SERVICE_NAME")
		r.Post(prefix+"/v1/shorten", proxy.ProxyToService(urlShortenerSvcName))
		r.Get(prefix+"/v1/redirect/{shortcode}", proxy.ProxyToService(urlShortenerSvcName))
	})
}
