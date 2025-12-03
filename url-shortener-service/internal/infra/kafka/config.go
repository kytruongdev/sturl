package kafka

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

// Config holds configuration values for connecting to and optimizing a Kafka cluster.
type Config struct {
	Brokers       []string // List of Kafka bootstrap brokers
	ClientID      string   // Logical name of this client
	WorkerCount   int      // Number of concurrent consumer workers (default: 10)
	MinBytes      int      // Minimum bytes to fetch per request (default: 10KB)
	MaxBytes      int      // Maximum bytes to fetch per request (default: 10MB)
	MaxWait       int      // Maximum wait time in ms for MinBytes (default: 1000ms)
	ChannelBuffer int      // Size of the message channel buffer (default: workerCount * 2)
}

// getIntEnv parses an integer from environment variable with a default fallback.
// Returns the parsed value if valid and positive, otherwise returns the default.
func getIntEnv(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultVal
}

func NewConfig() Config {
	workerCount := getIntEnv("KAFKA_CONSUMER_WORKERS", 10)
	minBytes := getIntEnv("KAFKA_MIN_BYTES", 10*1024)      // 10 KB
	maxBytes := getIntEnv("KAFKA_MAX_BYTES", 10*1024*1024) // 10 MB
	maxWait := getIntEnv("KAFKA_MAX_WAIT_MS", 1000)        // 1 second
	channelBuffer := getIntEnv("KAFKA_CHANNEL_BUFFER", 0)  // override
	clientID := os.Getenv("KAFKA_CLIENT_ID")

	if channelBuffer <= 0 {
		channelBuffer = workerCount * 2 // default rule
	}

	brokerStr := os.Getenv("KAFKA_BROKERS")
	brokers := []string{}
	for _, b := range strings.Split(brokerStr, ",") {
		b = strings.TrimSpace(b)
		if b != "" {
			brokers = append(brokers, b)
		}
	}

	return Config{
		Brokers:       brokers,
		ClientID:      clientID,
		WorkerCount:   workerCount,
		MinBytes:      minBytes,
		MaxBytes:      maxBytes,
		MaxWait:       maxWait,
		ChannelBuffer: channelBuffer,
	}
}

func (c Config) Validate() error {
	if len(c.Brokers) == 0 {
		return errors.New("missing KAFKA_BROKERS")
	}
	if c.WorkerCount <= 0 {
		return errors.New("WorkerCount must be > 0")
	}
	return nil
}
