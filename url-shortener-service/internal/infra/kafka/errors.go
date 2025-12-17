package kafka

// KafkaError wraps an error with retry information for Kafka message processing.
// It indicates whether the error is transient (retryable) or permanent (non-retryable).
type KafkaError struct {
	error
	canRetry bool
}

// NewKafkaError creates a new KafkaError with the given error and retry flag.
func NewKafkaError(err error, canRetry bool) *KafkaError {
	return &KafkaError{
		error:    err,
		canRetry: canRetry,
	}
}

// isRetryable returns true if the error can be retried with backoff.
func (e KafkaError) isRetryable() bool {
	return e.canRetry
}

// Unwrap returns the underlying error for error chain unwrapping.
func (e KafkaError) Unwrap() error {
	return e.error
}
