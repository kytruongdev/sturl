package transportmeta

import (
	"net/http"

	"github.com/rs/xid"
)

// Middleware ensures inbound requests have standard metadata headers and propagates them through the context
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
