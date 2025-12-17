package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	shortUrlCtrl "github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/id"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/kafka"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	kafkago "github.com/segmentio/kafka-go"
)

var newIDFunc = id.New

// MetadataRequested creates a Kafka message handler for metadata requested events.
// It processes incoming metadata request events, simulates metadata crawling,
// and publishes a metadata crawled event to the outbox for downstream processing.
func MetadataRequested(
	shortURLCtrl shortUrlCtrl.Controller,
) kafka.MessageHandler {
	return kafka.HandlerFunc(func(ctx context.Context, msg kafkago.Message) *kafka.KafkaError {
		var payload model.Payload
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			monitoring.Log(ctx).Error().Err(err).Msg("[MetadataRequested] failed to unmarshal payload")
			return kafka.NewKafkaError(err, false)
		}

		shortCode := payload.Data["short_code"]
		if shortCode == "" {
			return kafka.NewKafkaError(errors.New("[MetadataRequested] short_code is empty"), false)
		}

		metadata := monitoring.SpanMetadata{
			TraceID:       payload.TraceID,
			SpanID:        payload.SpanID,
			CorrelationID: payload.CorrelationID,
		}
		fmt.Printf("[MetadataRequested] metadata: %+v\n", metadata)
		newCtx, err := monitoring.EnrichContextWithSpanMetadata(ctx, metadata)
		if err != nil {
			return kafka.NewKafkaError(err, false)
		}

		spanCtx, span := monitoring.Start(newCtx, "Consumer.ConsumeMessage | Topic: "+msg.Topic)
		defer monitoring.End(span, &err)

		log := monitoring.Log(spanCtx).
			Field("topic", msg.Topic).
			Field("partition", msg.Partition).
			Field("offset", msg.Offset).
			Field("event_id", payload.EventID)

		log.Info().Msg("[MetadataRequested] handling message")

		_, err = shortURLCtrl.CrawlURLMetadata(spanCtx, shortCode)
		if err != nil {
			log.Error().Err(err).Msg("[MetadataRequested] failed to crawl url")
			return kafka.NewKafkaError(err, true)
		}

		log.Info().Msg("[MetadataRequested] metadata crawled successfully")

		return nil
	})
}
