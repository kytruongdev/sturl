package handler

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/handler/rest/public"
)

type Router struct {
	Ctx          context.Context
	CorsOrigins  []string
	ShortURLCtrl shorturl.Controller
}

func (rtr Router) Routes(r chi.Router) {
	r.Group(rtr.public)
}

func (rtr Router) public(r chi.Router) {
	const prefix = "/api/public"
	r.Group(func(r chi.Router) {
		shortURLHandler := public.New(rtr.ShortURLCtrl)
		r.Post(prefix+"/v1/shorten", shortURLHandler.Shorten())
	})
}
