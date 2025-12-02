package kafka

import (
	"context"

	kafkago "github.com/segmentio/kafka-go"
)

// MessageHandler defines how to process a single Kafka message.
type MessageHandler interface {
	ConsumeMessage(ctx context.Context, msg kafkago.Message) error
}

// HandlerFunc is a function type that implements MessageHandler.
type HandlerFunc func(ctx context.Context, msg kafkago.Message) error

// ConsumeMessage implements the MessageHandler interface by calling
// the underlying function with the provided context and message.
func (f HandlerFunc) ConsumeMessage(ctx context.Context, msg kafkago.Message) error {
	return f(ctx, msg)
}
