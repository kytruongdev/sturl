package shorturl

import (
	"context"
	"database/sql"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/redis/go-redis/v9"
)

// Repository provides the specification of the functionality provided by this pkg
type Repository interface {
	Insert(context.Context, model.ShortUrl) (model.ShortUrl, error)
	GetByOriginalURL(context.Context, string) (model.ShortUrl, error)
	GetByShortCode(context.Context, string) (model.ShortUrl, error)
	GetByShortCodeFromCache(context.Context, string) (model.ShortUrl, error)
	SetToCache(context.Context, model.ShortUrl) error
}

// impl is the implementation of the repository
type impl struct {
	db          *sql.DB
	redisClient *redis.Client
}

// New returns a new instance of the repository
func New(db *sql.DB, redisClient *redis.Client) Repository {
	return &impl{db: db, redisClient: redisClient}
}
