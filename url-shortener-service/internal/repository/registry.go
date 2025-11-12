package repository

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/cenkalti/backoff/v4"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/db/pg"
	redisRepo "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/redis"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/shorturl"
)

// Registry defines an abstraction layer over all repositories,
// providing unified access and transaction management.
type Registry interface {
	ShortUrl() shorturl.Repository
	DoInTx(ctx context.Context, policy backoff.BackOff, fn func(ctx context.Context, txRepo Registry) error) error
}

// impl provides a concrete implementation of Registry.
type impl struct {
	db       *sql.DB
	tx       boil.ContextExecutor
	shortUrl shorturl.Repository
}

// New creates a new non-transactional repository registry.
func New(db *sql.DB, redisClient redisRepo.RedisClient) Registry {
	return impl{
		db:       db,
		shortUrl: shorturl.New(db, redisClient),
	}
}

// ShortUrl returns the shorturl repository.
func (i impl) ShortUrl() shorturl.Repository {
	return i.shortUrl
}

// DoInTx runs the provided function within a database transaction,
// automatically handling retries for transient errors (e.g., deadlocks,
// serialization failures) using an exponential backoff strategy.
//
// The 'policy' parameter defines the retry policy; if nil, a default one is used.
//
// Inside 'fn', a new transactional Registry instance is passed,
// where repository operations share the same *sql.Tx context.
func (i impl) DoInTx(ctx context.Context, policy backoff.BackOff, fn func(ctx context.Context, txRepo Registry) error) error {
	if policy == nil {
		policy = backoff.NewExponentialBackOff()
	}

	return pg.TxWithBackoff(ctx, i.db, policy, func(ctx context.Context, tx boil.ContextExecutor) error {
		return fn(ctx, impl{
			tx:       tx,
			shortUrl: i.shortUrl,
		})
	})
}
