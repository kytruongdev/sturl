package httpserver

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/logger"
)

func Handler(
	ctx context.Context,
	corsConf CORSConfig,
	routerFn func(r chi.Router),
) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.New(cors.Options{
		AllowedOrigins:   corsConf.allowedOrigins,
		AllowedMethods:   corsConf.allowedMethods,
		AllowedHeaders:   corsConf.allowedHeaders,
		ExposedHeaders:   corsConf.exposedHeaders,
		AllowCredentials: corsConf.allowCredentials,
		MaxAge:           corsConf.maxAge, // Maximum value not ignored by any of major browsers
	}).Handler)

	r.Use(NewIdentifier(IdentifierConfig{
		EnableXCorrelationID: os.Getenv("ENABLE_X_CORRELATION_ID") == "1",
		EnableXRequestID:     os.Getenv("ENABLE_X_REQUEST_ID") == "1",
	}).Middleware)

	r.Use(logger.RequestLogger(ctx))

	r.Get("/", checkLiveness)

	r.Group(routerFn)

	return r
}

func checkLiveness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write([]byte("ok!"))
}
