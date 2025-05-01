package handler

import (
	"context"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	Ctx         context.Context
	CorsOrigins []string
}

func (rtr Router) Routes(r chi.Router) {
	r.Group(rtr.public)
}

func (rtr Router) public(r chi.Router) {
	const prefix = "/api/public"
	r.Group(func(r chi.Router) {
		r.Get(prefix+"/v1/users", ProxyToURLService)
		r.Post(prefix+"/v1/users", ProxyToURLService)
	})
}
