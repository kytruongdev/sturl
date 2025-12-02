package kafka

import (
	"context"
	"encoding/json"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/kafka"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	kafkago "github.com/segmentio/kafka-go"
)

// MetadataCrawled handles "urlshortener.metadata.crawled.v1".
func MetadataCrawled(
// TODO: inject notify service later
) kafka.MessageHandler {
	return kafka.HandlerFunc(func(ctx context.Context, msg kafkago.Message) error {
		var payload model.Payload
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			monitoring.Log(ctx).Error().Err(err).
				Msg("[MetadataCrawled] failed to unmarshal payload")
			return err
		}

		// Rebuild tracing/correlation context
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

		log.Info().Msg("[MetadataCrawled] handling message")

		if err = notifyMetadataCrawled(spanCtx); err != nil {
			log.Error().Err(err).Msg("[MetadataCrawled] failed to notify user")
			return err
		}

		log.Info().Msg("[MetadataCrawled] notification success")
		return nil
	})
}

// notifyMetadataCrawled handles notification logic when metadata has been crawled.
// Currently a placeholder that will be implemented to notify users or downstream services.
func notifyMetadataCrawled(ctx context.Context) error {
	return nil
}
