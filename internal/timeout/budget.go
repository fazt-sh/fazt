package timeout

import (
	"context"
	"fmt"
	"time"
)

// Config defines timeout budgets for serverless execution.
type Config struct {
	RequestTimeout     time.Duration // Overall request timeout (default: 10s)
	StorageOpTimeout   time.Duration // Max time for a single storage operation (default: 3s)
	MinOperationTime   time.Duration // Minimum time required to start an operation (default: 500ms)
	QueueAdmissionTime time.Duration // Minimum time required to enter write queue (default: 1s)
}

// DefaultConfig returns sensible defaults for serverless execution.
func DefaultConfig() Config {
	return Config{
		RequestTimeout:     10 * time.Second,
		StorageOpTimeout:   3 * time.Second,
		MinOperationTime:   500 * time.Millisecond,
		QueueAdmissionTime: 1 * time.Second,
	}
}

// Budget tracks remaining time for a request and enforces admission control.
type Budget struct {
	deadline time.Time
	config   Config
}

// NewBudget creates a Budget from a context and config.
// If the context has no deadline, uses config.RequestTimeout from now.
func NewBudget(ctx context.Context, cfg Config) *Budget {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(cfg.RequestTimeout)
	}
	return &Budget{deadline: deadline, config: cfg}
}

// Remaining returns time until the budget's deadline.
func (b *Budget) Remaining() time.Duration {
	return time.Until(b.deadline)
}

// CanStartOperation returns true if there's enough time to start a storage op.
func (b *Budget) CanStartOperation() bool {
	return b.Remaining() >= b.config.MinOperationTime
}

// CanEnterQueue returns true if there's enough time to wait in the write queue.
func (b *Budget) CanEnterQueue() bool {
	return b.Remaining() >= b.config.QueueAdmissionTime
}

// StorageContext creates a bounded context for one storage operation.
// The timeout is min(remaining/2, StorageOpTimeout) to leave room for more ops.
// Returns ErrInsufficientTime if there's not enough time to start.
func (b *Budget) StorageContext(parent context.Context) (context.Context, context.CancelFunc, error) {
	if !b.CanStartOperation() {
		return nil, nil, ErrInsufficientTime
	}

	remaining := b.Remaining()
	opTimeout := remaining / 2
	if opTimeout > b.config.StorageOpTimeout {
		opTimeout = b.config.StorageOpTimeout
	}

	ctx, cancel := context.WithTimeout(parent, opTimeout)
	return ctx, cancel, nil
}

// Config returns the timeout configuration.
func (b *Budget) Config() Config {
	return b.config
}

// ErrInsufficientTime is returned when there's not enough time remaining.
var ErrInsufficientTime = fmt.Errorf("insufficient time remaining")
