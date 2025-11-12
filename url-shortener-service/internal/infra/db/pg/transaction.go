package pg

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgconn"
)

// TxWithBackoff retries a DB transaction on transient errors (deadlock, serialization failure, etc.)
// using exponential backoff (github.com/cenkalti/backoff/v5).
func TxWithBackoff(ctx context.Context, db boil.ContextBeginner, policy backoff.BackOff, fn func(ctx context.Context, exec boil.ContextExecutor) error) error {
	operation := func() error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		defer func() {
			if p := recover(); p != nil {
				_ = tx.Rollback()
				panic(p)
			}
		}()

		// Use tx as the executor for sqlboiler
		if err = fn(ctx, tx); err != nil {
			_ = tx.Rollback()
			if isRetryable(err) {
				return err
			}

			return backoff.Permanent(err)
		}

		return tx.Commit()
	}

	// Backoff policy
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 100 * time.Millisecond
	b.MaxInterval = 2 * time.Second
	b.MaxElapsedTime = 8 * time.Second

	bCtx := backoff.WithContext(policy, ctx)
	return backoff.Retry(operation, bCtx)
}

func isRetryable(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "40001": // serialization_failure
		case "40P01": // deadlock_detected
			return true
		}
		return false
	}
	// fallback cho error mạng
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "timeout")
}
