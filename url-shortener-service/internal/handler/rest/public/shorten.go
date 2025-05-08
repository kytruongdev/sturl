package public

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/pkg/validator"
)

type ShortenRequest struct {
	OriginalURL string `json:"original_url"`
}

func (h *Handler) Shorten() http.HandlerFunc {
	return httpserver.HandlerErr(func(w http.ResponseWriter, r *http.Request) error {
		inp, err := mapToShortenInput(r)
		if err != nil {
			return err
		}

		rs, err := h.shortUrlCtrl.Shorten(r.Context(), inp)
		if err != nil {
			return convertControllerError(err)
		}

		httpserver.RespondJSON(w, rs)

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
