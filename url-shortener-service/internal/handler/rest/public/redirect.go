package public

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/logger"
)

// Redirect redirects to original url
func (h *Handler) Redirect() http.HandlerFunc {
	return httpserver.HandlerErr(func(w http.ResponseWriter, r *http.Request) error {
		l := logger.FromContext(r.Context())

		shortCode := chi.URLParam(r, "shortcode")
		l.Info().Str("shortcode", shortCode).Msg("[Redirect] starting redirect from short code")

		if shortCode == "" {
			return WebErrEmptyShortCode
		}

		m, err := h.shortUrlCtrl.Retrieve(r.Context(), shortCode)
		if err != nil {
			l.Error().Stack().Err(err).Msg("[Redirect] h.shortUrlCtrl.Retrieve err")
			return convertControllerError(err)
		}

		l.Info().Str("original URL", m.OriginalURL).Str("shortcode", shortCode).Msg("[Redirect] redirecting to original URL")

		http.Redirect(w, r, m.OriginalURL, http.StatusMovedPermanently)

		return nil
	})
}
