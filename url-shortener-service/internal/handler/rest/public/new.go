package public

import (
	"github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
)

// Handler represents the HTTP handler for public short URL endpoints.
type Handler struct {
	shortUrlCtrl shorturl.Controller
}

// New creates and returns a new Handler instance with the provided controller.
func New(shortUrlCtrl shorturl.Controller) *Handler {
	return &Handler{shortUrlCtrl: shortUrlCtrl}
}
