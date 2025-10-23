package logger

import (
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	bytesOut   int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	// default to 200 until WriteHeader is called
	return &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := lrw.ResponseWriter.Write(b)
	lrw.bytesOut += n
	return n, err
}

// RequestLogger injects a request-scoped logger with correlation id, timing, and http metadata.
// It also recovers from panics and logs them.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// try read inbound request id; if empty, generate
		corrID := r.Header.Get("X-Correlation-ID")
		if corrID == "" {
			corrID = xid.New().String()
		}

		// prepare per-request logger
		base := Get() // panic if not initialized (desired)
		reqLogger := base.With().
			Str("correlation_id", corrID).
			Str("request_id", xid.New().String()).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_ip", clientIP(r)).
			Str("user_agent", r.UserAgent()).
			Logger()

		// put logger + correlation id into context
		ctx := WithCorrelationID(r.Context(), corrID)
		ctx = WithLogger(ctx, &reqLogger)

		lrw := newLoggingResponseWriter(w)
		// reflect request id back to client
		lrw.Header().Set("X-Correlation-ID", corrID)
		lrw.Header().Set("X-Request-ID", corrID)

		defer func() {
			elapsed := time.Since(start)

			if rec := recover(); rec != nil {
				lrw.statusCode = http.StatusInternalServerError
				reqLogger.Error().
					Interface("panic", rec).
					Str("stacktrace", string(debug.Stack())).
					Dur("elapsed_ms", elapsed).
					Msg("panic recovered")
				panic(rec) // keep behavior consistent with current code
			}

			// decide level by status code
			lvl := levelByStatus(lrw.statusCode)
			e := reqLogger.WithLevel(lvl).
				Int("status_code", lrw.statusCode).
				Dur("elapsed_ms", elapsed)

			if lvl <= zerolog.DebugLevel {
				e = e.Int("bytes_out", lrw.bytesOut)
			}

			e.Msg("http_request")
		}()

		next.ServeHTTP(lrw, r.WithContext(ctx))
	})
}

// levelByStatus maps http status codes to log level.
func levelByStatus(code int) zerolog.Level {
	switch {
	case code >= 500:
		return zerolog.ErrorLevel
	case code >= 400:
		return zerolog.WarnLevel
	default:
		return zerolog.InfoLevel
	}
}

// clientIP tries best-effort to extract client ip (behind proxy friendly).
func clientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// may contain multiple, take first non-empty
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			p := strings.TrimSpace(parts[0])
			if p != "" {
				return p
			}
		}
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
