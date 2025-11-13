package kafkaoutboxevent

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
)

// Repository defines the interface for short URL data access operations.
// It provides the specification of the functionality provided by this package.
type Repository interface {
	GetByStatus(ctx context.Context, status string) ([]model.KafkaOutboxEvent, error)
	Insert(context.Context, model.KafkaOutboxEvent) (model.KafkaOutboxEvent, error)
}

// impl is the implementation of the repository
type impl struct {
	db boil.ContextExecutor
}

// New creates and returns a new Repository instance with the provided database and Redis client.
// It returns a new instance of the repository for accessing short URL data.
func New(db boil.ContextExecutor) Repository {
	return &impl{db: db}
}
