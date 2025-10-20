package public

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
)

// Redirect redirects to original url
func (h *Handler) Redirect() http.HandlerFunc {
	return httpserver.HandlerErr(func(w http.ResponseWriter, r *http.Request) error {
		shortCode := chi.URLParam(r, "shortcode")
		log.Printf("[Redirect] starting redirect to short code %s", shortCode)
		if shortCode == "" {
			return &httpserver.Error{
				Status: http.StatusBadRequest,
				Code:   "empty_short_code",
				Desc:   "Empty short_code",
			}
		}

		m, err := h.shortUrlCtrl.Retrieve(r.Context(), shortCode)
		if err != nil {
			log.Println("[Redirect] error:", err)
			return convertControllerError(err)
		}

		log.Printf("[Redirect] end to redirect to original URL %v from shortcode %v\n", m.OriginalURL, shortCode)

		http.Redirect(w, r, m.OriginalURL, http.StatusMovedPermanently)

		return nil
	})
}
