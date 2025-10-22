package public

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/pkg/validator"
)

// ShortenRequest represents shorten request payload
type ShortenRequest struct {
	OriginalURL string `json:"original_url"`
}

// ShortenResponse represents shorten response
type ShortenResponse struct {
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Shorten creates short code from original URL
func (h *Handler) Shorten() http.HandlerFunc {
	return httpserver.HandlerErr(func(w http.ResponseWriter, r *http.Request) error {
		inp, err := mapToShortenInput(r)
		if err != nil {
			return err
		}

		log.Printf("[Shorten] starting to make shorten url with inp: %+v\n", inp)

		rs, err := h.shortUrlCtrl.Shorten(r.Context(), inp)
		if err != nil {
			log.Printf("[Shorten] error: %+v\n", err)
			return convertControllerError(err)
		}

		log.Printf("[Shorten] end to shorten url with inp: %+v\n", inp)

		httpserver.RespondJSON(w, toShortenResponse(rs))

		return nil
	})
}

func mapToShortenInput(r *http.Request) (shorturl.ShortenInput, error) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return shorturl.ShortenInput{}, err
	}

	if req.OriginalURL == "" {
		return shorturl.ShortenInput{}, WebErrEmptyOriginalURL
	}

	if err := validator.ValidateURL(req.OriginalURL); err != nil {
		log.Println("Invalid original_url. Error detail: " + err.Error())
		return shorturl.ShortenInput{}, WebErrInvalidOriginalURL
	}

	return shorturl.ShortenInput{
		OriginalURL: req.OriginalURL,
	}, nil
}

func toShortenResponse(m model.ShortUrl) ShortenResponse {
	return ShortenResponse{
		ShortCode:   m.ShortCode,
		OriginalURL: m.OriginalURL,
		Status:      m.Status.String(),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
