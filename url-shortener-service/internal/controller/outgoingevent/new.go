package outgoingevent

import (
	"context"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/kafka"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository"
)

// Controller defines the interface for short URL business logic operations.
// It provides the specification of the functionality provided by this package.
type Controller interface {
	ProcessPending(context.Context, ProcessPendingInput) error
}

// impl is the implementation of the controller
type impl struct {
	repo     repository.Registry
	producer kafka.Producer
}

// New creates and returns a new Controller instance with the provided repository.
// It returns a new instance of the controller for handling short URL operations.
func New(repo repository.Registry, producer kafka.Producer) Controller {
	return &impl{
		repo:     repo,
		producer: producer,
	}
}
