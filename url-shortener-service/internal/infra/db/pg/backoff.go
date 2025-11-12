package pg

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

// ExponentialBackOff returns the exponential backoff configuration based on Azure best practices
// Reference: https://docs.microsoft.com/en-us/azure/postgresql/concepts-connectivity#handling-transient-errors
func ExponentialBackOff(maxRetries uint64, maxElapsedTime time.Duration) backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	// 1. Wait for 5 seconds before your first retry. (for simplicity, we're just using backoff.InitialInterval to simulate)
	b.InitialInterval = 5 * time.Second
	b.RandomizationFactor = 0
	// 2. For each following retry, the increase the wait exponentially, up to 60 seconds.
	b.MaxElapsedTime = maxElapsedTime
	// 3. Set a max number of retries at which point your application considers the operation failed.
	return backoff.WithMaxRetries(b, maxRetries)
}
