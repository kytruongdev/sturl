package kafka

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	kafkago "github.com/segmentio/kafka-go"
)

const (
	MaxRetries         = 4
	DefaultWorkerCount = 10

	// ChannelBufferFactor = 2 means:
	// bufferSize = workerCount * 2
	// → ensures workers always have pending messages
	//   and prevents worker starvation when fetcher is briefly blocked.
	ChannelBufferFactor = 2
)

// Consumer defines the minimal interface required to consume messages from Kafka.
type Consumer interface {
	// Start begins consuming messages in a separate goroutine.
	Start(ctx context.Context) error
	// Close releases any resources held by the consumer.
	Close() error
}

// NewConsumer creates a new Kafka consumer for a single topic and consumer group.
func NewConsumer(cfg Config, topic, groupID string, handler MessageHandler, dlqProducer Producer) Consumer {
	workerCount := cfg.WorkerCount
	if workerCount <= 0 {
		workerCount = DefaultWorkerCount
	}

	// Channel buffer = workerCount * 2 by default
	// Can be overridden via cfg.CommitBuffer if needed.
	channelBuffer := workerCount * ChannelBufferFactor
	if cfg.ChannelBuffer > 0 {
		channelBuffer = cfg.ChannelBuffer
	}

	// Kafka reader performance tuning
	minBytes := cfg.MinBytes
	if minBytes <= 0 {
		minBytes = 10e3 // 10KB
	}

	maxBytes := cfg.MaxBytes
	if maxBytes <= 0 {
		maxBytes = 10e6 // 10MB
	}

	maxWait := cfg.MaxWait
	if maxWait <= 0 {
		maxWait = 1000 // 1s
	}

	return &readerConsumer{
		handler:       handler,
		dlqProducer:   dlqProducer,
		topic:         topic,
		groupID:       groupID,
		workerCount:   workerCount,
		channelBuffer: channelBuffer,
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:  cfg.Brokers,
			GroupID:  groupID,
			Topic:    topic,
			MinBytes: minBytes,
			MaxBytes: maxBytes,
			MaxWait:  time.Duration(maxWait) * time.Millisecond,
			// avoid reprocessing old messages on first startup
			StartOffset: kafkago.LastOffset,
			// prevent long hangs on broker issues
			ReadBackoffMin: 100 * time.Millisecond,
			ReadBackoffMax: 1 * time.Second,
		}),
	}
}

type readerConsumer struct {
	topic         string
	groupID       string
	reader        *kafkago.Reader
	handler       MessageHandler
	dlqProducer   Producer
	workerCount   int
	channelBuffer int
}

// Start begins the consume loop in a background goroutine.
func (c *readerConsumer) Start(ctx context.Context) error {
	log := monitoring.Log(ctx).
		Field("topic", c.topic).
		Field("group_id", c.groupID).
		Field("workers", c.workerCount).
		Field("channel_buffer", c.channelBuffer)

	// shared queue for all workers
	msgChan := make(chan kafkago.Message, c.channelBuffer)

	// start worker pool
	for i := 0; i < c.workerCount; i++ {
		go c.worker(ctx, msgChan, i)
	}

	// fetcher goroutine: continuously fetch messages from Kafka
	go func() {
		log.Info().Msg("[KafkaConsumer] fetcher started")
		defer log.Info().Msg("[KafkaConsumer] fetcher stopped")
		defer close(msgChan) // closing the channel gracefully shuts down workers

		for {
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					log.Info().Msg("[KafkaConsumer] context canceled, stopping fetcher")
					return
				}
				if errors.Is(err, io.EOF) {
					log.Info().Msg("[KafkaConsumer] EOF, stopping fetcher")
					return
				}
				log.Error().Err(err).Msg("[KafkaConsumer] FetchMessage error")
				// small backoff to avoid spamming the broker on repeated errors
				time.Sleep(100 * time.Millisecond)
				continue
			}

			log.Info().
				Str("topic", msg.Topic).
				Int("partition", msg.Partition).
				Int64("offset", msg.Offset).
				Msg("[KafkaConsumer] partition assigned")

			// send message to workers (blocks if channel is full → natural backpressure)
			select {
			case msgChan <- msg:
				// successfully queued
			case <-ctx.Done():
				log.Info().Msg("[KafkaConsumer] context canceled while sending to channel")
				return
			}
		}
	}()

	log.Info().Msg("[KafkaConsumer] started")
	return nil
}

// worker processes messages concurrently.
// Workers read from the shared channel until:
// - channel is closed, or
// - context is canceled.
func (c *readerConsumer) worker(ctx context.Context, msgChan <-chan kafkago.Message, workerID int) {
	log := monitoring.Log(ctx).
		Field("worker_id", workerID).
		Field("topic", c.topic)

	log.Info().Msg("[KafkaConsumer-Worker] started")
	defer log.Info().Msg("[KafkaConsumer-Worker] stopped")

	var processedCount int
	var errorCount int

	// Periodic worker health metrics
	statsTicker := time.NewTicker(1 * time.Minute)
	defer statsTicker.Stop()

	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				// channel closed → fetcher stopped → safe shutdown
				log.Info().
					Int("processed", processedCount).
					Int("errors", errorCount).
					Msg("[KafkaConsumer-Worker] channel closed, exiting")
				return
			}

			// add per-message timeout to prevent long-running tasks from blocking the worker forever
			msgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			processErr := c.processMessage(msgCtx, msg)
			cancel()

			if processErr != nil {
				errorCount++

				log.Error().
					Err(processErr).
					Int("partition", msg.Partition).
					Int64("offset", msg.Offset).
					Msg("[KafkaConsumer-Worker] error occurred → sending to DLQ")

				meta := monitoring.SpanMetadataFromContext(msgCtx)

				dlqMsg := DLQMessage{
					Topic:         c.topic,
					Payload:       msg.Value,
					Error:         processErr.Error(),
					TraceID:       meta.TraceID,
					SpanID:        meta.SpanID,
					CorrelationID: meta.CorrelationID,
					Partition:     msg.Partition,
					Offset:        msg.Offset,
					Timestamp:     time.Now(),
				}

				if err := publishDLQ(msgCtx, c.dlqProducer, dlqMsg, c.topic+".dlq"); err != nil {
					log.Error().Err(err).Msg("[KafkaConsumer-Worker] DLQ publish failed")
				}

				// Commit offset to avoid poison message loop
				if err := c.reader.CommitMessages(ctx, msg); err != nil {
					log.Error().
						Err(err).
						Int("partition", msg.Partition).
						Int64("offset", msg.Offset).
						Msg("[KafkaConsumer-Worker] commit failed after DLQ")
				}

				continue
			}

			// Successful processing → commit offset
			processedCount++

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Error().
					Err(err).
					Int("partition", msg.Partition).
					Int64("offset", msg.Offset).
					Msg("[KafkaConsumer-Worker] commit failed")
			}

		case <-statsTicker.C:
			total := processedCount + errorCount
			errorRate := 0.0
			if total > 0 {
				errorRate = float64(errorCount) / float64(total) * 100
			}

			log.Debug().
				Int("processed", processedCount).
				Int("errors", errorCount).
				Float64("error_rate_pct", errorRate).
				Msg("[KafkaConsumer-Worker] stats")

		case <-ctx.Done():
			log.Info().
				Int("processed", processedCount).
				Int("errors", errorCount).
				Msg("[KafkaConsumer-Worker] context canceled, exiting")
			return
		}
	}
}

// processMessage handles retry logic and returns the final error (if any).
// Worker is responsible for DLQ logic and committing offsets.
func (c *readerConsumer) processMessage(ctx context.Context, msg kafkago.Message) error {
	log := monitoring.Log(ctx)

	// First attempt
	err := c.handler.ConsumeMessage(ctx, msg)
	if err == nil {
		return nil
	}

	// Non-retryable → return immediately
	if !err.isRetryable() {
		log.Warn().
			Err(err).
			Int("partition", msg.Partition).
			Int64("offset", msg.Offset).
			Msg("[KafkaConsumer] non-retryable error")
		return err.error
	}

	// Retryable → apply backoff policy
	policy := backoff.WithMaxRetries(DefaultBackoffPolicy(), MaxRetries)

	lastErr := retryKafka(ctx, policy, func(ctx context.Context) error {
		return c.handler.ConsumeMessage(ctx, msg)
	})

	if lastErr != nil {
		log.Warn().
			Err(lastErr).
			Int("partition", msg.Partition).
			Int64("offset", msg.Offset).
			Msg("[KafkaConsumer] retry exhausted")
	}

	return lastErr
}

// Close closes the underlying Kafka reader and frees associated resources.
func (c *readerConsumer) Close() error {
	return c.reader.Close()
}
