// internal/handler/kafka/consumer.go
package main

import (
	"context"
	"os"

	shortUrlCtrl "github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/handler/kafka"
	infraKafka "github.com/kytruongdev/sturl/url-shortener-service/internal/infra/kafka"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
)

// Consumer manages multiple Kafka consumers for different topics.
// It coordinates the lifecycle of all consumers and provides a unified
// interface for starting and stopping them.
type Consumer struct {
	consumers map[string]infraKafka.Consumer
}

// New creates a new Consumer instance with handlers for all configured topics.
func New(cfg infraKafka.Config, shortURLCtrl shortUrlCtrl.Controller, producer infraKafka.Producer) Consumer {
	consumers := make(map[string]infraKafka.Consumer)
	consumers[model.TopicMetadataRequestedV1.String()] = infraKafka.NewConsumer(
		cfg,
		model.TopicMetadataRequestedV1.String(),
		os.Getenv("METADATA_REQUESTED_CONSUMER_GROUP"),
		kafka.MetadataRequested(shortURLCtrl),
		producer,
	)
	consumers[model.TopicMetadataCrawledV1.String()] = infraKafka.NewConsumer(
		cfg,
		model.TopicMetadataCrawledV1.String(),
		os.Getenv("METADATA_CRAWLED_CONSUMER_GROUP"),
		kafka.MetadataCrawled(),
		producer,
	)

	return Consumer{
		consumers: consumers,
	}
}

// Start begins consuming messages from all configured Kafka topics.
// It starts each consumer in the background and blocks until the context is canceled.
// Returns an error if any consumer fails to start.
func (m *Consumer) Start(ctx context.Context) error {
	log := monitoring.Log(ctx)
	for k, c := range m.consumers {
		log.Info().Msg("[Consumer.Start] starting consumer with topic: " + k)
		if err := c.Start(ctx); err != nil {
			log.Error().Err(err).Msg("[KafkaManager] consumer failed to start")
			return err
		}
	}

	monitoring.Log(ctx).Info().
		Int("count", len(m.consumers)).
		Msg("[KafkaManager] all consumers started")

	<-ctx.Done()
	return nil
}

// Stop gracefully shuts down all Kafka consumers.
// It closes each consumer and returns any errors encountered during shutdown.
func (m *Consumer) Stop(ctx context.Context) error {
	for _, c := range m.consumers {
		_ = c.Close()
	}
	return nil
}
