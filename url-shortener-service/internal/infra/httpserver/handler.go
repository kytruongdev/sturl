package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/transportmeta"
)

// HandlerWithHealth builds the root chi.Router with middlewares, routes, and health checks.
// Use this when you need to pass database and Redis clients for readiness checks.
func HandlerWithHealth(
	corsConf CORSConfig,
	transportMetaConf transportmeta.Config,
	routerFn func(r chi.Router),
	readinessCfg ReadinessConfig,
) http.Handler {
	r := chi.NewRouter()

	r.Use(transportmeta.Middleware(transportMetaConf))
	r.Use(monitoring.Middleware())
	r.Use(cors.New(cors.Options{
		AllowedOrigins:   corsConf.allowedOrigins,
		AllowedMethods:   corsConf.allowedMethods,
		AllowedHeaders:   corsConf.allowedHeaders,
		ExposedHeaders:   corsConf.exposedHeaders,
		AllowCredentials: corsConf.allowCredentials,
		MaxAge:           corsConf.maxAge, // Maximum value not ignored by any of major browsers
	}).Handler)

	// Health check endpoints
	r.Get("/", checkLiveness)
	r.Get("/health/ready", checkReadiness(readinessCfg))

	r.Group(routerFn)

	return r
}

// checkLiveness handles the root path health check endpoint.
// It returns a simple "ok!" response to indicate the server is running.
func checkLiveness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write([]byte("ok!"))
}
