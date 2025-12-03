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
			// Target Kafka brokers
			Addr: kafkago.TCP(cfg.Brokers...),

			// Distribute load based on the broker with the least accumulated bytes.
			Balancer: &kafkago.LeastBytes{},
			// Wait for all in-sync replicas to acknowledge the write.
			RequiredAcks: kafkago.RequireAll,
			// Use synchronous writes to keep error handling simple and predictable.
			Async: false,

			// Compression for better throughput and lower network usage.
			Compression: kafkago.Snappy,

			// Batching configuration:
			// Writer will flush when any of these conditions is met:
			// - BatchSize messages
			// - BatchBytes bytes
			// - BatchTimeout elapsed
			BatchSize:    100,                   // up to 100 messages per batch
			BatchBytes:   1 * 1024 * 1024,       // or 1MB of messages
			BatchTimeout: 10 * time.Millisecond, // or every 10ms

			// Timeouts to avoid hanging indefinitely on broker/network issues.
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,

			// Internal retry configuration:
			// Writer will retry transient failures up to MaxAttempts times
			// before returning an error to the caller.
			MaxAttempts: 3,

			// Transport configuration, including ClientID and connection timeouts.
			Transport: &kafkago.Transport{
				ClientID:    cfg.ClientID,
				DialTimeout: 10 * time.Second,
				IdleTimeout: 9 * time.Minute,
			},
		},
	}
}

// Publish sends a single message to the specified topic using the underlying Kafka writer.
// It relies on kafka-go's internal batching and retry mechanisms.
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
