package storage

import (
	"context"
	"strings"
	"time"
)

const (
	maxRetries     = 5
	initialBackoff = 20 * time.Millisecond
)

// withRetry executes an operation with exponential backoff on transient errors.
func withRetry(ctx context.Context, op func() error) error {
	backoff := initialBackoff
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		err := op()
		if err == nil {
			return nil
		}

		// Only retry on transient SQLite errors
		if !isRetryable(err) {
			return err
		}

		lastErr = err

		// Check if context is done before sleeping
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
		}
	}

	return lastErr
}

// isRetryable checks if an error is a transient SQLite error worth retrying.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "SQLITE_BUSY") ||
		strings.Contains(errStr, "database is locked") ||
		strings.Contains(errStr, "database table is locked")
}
