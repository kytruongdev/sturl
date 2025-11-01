package httpserver

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/common"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring/logging"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func Handler(
	ctx context.Context,
	corsConf CORSConfig,
	routerFn func(r chi.Router),
) http.Handler {
	r := chi.NewRouter()

	r.Use(NewIdentifier(IdentifierConfig{
		EnableXCorrelationID: os.Getenv("ENABLE_X_CORRELATION_ID") == "1",
		EnableXRequestID:     os.Getenv("ENABLE_X_REQUEST_ID") == "1",
	}).Middleware)

	// Tracing middleware (auto start inbound spans)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			otelhttp.NewHandler(next, common.OpInbound).ServeHTTP(w, r)
		})
	})

	r.Use(logging.RequestLogger(ctx))

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
