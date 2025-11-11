package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/handler/rest/public"
)

// Router represents the HTTP router configuration for the URL shortener service.
type Router struct {
	CorsOrigins  []string
	ShortURLCtrl shorturl.Controller
}

// Routes registers all routes on the provided chi.Router.
func (rtr Router) Routes(r chi.Router) {
	r.Group(rtr.public)
}

func (rtr Router) public(r chi.Router) {
	const prefix = "/api/public"
	r.Group(func(r chi.Router) {
		shortURLHandler := public.New(rtr.ShortURLCtrl)
		r.Post(prefix+"/v1/shorten", shortURLHandler.Shorten())
		r.Get(prefix+"/v1/redirect/{shortcode}", shortURLHandler.Redirect())
	})
}
