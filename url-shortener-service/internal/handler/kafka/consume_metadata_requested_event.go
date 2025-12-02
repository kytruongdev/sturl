package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/id"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/kafka"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository"
	kafkago "github.com/segmentio/kafka-go"
)

var newIDFunc = id.New

// MetadataRequested creates a Kafka message handler for metadata requested events.
// It processes incoming metadata request events, simulates metadata crawling,
// and publishes a metadata crawled event to the outbox for downstream processing.
func MetadataRequested(
	repo repository.Registry,
) kafka.MessageHandler {
	return kafka.HandlerFunc(func(ctx context.Context, msg kafkago.Message) error {
		var payload model.Payload
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			monitoring.Log(ctx).Error().Err(err).Msg("[MetadataRequested] failed to unmarshal payload")
			return err
		}

		newCtx, err := monitoring.EnrichContextWithSpanMetadata(ctx, monitoring.SpanMetadata{
			TraceID:       payload.TraceID,
			SpanID:        payload.SpanID,
			CorrelationID: payload.CorrelationID,
		})
		if err != nil {
			return err
		}

		spanCtx, span := monitoring.Start(newCtx, "Consumer.ConsumeMessage | Topic: "+msg.Topic)
		defer monitoring.End(span, &err)

		log := monitoring.Log(spanCtx).
			Field("topic", msg.Topic).
			Field("partition", msg.Partition).
			Field("offset", msg.Offset).
			Field("event_id", payload.EventID)

		log.Info().Msg("[MetadataRequested] handling message")

		// TODO: implement actual metadata fetching/crawling logic here.
		time.Sleep(5 * time.Second)
		log.Info().Msg("[MetadataRequested] [TEST] => message handled successfully")

		_, err = repo.OutgoingEvent().Insert(spanCtx, model.OutgoingEvent{
			ID:            newIDFunc(),
			Topic:         model.TopicMetadataCrawledV1,
			Status:        model.OutgoingEventStatusPending,
			CorrelationID: payload.CorrelationID,
			TraceID:       payload.TraceID,
			SpanID:        payload.SpanID,
			Payload: model.Payload{
				EventID:    newIDFunc(),
				OccurredAt: time.Now().UTC(),
				Data:       map[string]string{
					// TODO: fulfill later
				},
			},
		})
		if err != nil {
			log.Error().Err(err).Msg("[Shorten] OutgoingEventRepo.Insert err")
			return err
		}

		return nil
	})
}
