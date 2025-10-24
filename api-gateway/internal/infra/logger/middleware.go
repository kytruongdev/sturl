package logger

import (
	"context"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

// RequestLogger creates middleware using base logger from rootCtx
func RequestLogger(rootCtx context.Context) func(http.Handler) http.Handler {
	base := FromContext(rootCtx)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			corrID := r.Header.Get("X-Correlation-ID")
			if corrID == "" {
				corrID = xid.New().String()
			}

			reqLog := base.With().
				Str("correlation_id", corrID).
				Str("request_id", xid.New().String()).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_ip", clientIP(r)).
				Str("user_agent", r.UserAgent()).
				Logger()

			ctx := ToContext(r.Context(), Logger{zLog: &reqLog})
			lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			// reflect request id back to client
			lrw.Header().Set("X-Correlation-ID", corrID)
			lrw.Header().Set("X-Request-ID", corrID)

			defer func() {
				elapsed := time.Since(start)
				if rec := recover(); rec != nil {
					reqLog.Error().
						Interface("panic", rec).
						Str("stacktrace", string(debug.Stack())).
						Dur("elapsed_ms", elapsed).
						Msg("panic recovered")
					panic(rec)
				}

				lvl := levelByStatus(lrw.statusCode)
				e := reqLog.WithLevel(lvl).
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

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	bytesOut   int
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

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
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
