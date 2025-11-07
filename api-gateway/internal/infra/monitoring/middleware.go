package monitoring

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/kytruongdev/sturl/api-gateway/internal/infra/monitoring/logging"
	"github.com/kytruongdev/sturl/api-gateway/internal/infra/transportmeta"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

// Middleware attaches structured logging, tracing, and panic recovery
// to each incoming HTTP request.
//
// The middleware performs the following steps:
//  1. Starts an OpenTelemetry span for the request.
//  2. Injects the span and a contextual logger into the request context.
//  3. Handles any panic raised downstream, logs it, and returns a 500 error.
//  4. Emits request metrics such as latency, method, and status code.
func (m *Monitor) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ctx := r.Context()

			// extract trace info
			sc := trace.SpanContextFromContext(ctx)
			traceID, spanID := "", ""
			if sc.HasTraceID() {
				traceID = sc.TraceID().String()
			}
			if sc.HasSpanID() {
				spanID = sc.SpanID().String()
			}

			// extract meta (correlation_id, request_id, ...)
			meta := transportmeta.FromContext(ctx)

			reqLogger := m.base.With(
				logging.String("correlation_id", meta.CorrelationID),
				logging.String("request_id", meta.RequestID),
				logging.String("trace_id", traceID),
				logging.String("span_id", spanID),
				logging.String("method", r.Method),
				logging.String("path", r.URL.Path),
				logging.String("remote_ip", clientIP(r)),
			)

			// inject logger + monitor back to context
			ctx = logging.ToContext(ctx, reqLogger)
			ctx = ToContext(ctx, &Monitor{base: reqLogger})
			r = r.WithContext(ctx)

			ww := wrapResponseWriter(w)

			defer func() {
				if rec := recover(); rec != nil {
					var err error

					switch recTyped := rec.(type) {
					case error:
						err = recTyped
					default:
						err = fmt.Errorf("%v", rec)
					}

					switch {
					// typical reverseproxy abort panic — safe to ignore
					case errors.Is(err, http.ErrAbortHandler):
						reqLogger.Warn().Msg("client aborted connection or proxy aborted response — ignored")
						return
					// often debug.Stack() prints {} for nil/empty panic
					case rec == nil || fmt.Sprintf("%v", rec) == "{}":
						reqLogger.Debug().Msg("empty panic recovered (likely reverseproxy cleanup)")
						return
					default:
						reqLogger.Error().
							Str("panic_type", fmt.Sprintf("%T", rec)).
							Str("panic_value", fmt.Sprintf("%v", rec)).
							Bytes("stack", debug.Stack()).
							Msg("panic recovered in Monitor.Middleware")
						// return clean 500 (no double WriteHeader)
						http.Error(ww, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					}
				}

				// always log completion
				lat := time.Since(start)
				ev := reqLogger.Info().
					Int("status", ww.status).
					Dur("latency", lat).
					Int64("bytes", ww.bytes)

				if ww.status >= 500 {
					ev = reqLogger.Error().
						Int("status", ww.status).
						Dur("latency", lat).
						Int64("bytes", ww.bytes)
				}
				ev.Msg("request completed")
			}()

			next.ServeHTTP(ww, r)
		})

		return otelhttp.NewHandler(inner, "",
			otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
				return r.Method + " " + r.URL.Path
			}),
		)
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

// WriteHeader captures the HTTP response status code before passing it to the underlying ResponseWriter
func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Write delegates the response body write operation to the underlying
// ResponseWriter and ensures the byte count is tracked for metrics
func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += int64(n)
	return n, err
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return xr
	}
	h, _, _ := net.SplitHostPort(r.RemoteAddr)
	return h
}
