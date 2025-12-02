package main

import (
	"context"
	"encoding/json"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
)

// process loads a batch of pending events and attempts to publish
// each event to Kafka, updating their status accordingly.
//
// It is designed to be called repeatedly by the producer loop.
func (p *Producer) process(ctx context.Context) error {
	log := monitoring.Log(ctx)

	events, err := p.repo.OutgoingEvent().
		GetPendingEventsToRetry(ctx, model.OutgoingEventStatusPending.String(), p.config.batchSize)
	if err != nil {
		log.Error().Err(err).Msg("[processPendingBatch] GetPendingEventsToRetry err")
		return err
	}

	if len(events) == 0 {
		log.Info().Msg("[processPendingBatch] No outgoing events")
		return nil
	}

	for _, e := range events {
		producerCtx, err := toProducerContext(e)
		if err != nil {
			log.Error().Err(err).Msg("[processPendingBatch] newContextFromOutbox err")
			producerCtx = ctx
		}

		p.publishMessageToKafka(producerCtx, e)
	}

	return nil
}

// publishMessageToKafka handles the publish lifecycle of a single outbox event:
//  1. Publish to Kafka.
//  2. On success, mark as PUBLISHED.
//  3. On failure, increment retry counter or mark as FAILED.
func (p *Producer) publishMessageToKafka(
	ctx context.Context,
	m model.OutgoingEvent,
) {
	var err error
	spanCtx, span := monitoring.Start(ctx, "Producer.PublishMessage | Topic: "+m.Topic.String())
	defer monitoring.End(span, &err)

	log := monitoring.Log(spanCtx).
		Field("event_id", m.ID).
		Field("topic", m.Topic)

	if val, ok := m.Payload.Data["short_code"]; ok {
		log = log.Field("short_code", val)
	}

	// Ensure payload carries trace + correlation info
	m.Payload.CorrelationID = m.CorrelationID
	m.Payload.TraceID = m.TraceID
	m.Payload.SpanID = m.SpanID

	payload, err := json.Marshal(m.Payload)
	if err != nil {
		log.Error().Err(err).Msg("[publishMessageToKafka] json.Marshal err")
		return
	}

	// publish
	if err = p.producer.Publish(spanCtx, m.Topic.String(), payload); err != nil {
		errMsg := err.Error()
		nextRetry := m.RetryCount + 1

		// retry reached
		if nextRetry > p.config.maxRetry {
			log.Warn().
				Int("retry", nextRetry).
				Str("error", errMsg).
				Msg("[publishMessageToKafka] retry count exceeded, update status FAILED and last err to db")
			u := model.OutgoingEvent{
				LastError: errMsg,
				Status:    model.OutgoingEventStatusFailed,
			}
			if err = p.repo.OutgoingEvent().Update(spanCtx, u, m.ID); err != nil {
				log.Error().Err(err).Msg("[publishMessageToKafka] mark FAILED error")
			}
			return
		}

		// retry
		log.Warn().
			Int("retry", nextRetry).
			Str("error", errMsg).
			Msg("[publishMessageToKafka] publish failed → increment retry")

		u := model.OutgoingEvent{
			LastError:  errMsg,
			RetryCount: nextRetry,
		}
		if err = p.repo.OutgoingEvent().Update(spanCtx, u, m.ID); err != nil {
			log.Error().Err(err).Msg("[publishMessageToKafka] retry update error")
		}
		return
	}

	// success
	if err = p.repo.OutgoingEvent().Update(
		spanCtx,
		model.OutgoingEvent{Status: model.OutgoingEventStatusPublished},
		m.ID,
	); err != nil {
		log.Error().Err(err).Msg("[publishMessageToKafka] mark PUBLISHED error")
	}
}

// toProducerContext reconstructs a remote parent span context from an outgoing event.
// It extracts trace ID, span ID, and correlation ID from the event and creates
// a new context with the remote span context for distributed tracing.
func toProducerContext(ev model.OutgoingEvent) (context.Context, error) {
	return monitoring.NewContextFromSpanMetadata(context.Background(), monitoring.SpanMetadata{
		TraceID:       ev.TraceID,
		SpanID:        ev.SpanID,
		CorrelationID: ev.CorrelationID,
	})
}
