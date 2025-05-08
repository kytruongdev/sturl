package shorturl

import (
	"context"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/shorturl"
)

// Controller provides the specification of the functionality provided by this pkg
type Controller interface {
	Shorten(context.Context, ShortenInput) (model.ShortUrl, error)
	Retrieve(context.Context, string) (model.ShortUrl, error)
}

// impl is the implementation of the controller
type impl struct {
	shortUrlRepo shorturl.Repository
}

// New returns a new instance of the controller
func New(shortUrlRepo shorturl.Repository) Controller {
	return &impl{shortUrlRepo: shortUrlRepo}
}
