package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/transportmeta"
)

// Handler builds the root chi.Router with middlewares and routes
func Handler(
	mon *monitoring.Monitor,
	corsConf CORSConfig,
	routerFn func(r chi.Router),
) http.Handler {
	r := chi.NewRouter()

	r.Use(transportmeta.Middleware(transportmeta.LoadConfigFromEnv()))
	r.Use(mon.Middleware())
	r.Use(cors.New(cors.Options{
		AllowedOrigins:   corsConf.allowedOrigins,
		AllowedMethods:   corsConf.allowedMethods,
		AllowedHeaders:   corsConf.allowedHeaders,
		ExposedHeaders:   corsConf.exposedHeaders,
		AllowCredentials: corsConf.allowCredentials,
		MaxAge:           corsConf.maxAge, // Maximum value not ignored by any of major browsers
	}).Handler)

	r.Get("/", checkLiveness)

	r.Group(routerFn)

	return r
}

func checkLiveness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write([]byte("ok!"))
}
