package model

import (
	"encoding/json"
	"time"

	pkgerrors "github.com/pkg/errors"
)

type OutgoingEventStatus string
type Topic string

const (
	OutgoingEventStatusPending   OutgoingEventStatus = "PENDING"
	OutgoingEventStatusPublished OutgoingEventStatus = "PUBLISHED"
	OutgoingEventStatusFailed    OutgoingEventStatus = "FAILED"

	TopicMetadataRequestedV1 Topic = "urlshortener.metadata.requested.v1"
	TopicMetadataCrawledV1   Topic = "urlshortener.metadata.crawled.v1"
)

func (e OutgoingEventStatus) String() string {
	return string(e)
}

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
