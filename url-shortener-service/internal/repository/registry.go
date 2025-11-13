package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/cenkalti/backoff/v4"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/db/pg"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/kafkaoutboxevent"
	redisRepo "github.com/kytruongdev/sturl/url-shortener-service/internal/repository/redis"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/shorturl"
)

// Registry defines an abstraction layer over all repositories,
// providing unified access and transaction management.
type Registry interface {
	ShortUrl() shorturl.Repository
	KafkaOutboxEvent() kafkaoutboxevent.Repository
	DoInTx(ctx context.Context, backoffPolicy backoff.BackOff, fn func(ctx context.Context, txRepo Registry) error) error
}

// impl provides a concrete implementation of Registry.
type impl struct {
	db               *sql.DB
	tx               boil.ContextExecutor
	redisClient      redisRepo.RedisClient
	shortUrl         shorturl.Repository
	kafkaOutboxEvent kafkaoutboxevent.Repository
}

// New creates a new non-transactional repository registry.
func New(db *sql.DB, redisClient redisRepo.RedisClient) Registry {
	return impl{
		db:               db,
		redisClient:      redisClient,
		shortUrl:         shorturl.New(db, redisClient),
		kafkaOutboxEvent: kafkaoutboxevent.New(db),
	}
}

// ShortUrl returns the shorturl repository.
func (i impl) ShortUrl() shorturl.Repository {
	return i.shortUrl
}

// KafkaOutboxEvent returns the kafkaoutboxevent repository.
func (i impl) KafkaOutboxEvent() kafkaoutboxevent.Repository {
	return i.kafkaOutboxEvent
}

// DoInTx runs the provided function within a database transaction,
// automatically handling retries for transient errors (e.g., deadlocks,
// serialization failures) using an exponential backoff strategy.
//
// The 'policy' parameter defines the retry policy; if nil, a default one is used.
//
// Inside 'fn', a new transactional Registry instance is passed,
// where repository operations share the same *sql.Tx context.
func (i impl) DoInTx(ctx context.Context, backoffPolicy backoff.BackOff, fn func(ctx context.Context, txRepo Registry) error) error {
	if backoffPolicy == nil {
		backoffPolicy = pg.ExponentialBackOff(3, time.Minute)
	}

	return pg.TxWithBackoff(ctx, i.db, backoffPolicy, func(ctx context.Context, tx boil.ContextExecutor) error {
		return fn(ctx, impl{
			tx:               tx,
			shortUrl:         shorturl.New(tx, i.redisClient),
			kafkaOutboxEvent: kafkaoutboxevent.New(tx),
		})
	})
}
