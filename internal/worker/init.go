package worker

import (
	"context"
	"database/sql"
	"sync"

	"github.com/fazt-sh/fazt/internal/debug"
)

var (
	globalPool *Pool
	poolMu     sync.RWMutex
)

// Init initializes the global worker pool.
func Init(db *sql.DB) error {
	poolMu.Lock()
	defer poolMu.Unlock()

	if globalPool != nil {
		return nil // Already initialized
	}

	globalPool = NewPool(db, DefaultPoolConfig())

	debug.Log("worker", "initialized global worker pool")
	return nil
}

// InitWithConfig initializes the global worker pool with custom config.
func InitWithConfig(db *sql.DB, cfg PoolConfig) error {
	poolMu.Lock()
	defer poolMu.Unlock()

	if globalPool != nil {
		return nil
	}

	globalPool = NewPool(db, cfg)

	debug.Log("worker", "initialized global worker pool with custom config")
	return nil
}

// GetPool returns the global worker pool.
func GetPool() *Pool {
	poolMu.RLock()
	defer poolMu.RUnlock()
	return globalPool
}

// SetExecutor sets the executor for the global pool.
func SetExecutor(exec JobExecutor) {
	poolMu.RLock()
	defer poolMu.RUnlock()
	if globalPool != nil {
		globalPool.SetExecutor(exec)
	}
}

// SetListenerCountFunc sets the listener count function for idle timeout checking.
func SetListenerCountFunc(fn ListenerCountFunc) {
	poolMu.RLock()
	defer poolMu.RUnlock()
	if globalPool != nil {
		globalPool.SetListenerCountFunc(fn)
	}
}

// Shutdown gracefully shuts down the global worker pool.
func Shutdown(ctx context.Context) error {
	poolMu.Lock()
	defer poolMu.Unlock()

	if globalPool == nil {
		return nil
	}

	err := globalPool.Shutdown(ctx)
	globalPool = nil
	return err
}

// RestoreDaemons restores daemon jobs from the database.
func RestoreDaemons() error {
	poolMu.RLock()
	defer poolMu.RUnlock()

	if globalPool == nil {
		return nil
	}

	return globalPool.RestoreDaemons()
}

// Spawn creates a new job using the global pool.
func Spawn(appID, handler string, cfg JobConfig) (*Job, error) {
	poolMu.RLock()
	pool := globalPool
	poolMu.RUnlock()

	if pool == nil {
		return nil, ErrPoolNotInitialized
	}

	return pool.Spawn(appID, handler, cfg)
}

// Cancel cancels a job by ID.
func Cancel(jobID string) error {
	poolMu.RLock()
	pool := globalPool
	poolMu.RUnlock()

	if pool == nil {
		return ErrPoolNotInitialized
	}

	return pool.Cancel(jobID)
}

// Get returns a job by ID.
func Get(jobID string) (*Job, error) {
	poolMu.RLock()
	pool := globalPool
	poolMu.RUnlock()

	if pool == nil {
		return nil, ErrPoolNotInitialized
	}

	return pool.Get(jobID)
}

// List returns jobs matching the filter.
func List(appID string, status *JobStatus, limit int) ([]*Job, error) {
	poolMu.RLock()
	pool := globalPool
	poolMu.RUnlock()

	if pool == nil {
		return nil, ErrPoolNotInitialized
	}

	return pool.List(appID, status, limit)
}

// Stats returns current pool statistics.
func Stats() *PoolStats {
	poolMu.RLock()
	pool := globalPool
	poolMu.RUnlock()

	if pool == nil {
		return nil
	}

	stats := pool.Stats()
	return &stats
}
