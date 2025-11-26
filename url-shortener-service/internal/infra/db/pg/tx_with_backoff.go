package pg

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgconn"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
)

// TxWithBackoff wraps a transactional function with retry logic.
// The flow is:
//
//  1. retryWithBackoff()
//  2. Each retry attempts withTx()
//  3. If fn returns a retryable error → retry
//  4. If fn returns a non-retryable error → stop immediately
//
// Logging is included to indicate retry attempts and transaction outcomes.
func TxWithBackoff(ctx context.Context, db boil.ContextBeginner, policy backoff.BackOff, fn func(context.Context, boil.ContextExecutor) error) error {
	l := monitoring.Log(ctx)

	if policy == nil {
		policy = ExponentialBackOff(5, 8*time.Second)
	}

	op := func() error {
		l.Debug().Msg("TxWithBackoff: starting transactional operation")
		err := withTx(ctx, db, l, fn)

		if err != nil {
			if isRetryable(err) {
				l.Warn().Err(err).Msg("TxWithBackoff: retryable error")
				return err
			}
			l.Error().Err(err).Msg("TxWithBackoff: non-retryable error")
			return backoff.Permanent(err)
		}

		return nil
	}

	return retryWithBackoff(ctx, policy, l, op)
}

// withTx executes the given function inside a database transaction.
// Any error returned by fn will cause the transaction to roll back.
// A panic inside fn is recovered, the transaction is rolled back,
// and the panic is re-thrown.
//
// This function does NOT perform retry logic.
// That responsibility belongs to higher-level wrappers (TxWithBackoff).
func withTx(ctx context.Context, db boil.ContextBeginner, log monitoring.Logger, fn func(context.Context, boil.ContextExecutor) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("withTx: failed to begin transaction")
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			log.Error().Interface("panic", p).Msg("withTx: panic recovered → rollback")
			panic(p)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback()
		log.Warn().Err(err).Msg("withTx: fn returned error → rollback")
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("withTx: commit failed")
		return err
	}

	log.Debug().Msg("withTx: committed successfully")
	return nil
}

// retryWithBackoff executes the given operation with an exponential backoff retry strategy.
// If operation returns a retryable error, the retry will continue according to the provided policy.
// If operation returns backoff.Permanent(err), retry stops immediately.
//
// Logs are included so failures are visible and traceable.
func retryWithBackoff(ctx context.Context, policy backoff.BackOff, l monitoring.Logger, operation func() error) error {
	return backoff.RetryNotify(
		operation,
		backoff.WithContext(policy, ctx),
		func(err error, d time.Duration) {
			l.Warn().
				Err(err).
				Dur("retry_in", d).
				Msg("retryWithBackoff: transient error → retrying")
		},
	)
}

// isRetryable determines whether a given error is safe to retry.
// It inspects Postgres error codes and common transient network errors.
//
// This separation allows cleaner retry logic and easier unit testing.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "40001", // serialization_failure
			"40P01": // deadlock_detected
			return true
		}
		return false
	}

	// Fallback: match common transient network errors
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset")
}
