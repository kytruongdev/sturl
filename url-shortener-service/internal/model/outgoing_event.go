package model

import (
	"encoding/json"
	"time"

	pkgerrors "github.com/pkg/errors"
)

// OutgoingEventStatus represents the processing status of an outbox event.
type OutgoingEventStatus string

// Topic represents a Kafka topic name.
type Topic string

const (
	// OutgoingEventStatusPending indicates the event is waiting to be published.
	OutgoingEventStatusPending OutgoingEventStatus = "PENDING"
	// OutgoingEventStatusPublished indicates the event has been successfully published to Kafka.
	OutgoingEventStatusPublished OutgoingEventStatus = "PUBLISHED"
	// OutgoingEventStatusFailed indicates the event failed after max retries.
	OutgoingEventStatusFailed OutgoingEventStatus = "FAILED"

	// TopicMetadataRequestedV1 is the Kafka topic for metadata request events.
	TopicMetadataRequestedV1 Topic = "urlshortener.metadata.requested.v1"
	// TopicMetadataCrawledV1 is the Kafka topic for metadata crawled events.
	TopicMetadataCrawledV1 Topic = "urlshortener.metadata.crawled.v1"
)

// String returns the string representation of the outgoing event status.
func (e OutgoingEventStatus) String() string {
	return string(e)
}

// String returns the string representation of the topic.
func (t Topic) String() string { return string(t) }

// OutgoingEvent represents the domain-level event pushed into the outgoing event.
type OutgoingEvent struct {
	ID            int64
	RetryCount    int
	LastError     string
	CorrelationID string
	TraceID       string
	SpanID        string
	Topic         Topic
	Payload       Payload
	Status        OutgoingEventStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Payload defines the structure stored inside the JSONB payload column.
type Payload struct {
	EventID       int64             `json:"event_id"` // Unique event identifier for idempotency
	CorrelationID string            `json:"correlation_id"`
	TraceID       string            `json:"trace_id"`
	SpanID        string            `json:"span_id"`
	OccurredAt    time.Time         `json:"occurred_at"` // Timestamp when the event occurred
	Data          map[string]string `json:"data"`        // Actual business payload (generic)
}

// MarshalPayload marshals Payload field to byte array
func (m OutgoingEvent) MarshalPayload() ([]byte, error) {
	b, err := json.Marshal(m.Payload)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}
	return b, nil
}
