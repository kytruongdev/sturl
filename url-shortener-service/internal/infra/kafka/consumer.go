package kafka

import (
	"context"
	"errors"
	"io"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	kafkago "github.com/segmentio/kafka-go"
)

// Consumer defines the minimal interface required to consume messages from Kafka.
type Consumer interface {
	// Start begins consuming messages in a separate goroutine.
	// It should return quickly; the processing loop runs in the background.
	Start(ctx context.Context) error
	// Close releases any resources held by the consumer.
	Close() error
}

// NewConsumer creates a new Kafka consumer for a single topic and consumer group.
// Each topic should have its own Consumer instance to avoid switch-case routing.
func NewConsumer(cfg Config, topic, groupID string, handler MessageHandler) Consumer {
	r := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: cfg.Brokers,
		GroupID: groupID,
		Topic:   topic,
		// NOTE: can tune MinBytes/MaxBytes later if needed.
	})

	return &readerConsumer{
		topic:   topic,
		groupID: groupID,
		reader:  r,
		handler: handler,
	}
}

// readerConsumer is a concrete implementation of Consumer using kafka-go Reader.
type readerConsumer struct {
	topic   string
	groupID string
	reader  *kafkago.Reader
	handler MessageHandler
}

// Start begins the consume loop in a background goroutine.
func (c *readerConsumer) Start(ctx context.Context) error {
	log := monitoring.Log(ctx).Field("kafka_topic", c.topic).Field("kafka_group_id", c.groupID)

	go func() {
		log.Info().Msg("[KafkaConsumer] started")
		defer log.Info().Msg("[KafkaConsumer] stopped")

		for {
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				// Graceful exit when context is canceled.
				if errors.Is(err, context.Canceled) {
					log.Info().Msg("[KafkaConsumer] context canceled, stopping")
					return
				}
				if errors.Is(err, io.EOF) {
					log.Info().Msg("[KafkaConsumer] EOF received, stopping")
					return
				}

				log.Error().Err(err).Msg("[KafkaConsumer] ReadMessage error")
				continue
			}

			if handleErr := c.handler.ConsumeMessage(ctx, msg); handleErr != nil {
				log.Error().
					Err(handleErr).
					Int("partition", msg.Partition).
					Int64("offset", msg.Offset).
					Msg("[ConsumeMessage] error")
			}
		}
	}()

	return nil
}

// Close closes the underlying Kafka reader and frees associated resources.
func (c *readerConsumer) Close() error {
	return c.reader.Close()
}
