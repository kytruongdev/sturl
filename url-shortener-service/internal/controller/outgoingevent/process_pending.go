package outgoingevent

import (
	"context"
	"encoding/json"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
)

// ProcessPendingInput contains the configuration for processing pending events.
type ProcessPendingInput struct {
	BatchSize int
	MaxRetry  int
}

// ProcessPending loads a batch of pending events and attempts to publish
// each event to Kafka, updating their status accordingly.
//
// It is designed to be called repeatedly by a background worker.
func (i impl) ProcessPending(ctx context.Context, inp ProcessPendingInput) error {
	l := monitoring.Log(ctx)

	events, err := i.repo.OutgoingEvent().GetPendingEventsToRetry(ctx, model.OutgoingEventStatusPending.String(), inp.BatchSize)
	if err != nil {
		l.Error().Err(err).Msg("[ProcessPending] GetByStatus err")
		return err
	}

	if len(events) == 0 {
		l.Info().Msg("[ProcessPending] No outgoing events")
		return nil
	}

	for _, e := range events {
		logEvent := l.With().Field("event_id", e.ID).Field("topic", e.Topic)
		if val, ok := e.Payload.Data["short_code"]; ok {
			logEvent = logEvent.With().Field("short_code", val)
		}

		if err = i.processSingleEvent(ctx, e, inp.MaxRetry, logEvent); err != nil {
			l.Error().Err(err).Msg("[ProcessPending] processSingleEvent err")
		}
	}

	return nil
}

// processSingleEvent handles the publish lifecycle of a single outbox event:
//  1. Publish to Kafka.
//  2. On success, mark as PUBLISHED.
//  3. On failure, increment retry counter or mark as FAILED.
func (i impl) processSingleEvent(
	ctx context.Context,
	m model.OutgoingEvent,
	maxRetry int,
	l monitoring.Logger,
) error {
	payload, err := json.Marshal(m.Payload)
	if err != nil {
		return err
	}

	// publish
	if err = i.producer.Publish(ctx, m.Topic, payload); err != nil {
		// If we have reached the maximum retry count, mark the event as failed.
		errMsg := err.Error()
		nextRetry := m.RetryCount + 1

		// retry reached
		if nextRetry > maxRetry {
			l.Warn().
				Int("retry", nextRetry).
				Str("error", errMsg).
				Msg("[processSingleEvent] retry count exceeded, update status FAILED and last err to db")
			u := model.OutgoingEvent{LastError: errMsg, Status: model.OutgoingEventStatusFailed}
			if errMark := i.repo.OutgoingEvent().Update(ctx, u, m.ID); errMark != nil {
				l.Error().Err(errMark).Msg("[processSingleEvent] mark FAILED error")
				return errMark
			}

			return nil
		}

		// retry
		l.Warn().
			Int("retry", nextRetry).
			Str("error", errMsg).
			Msg("[processSingleEvent] publish failed → increment retry")

		u := model.OutgoingEvent{LastError: errMsg, RetryCount: nextRetry}
		if errRetry := i.repo.OutgoingEvent().Update(ctx, u, m.ID); errRetry != nil {
			l.Error().Err(errRetry).Msg("[processSingleEvent] retry update error")
			return errRetry
		}

		return nil
	}

	// success
	if err = i.repo.OutgoingEvent().Update(ctx, model.OutgoingEvent{Status: model.OutgoingEventStatusPublished}, m.ID); err != nil {
		l.Error().Err(err).Msg("[processSingleEvent] mark PUBLISHED error")
		return err
	}

	return nil
}
