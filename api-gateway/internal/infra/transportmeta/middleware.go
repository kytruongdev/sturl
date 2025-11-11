package transportmeta

import (
	"net/http"

	"github.com/rs/xid"
)

// Middleware creates an HTTP middleware that ensures inbound requests have standard metadata headers.
// It propagates metadata headers through the context and generates missing required headers automatically.
func Middleware(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			meta := ExtractFromRequest(r)

			// fill missing required keys
			for _, key := range cfg.Required {
				if meta.getByKey(key) == "" {
					// generate simple xid for both keys
					meta.setByKey(key, xid.New().String())
				}
			}

			ctx := ToContext(r.Context(), meta)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
