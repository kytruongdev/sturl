package public

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring/logging"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring/tracing"
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
		ctx := r.Context()

		var err error
		ctx, span := tracing.StartWithName(ctx, "Handler.Shorten")
		defer tracing.End(&span, &err)

		l := logging.FromContext(ctx)
		defer logging.TimeTrack(l, time.Now(), "handler.Shorten")

		inp, err := mapToShortenInput(r)
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

func mapToShortenInput(r *http.Request) (shorturl.ShortenInput, error) {
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
