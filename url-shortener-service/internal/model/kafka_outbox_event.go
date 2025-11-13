package model

import (
	"encoding/json"
	"time"

	pkgerrors "github.com/pkg/errors"
)

type KafkaOutboxEventStatus string

const (
	KafkaOutboxEventStatusPending KafkaOutboxEventStatus = "PENDING"
	KafkaOutboxEventStatusSent    KafkaOutboxEventStatus = "SENT"
	KafkaOutboxEventStatusFailed  KafkaOutboxEventStatus = "FAILED"
)

func (e KafkaOutboxEventStatus) String() string {
	return string(e)
}

// KafkaOutboxEvent represents the domain-level event pushed into the outbox.
// It contains business meaning and is used as input for the outbox repository.
type KafkaOutboxEvent struct {
	ID        int64
	EventType string
	Payload   Payload
	Status    KafkaOutboxEventStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Payload defines the structure stored inside the JSONB payload column.
// This represents the business-level event information consumed by Kafka downstream services.
type Payload struct {
	EventID    int64             `json:"event_id"`    // Unique event identifier for idempotency
	OccurredAt time.Time         `json:"occurred_at"` // Timestamp when the event occurred
	Data       map[string]string `json:"data"`        // Actual business payload (generic)
}

// MarshalPayload marshals Payload field to byte array
func (m KafkaOutboxEvent) MarshalPayload() ([]byte, error) {
	b, err := json.Marshal(m.Payload)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}
	return b, nil
}
