package public

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring/logging"
	"go.opentelemetry.io/otel"
)

// Redirect redirects to original url
func (h *Handler) Redirect() http.HandlerFunc {
	return httpserver.HandlerErr(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		tracer := otel.Tracer("url-shortener.handler")
		ctx, span := tracer.Start(ctx, "Handler.Redirect")
		defer span.End()

		l := logging.FromContext(ctx)
		defer logging.TimeTrack(l, time.Now(), "handler.Redirect")

		shortCode := chi.URLParam(r, "shortcode")
		l.Info().Str("shortcode", shortCode).Msg("[Redirect] starting redirect from short code")

		if shortCode == "" {
			span.RecordError(WebErrEmptyShortCode)
			return WebErrEmptyShortCode
		}

		m, err := h.shortUrlCtrl.Retrieve(ctx, shortCode)
		if err != nil {
			span.RecordError(err)
			l.Error().Stack().Err(err).Msg("[Redirect] h.shortUrlCtrl.Retrieve err")
			return convertControllerError(err)
		}

		l.Info().Str("original URL", m.OriginalURL).Str("shortcode", shortCode).Msg("[Redirect] redirecting to original URL")

		http.Redirect(w, r, m.OriginalURL, http.StatusMovedPermanently)

		return nil
	})
}
