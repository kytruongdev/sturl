package kafka

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
)

// DefaultBackoffPolicy creates an exponential backoff policy.
func DefaultBackoffPolicy() backoff.BackOff {
	p := backoff.NewExponentialBackOff()
	p.InitialInterval = 500 * time.Millisecond
	p.MaxInterval = 5 * time.Second
	p.MaxElapsedTime = 0 // allow retries until max retries wrapper stops it
	return p
}

// retryKafka executes fn with retry logic using KafkaError.
// - nil → immediate success
// - non-retryable → stop immediately
// - retryable → retry according to policy
func retryKafka(
	ctx context.Context,
	policy backoff.BackOff,
	fn func(ctx context.Context) error,
) error {
	if policy == nil {
		policy = DefaultBackoffPolicy()
	}

	var attempt int
	var lastErr error

	op := func() error {
		attempt++

		log := monitoring.Log(ctx)

		select {
		case <-ctx.Done():
			log.Debug().Msgf("[retryKafka] ctx canceled at attempt %d", attempt)
			return backoff.Permanent(ctx.Err())
		default:
		}

		err := fn(ctx)

		if err == nil {
			log.Debug().Msgf("[retryKafka] success at attempt %d", attempt)
			lastErr = nil
			return nil
		}

		lastErr = err

		log.Debug().
			Int("attempt", attempt).
			Err(err).
			Msg("[retryKafka] retrying...")

		return err
	}

	backoff.Retry(op, policy)

	if lastErr != nil {
		monitoring.Log(ctx).Debug().
			Int("attempt", attempt).
			Err(lastErr).
			Msg("[retryKafka] final failure")
	} else {
		monitoring.Log(ctx).Debug().
			Int("attempt", attempt).
			Msg("[retryKafka] final success")
	}

	return lastErr
}
