package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
)

// DLQMessage represents a failed Kafka message that is sent to a Dead Letter Queue.
// It contains the original message payload, error details, and tracing metadata
// for debugging and potential reprocessing.
type DLQMessage struct {
	Topic         string          `json:"topic"`          // Original topic where the message failed
	Payload       json.RawMessage `json:"payload"`        // Original message payload
	Error         string          `json:"error"`          // Error message that caused the failure
	TraceID       string          `json:"trace_id"`       // OpenTelemetry trace ID for correlation
	SpanID        string          `json:"span_id"`        // OpenTelemetry span ID for correlation
	CorrelationID string          `json:"correlation_id"` // Correlation ID for request tracking
	Partition     int             `json:"partition"`      // Kafka partition where message was consumed
	Offset        int64           `json:"offset"`         // Kafka offset of the failed message
	Timestamp     time.Time       `json:"timestamp"`      // When the message failed
}

// publishDLQ sends a failed message to the Dead Letter Queue topic.
// This is an internal function used by the consumer to route failed messages.
func publishDLQ(ctx context.Context, producer Producer, msg DLQMessage, targetTopic string) error {
	log := monitoring.Log(ctx).Field("dlq_topic", targetTopic)

	raw, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("[DLQ] marshal failed")
		return err
	}

	if err := producer.Publish(ctx, targetTopic, raw); err != nil {
		log.Error().Err(err).Msg("[DLQ] publish failed")
		return err
	}

	log.Info().Msg("[DLQ] message sent")
	return nil
}
