package main

import (
	"context"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/kafka"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository"
)

// Producer handles the publishing loop for Kafka outbox events.
// It periodically checks for pending events and publishes them.
type Producer struct {
	repo     repository.Registry
	producer kafka.Producer
	config   ProducerConfig
}

// ProducerConfig defines the behavior of the Outbox Producer.
// These values control how frequently the producer scans the outbox table
// and how many events it processes per cycle.
type ProducerConfig struct {
	// pollingInterval defines how frequently the Producer checks the outbox table
	// for new pending events to publish.
	pollingInterval time.Duration
	// batchSize controls how many events are processed per cycle.
	batchSize int
	// maxRetry specifies how many times the producer should retry publishing
	// an event before marking it as FAILED.
	maxRetry int
	// maxConcurrency controls how many events are published concurrently.
	maxConcurrency int
}

// New creates a new Producer instance.
func New(repo repository.Registry, producer kafka.Producer, config ProducerConfig) Producer {
	return Producer{
		repo:     repo,
		producer: producer,
		config:   config,
	}
}

// start begins the publishing loop.
// This loop follows pattern: run batch → sleep → repeat.
// No ticker is used to avoid overlapping batches.
func (p *Producer) start(ctx context.Context) error {
	monitoring.Log(ctx).Info().
		Dur("polling_interval", p.config.pollingInterval).
		Int("batch_size", p.config.batchSize).
		Msg("[Producer.Start] Producer started")

	for {
		select {
		case <-ctx.Done():
			monitoring.Log(ctx).Info().Msg("[Producer.Start] context canceled, stopping producer loop")
			return nil

		default:
			start := time.Now()
			log := monitoring.Log(ctx).
				Field("polling_interval", p.config.pollingInterval).
				Field("batch_size", p.config.batchSize)

			log.Info().Msg("[Producer.Start] Producer batch started")

			p.runOnce(ctx)

			log.Info().Msgf("[Producer.Start] Producer batch completed in %s", time.Since(start))

			// Sleep after finishing batch (Beaver style)
			select {
			case <-ctx.Done():
				log.Info().Msg("[Producer.Start] Producer stopped during sleep period")
				return nil
			case <-time.After(p.config.pollingInterval):
			}
		}
	}
}

// runOnce processes a single batch of events.
func (p *Producer) runOnce(ctx context.Context) {
	if err := p.process(ctx); err != nil {
		monitoring.Log(ctx).Error().Err(err).Msg("[Producer.runOnce] failed to process pending outgoing events")
	}
}
