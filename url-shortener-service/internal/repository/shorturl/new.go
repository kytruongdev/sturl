package shorturl

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	redis2 "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/redis"
)

// Repository defines the interface for short URL data access operations.
// It provides the specification of the functionality provided by this package.
type Repository interface {
	GetByOriginalURL(context.Context, string) (model.ShortUrl, error)
	GetByShortCode(context.Context, string) (model.ShortUrl, error)
	Insert(context.Context, model.ShortUrl) (model.ShortUrl, error)
	Update(context.Context, model.ShortUrl, string) error
}

// impl is the implementation of the repository
type impl struct {
	db          boil.ContextExecutor
	redisClient redis2.RedisClient
}

// New creates and returns a new Repository instance with the provided database and Redis client.
// It returns a new instance of the repository for accessing short URL data.
func New(db boil.ContextExecutor, redisClient redis2.RedisClient) Repository {
	return &impl{db: db, redisClient: redisClient}
}
