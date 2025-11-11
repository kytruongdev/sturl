package public

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/pkg/validator"
)

// ShortenRequest represents the HTTP request payload for creating a short URL.
type ShortenRequest struct {
	OriginalURL string `json:"original_url"`
}

// ShortenResponse represents the HTTP response for a successful URL shortening operation.
type ShortenResponse struct {
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Shorten creates an HTTP handler function for shortening URLs.
// It validates the request, creates a short code from the original URL, and returns the result.
func (h *Handler) Shorten() http.HandlerFunc {
	return httpserver.HandlerErr(func(w http.ResponseWriter, r *http.Request) error {
		var err error
		ctx := r.Context()
		ctx, span := monitoring.Start(ctx, "Handler.Shorten")
		defer monitoring.End(span, &err)

		l := monitoring.Log(ctx)

		inp, err := validateAndMapToShortenInput(ctx, r)
		if err != nil {
			l.Error().Stack().Err(err).Msg("[Shorten] mapToShortenInput err")
			return err
		}

		l.Info().Interface("inp", inp).Msg("[Shorten] starting to make shorten url")

		rs, err := h.shortUrlCtrl.Shorten(ctx, inp)
		if err != nil {
			l.Error().Stack().Err(err).Msg("[Shorten] h.shortUrlCtrl.Shorten err")
			return convertControllerError(err)
		}

		l.Info().Msg("[Shorten] end to shorten url")

		httpserver.RespondJSON(w, toShortenResponse(rs))

		return nil
	})
}

func validateAndMapToShortenInput(ctx context.Context, r *http.Request) (shorturl.ShortenInput, error) {
	l := monitoring.Log(ctx)
	defer l.TimeTrack(time.Now(), "[Shorten] validate and map to ShortenInput")

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return shorturl.ShortenInput{}, err
	}

	if req.OriginalURL == "" {
		return shorturl.ShortenInput{}, WebErrEmptyOriginalURL
	}

	if err := validator.ValidateURL(req.OriginalURL); err != nil {
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
