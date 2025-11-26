package main

import (
	"context"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/controller/outgoingevent"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
)

// Producer handles the publishing loop for Kafka outbox events.
// It periodically checks for pending events and publishes them.
type Producer struct {
	outgoingEventCtrl outgoingevent.Controller
	config            ProducerConfig
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
}

// New creates a new Producer instance.
func New(outgoingEventCtrl outgoingevent.Controller, config ProducerConfig) Producer {
	return Producer{
		outgoingEventCtrl: outgoingEventCtrl,
		config:            config,
	}
}

// Start begins the publishing loop.
// This loop follows pattern: run batch → sleep → repeat.
// No ticker is used to avoid overlapping batches.
func (p *Producer) Start(ctx context.Context) error {
	monitoring.Log(ctx).Info().
		Dur("polling_interval", p.config.pollingInterval).
		Int("batch_size", p.config.batchSize).
		Msg("[Producer.Start] Producer started")

	for {
		// Context cancelled → exit immediately
		newCtx := monitoring.NewContext()
		log := monitoring.Log(newCtx)
		select {
		case <-ctx.Done():
			log.Info().Msg("[Producer.Start] Producer stopping: context cancelled")
			return nil
		default:
		}

		start := time.Now()
		log.Info().Msg("[Producer.Start] Producer batch started")

		p.runOnce(newCtx)

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

// runOnce processes a single batch of events.
func (p *Producer) runOnce(ctx context.Context) {
	// Main controller logic: fetch pending → process → update statuses
	if err := p.outgoingEventCtrl.ProcessPending(ctx, outgoingevent.ProcessPendingInput{
		BatchSize: p.config.batchSize,
		MaxRetry:  p.config.maxRetry,
	}); err != nil {
		monitoring.Log(ctx).Error().Err(err).Msg("[Producer.runOnce] failed to process pending outgoing events")
	}
}
