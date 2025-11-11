package monitoring

import (
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Middleware is the single entrypoint for all HTTP monitoring middlewares.
// It combines correlation and tracing in correct order.
func Middleware() func(http.Handler) http.Handler {
	// corr := CorrelationMiddleware()
	trace := middlewareHTTP()
	return func(next http.Handler) http.Handler {
		return trace(next)
	}
}

// middlewareHTTP adds tracing span and access log for each inbound request.
func middlewareHTTP() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, r)

			Log(r.Context()).Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Dur("latency", time.Since(start)).
				Msg("request completed")
		})
		return otelhttp.NewHandler(
			handler,
			"url-shortener.inbound",
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			}))
	}
}
