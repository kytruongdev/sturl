package shorturl

import (
	"context"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	redis2 "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/redis"
	"github.com/volatiletech/sqlboiler/boil"
)

// Repository provides the specification of the functionality provided by this pkg
type Repository interface {
	GetByOriginalURL(context.Context, string) (model.ShortUrl, error)
	GetByShortCode(context.Context, string) (model.ShortUrl, error)
	Insert(context.Context, model.ShortUrl) (model.ShortUrl, error)
}

// impl is the implementation of the repository
type impl struct {
	db          boil.ContextExecutor
	redisClient redis2.RedisClient
}

// New returns a new instance of the repository
func New(db boil.ContextExecutor, redisClient redis2.RedisClient) Repository {
	return &impl{db: db, redisClient: redisClient}
}
