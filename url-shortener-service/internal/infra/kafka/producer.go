package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	kafkago "github.com/segmentio/kafka-go"
)

// Producer defines the minimal interface required to publish messages to Kafka.
type Producer interface {
	// Publish sends a message with the given payload to the specified topic.
	Publish(ctx context.Context, topic string, payload []byte) error
	// Close releases any resources held by the producer.
	Close() error
}

// writerProducer is a concrete implementation of Producer using kafka-go's Writer.
type writerProducer struct {
	writer *kafkago.Writer
}

// NewProducer constructs a new Kafka Producer using the provided configuration.
// The caller is responsible for calling Close() when the producer is no longer needed.
func NewProducer(cfg Config) Producer {
	return &writerProducer{
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(cfg.Brokers...),
			Balancer:     &kafkago.LeastBytes{},
			RequiredAcks: kafkago.RequireAll,
			Async:        false,
			// Optional compression
			Compression: kafkago.Snappy,
			Transport:   &kafkago.Transport{ClientID: cfg.ClientID},
		},
	}
}

// Publish sends a single message to the specified topic using the underlying Kafka writer.
func (p *writerProducer) Publish(ctx context.Context, topic string, payload []byte) error {
	start := time.Now()
	log := monitoring.Log(ctx)

	msg := kafkago.Message{
		Topic: topic,
		Value: payload,
		Time:  time.Now(),
	}

	err := p.writer.WriteMessages(ctx, msg)
	latency := time.Since(start)
	if err != nil {
		log.Error().
			Str("topic", topic).
			Dur("latency", latency).
			Int("msg_size", len(payload)).
			Err(err).
			Msg("[writerProducer.Publish] kafka publish failed")
		return fmt.Errorf("kafka: failed to publish message: %w", err)
	}

	log.Info().
		Str("topic", topic).
		Dur("latency", latency).
		Int("msg_size", len(payload)).
		Msg("[writerProducer.Publish] kafka publish success")

	return nil
}

// Close closes the underlying Kafka writer and frees associated resources.
func (p *writerProducer) Close() error {
	return p.writer.Close()
}
