package public

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
)

// Redirect creates an HTTP handler function for redirecting short codes to their original URLs.
// It retrieves the original URL associated with the short code and returns an HTTP 301 redirect.
func (h *Handler) Redirect() http.HandlerFunc {
	return httpserver.HandlerErr(func(w http.ResponseWriter, r *http.Request) error {
		var err error
		ctx := r.Context()
		ctx, span := monitoring.Start(ctx, "Handler.Redirect")
		defer monitoring.End(span, &err)

		l := monitoring.Log(ctx)

		// Extract short code from URL path parameter
		shortCode := chi.URLParam(r, "shortcode")
		l.Info().Str("shortcode", shortCode).Msg("[Redirect] starting redirect from short code")

		if shortCode == "" {
			return WebErrEmptyShortCode
		}

		// Retrieve the original URL associated with the short code
		m, err := h.shortUrlCtrl.Retrieve(ctx, shortCode)
		if err != nil {
			l.Error().Stack().Err(err).Msg("[Redirect] h.shortUrlCtrl.Retrieve err")
			return convertControllerError(err)
		}

		l.Info().Str("original URL", m.OriginalURL).Str("shortcode", shortCode).Msg("[Redirect] redirecting to original URL")

		// Perform HTTP 301 (Moved Permanently) redirect to the original URL
		http.Redirect(w, r, m.OriginalURL, http.StatusMovedPermanently)

		return nil
	})
}
