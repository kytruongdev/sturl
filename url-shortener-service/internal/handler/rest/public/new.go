package public

import (
	"github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
)

// Handler is the web handler for this pkg
type Handler struct {
	shortUrlCtrl shorturl.Controller
}

// New instantiates a new Handler and returns it
func New(shortUrlCtrl shorturl.Controller) *Handler {
	return &Handler{shortUrlCtrl: shortUrlCtrl}
}
