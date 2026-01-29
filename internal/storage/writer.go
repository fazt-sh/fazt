package storage

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// WriteQueue serializes all write operations to prevent SQLITE_BUSY errors.
// All storage writes go through this queue, processed by a single goroutine.
type WriteQueue struct {
	queue    chan writeOp
	queueLen int32 // atomic counter for monitoring
	maxQueue int
	done     chan struct{}
	wg       sync.WaitGroup
}

type writeOp struct {
	fn   func() error
	done chan error
	ctx  context.Context
}

// WriteQueueConfig configures the write queue.
type WriteQueueConfig struct {
	// QueueSize is the max pending writes before blocking/rejecting.
	// Default: 1000
	QueueSize int

	// Workers is the number of write workers. Keep at 1 for SQLite.
	// Default: 1
	Workers int
}

// DefaultWriteQueueConfig returns sensible defaults.
func DefaultWriteQueueConfig() WriteQueueConfig {
	return WriteQueueConfig{
		QueueSize: 1000,
		Workers:   1, // SQLite only supports 1 writer
	}
}

// NewWriteQueue creates a new serialized write queue.
func NewWriteQueue(cfg WriteQueueConfig) *WriteQueue {
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 1000
	}
	if cfg.Workers <= 0 {
		cfg.Workers = 1
	}

	wq := &WriteQueue{
		queue:    make(chan writeOp, cfg.QueueSize),
		maxQueue: cfg.QueueSize,
		done:     make(chan struct{}),
	}

	// Start worker(s)
	for i := 0; i < cfg.Workers; i++ {
		wq.wg.Add(1)
		go wq.worker()
	}

	return wq
}

// worker processes writes sequentially.
func (wq *WriteQueue) worker() {
	defer wq.wg.Done()

	for {
		select {
		case op := <-wq.queue:
			atomic.AddInt32(&wq.queueLen, -1)

			// Check if context already cancelled
			select {
			case <-op.ctx.Done():
				op.done <- op.ctx.Err()
				continue
			default:
			}

			// Execute the write
			err := op.fn()
			op.done <- err

		case <-wq.done:
			return
		}
	}
}

// Write queues a write operation and waits for completion.
// Returns ErrQueueFull if the queue is at capacity.
// Returns ErrInsufficientTime if deadline won't allow operation to complete.
func (wq *WriteQueue) Write(ctx context.Context, fn func() error) error {
	// Admission control: check if we have enough time before queueing
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		// Estimate queue wait based on current depth (30ms per queued op estimate)
		queueDepth := atomic.LoadInt32(&wq.queueLen)
		estimatedWait := time.Duration(queueDepth) * 30 * time.Millisecond
		// Need estimated wait time + 500ms minimum for the operation itself
		minRequired := estimatedWait + 500*time.Millisecond

		if remaining < minRequired {
			return &StorageError{
				Op:        "write_queue",
				Cause:     fmt.Errorf("insufficient time: need %v, have %v (queue depth: %d)", minRequired, remaining, queueDepth),
				Retryable: true,
			}
		}
	}

	done := make(chan error, 1)
	op := writeOp{
		fn:   fn,
		done: done,
		ctx:  ctx,
	}

	// Try to queue the operation
	select {
	case wq.queue <- op:
		atomic.AddInt32(&wq.queueLen, 1)
	default:
		// Queue is full
		return ErrQueueFull
	}

	// Wait for completion or context cancellation
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// QueueDepth returns the current number of pending writes.
func (wq *WriteQueue) QueueDepth() int {
	return int(atomic.LoadInt32(&wq.queueLen))
}

// QueueCapacity returns the max queue size.
func (wq *WriteQueue) QueueCapacity() int {
	return wq.maxQueue
}

// Close stops the write queue gracefully.
func (wq *WriteQueue) Close() {
	close(wq.done)
	wq.wg.Wait()
}

// StorageError represents a storage-specific error with context.
type StorageError struct {
	Op         string // Operation: "insert", "update", "delete", etc.
	Collection string // Collection/table name
	Cause      error  // Underlying error
	Retryable  bool   // Whether the client should retry
}

func (e *StorageError) Error() string {
	if e.Collection != "" {
		return fmt.Sprintf("storage.%s(%s): %v", e.Op, e.Collection, e.Cause)
	}
	return fmt.Sprintf("storage.%s: %v", e.Op, e.Cause)
}

func (e *StorageError) Unwrap() error {
	return e.Cause
}

// Common storage errors
var (
	ErrQueueFull = &StorageError{
		Op:        "write",
		Cause:     fmt.Errorf("write queue full - server overloaded"),
		Retryable: true,
	}

	ErrTimeout = &StorageError{
		Op:        "write",
		Cause:     fmt.Errorf("operation timed out"),
		Retryable: true,
	}
)

// IsRetryableError checks if an error is retryable by the client.
func IsRetryableError(err error) bool {
	if se, ok := err.(*StorageError); ok {
		return se.Retryable
	}
	return isRetryable(err)
}

// WriteStats holds write queue statistics.
type WriteStats struct {
	QueueDepth    int     `json:"queue_depth"`
	QueueCapacity int     `json:"queue_capacity"`
	Utilization   float64 `json:"utilization"` // 0.0 - 1.0
}

// Stats returns current write queue statistics.
func (wq *WriteQueue) Stats() WriteStats {
	depth := wq.QueueDepth()
	capacity := wq.QueueCapacity()
	util := 0.0
	if capacity > 0 {
		util = float64(depth) / float64(capacity)
	}
	return WriteStats{
		QueueDepth:    depth,
		QueueCapacity: capacity,
		Utilization:   util,
	}
}
