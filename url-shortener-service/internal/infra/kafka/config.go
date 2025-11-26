package kafka

import (
	"os"
	"strings"
)

// Config holds configuration values for connecting to a Kafka cluster.
type Config struct {
	Brokers  []string // Brokers is the list of Kafka bootstrap brokers (e.g. []string{"localhost:9092"}).
	ClientID string   // ClientID is the logical name of this producer client.
}

func NewConfig() Config {
	return Config{
		Brokers:  strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
		ClientID: os.Getenv("KAFKA_CLIENT_ID"),
	}
}

func (c Config) Validate() error {
	// TODO: validate later
	return nil
}
