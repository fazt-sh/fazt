package worker

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fazt-sh/fazt/internal/debug"
)

// ResourceBudget tracks memory allocation for the worker pool.
// It provides soft limits using runtime.MemStats monitoring.
type ResourceBudget struct {
	// Configuration
	poolSize   int64 // Total pool size in bytes
	warnLevel  float64 // Warning threshold (0.8 = 80%)

	// Allocation tracking
	allocated int64 // Atomically updated

	// MemStats monitoring
	lastHeapAlloc uint64
	lastCheck     time.Time

	// Lifecycle
	done chan struct{}
	wg   sync.WaitGroup

	// Warnings state
	warnedRecently int32 // atomic: 1 if warned in last minute
}

// BudgetConfig configures the resource budget.
type BudgetConfig struct {
	PoolSize   int64   // Total memory pool in bytes (default: 256MB)
	WarnLevel  float64 // Warning threshold (default: 0.8)
	CheckInterval time.Duration // MemStats check interval (default: 100ms)
}

// DefaultBudgetConfig returns sensible defaults.
func DefaultBudgetConfig() BudgetConfig {
	return BudgetConfig{
		PoolSize:      256 * 1024 * 1024, // 256MB
		WarnLevel:     0.8,
		CheckInterval: 100 * time.Millisecond,
	}
}

// NewResourceBudget creates a new resource budget tracker.
func NewResourceBudget(cfg BudgetConfig) *ResourceBudget {
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = 256 * 1024 * 1024
	}
	if cfg.WarnLevel <= 0 || cfg.WarnLevel > 1 {
		cfg.WarnLevel = 0.8
	}
	if cfg.CheckInterval <= 0 {
		cfg.CheckInterval = 100 * time.Millisecond
	}

	rb := &ResourceBudget{
		poolSize:  cfg.PoolSize,
		warnLevel: cfg.WarnLevel,
		done:      make(chan struct{}),
	}

	// Start monitor goroutine
	rb.wg.Add(1)
	go rb.monitor(cfg.CheckInterval)

	return rb
}

// Request attempts to allocate memory from the pool.
// Returns true if allocation succeeded, false if pool is full.
func (rb *ResourceBudget) Request(bytes int64) bool {
	for {
		current := atomic.LoadInt64(&rb.allocated)
		if current+bytes > rb.poolSize {
			return false
		}
		if atomic.CompareAndSwapInt64(&rb.allocated, current, current+bytes) {
			// Check if we should warn
			usage := float64(current+bytes) / float64(rb.poolSize)
			if usage >= rb.warnLevel && atomic.CompareAndSwapInt32(&rb.warnedRecently, 0, 1) {
				debug.Log("worker", "WARNING: memory pool at %.0f%% (%d/%d MB)",
					usage*100,
					(current+bytes)/(1024*1024),
					rb.poolSize/(1024*1024))
				// Reset warning flag after 1 minute
				go func() {
					time.Sleep(time.Minute)
					atomic.StoreInt32(&rb.warnedRecently, 0)
				}()
			}
			return true
		}
	}
}

// Release returns memory to the pool.
func (rb *ResourceBudget) Release(bytes int64) {
	for {
		current := atomic.LoadInt64(&rb.allocated)
		newVal := current - bytes
		if newVal < 0 {
			newVal = 0
		}
		if atomic.CompareAndSwapInt64(&rb.allocated, current, newVal) {
			return
		}
	}
}

// Allocated returns the current allocated bytes.
func (rb *ResourceBudget) Allocated() int64 {
	return atomic.LoadInt64(&rb.allocated)
}

// Available returns the available bytes in the pool.
func (rb *ResourceBudget) Available() int64 {
	return rb.poolSize - atomic.LoadInt64(&rb.allocated)
}

// PoolSize returns the total pool size.
func (rb *ResourceBudget) PoolSize() int64 {
	return rb.poolSize
}

// Usage returns the current usage as a fraction (0.0 - 1.0).
func (rb *ResourceBudget) Usage() float64 {
	return float64(atomic.LoadInt64(&rb.allocated)) / float64(rb.poolSize)
}

// Stats returns current budget statistics.
func (rb *ResourceBudget) Stats() BudgetStats {
	allocated := atomic.LoadInt64(&rb.allocated)
	return BudgetStats{
		PoolSize:  rb.poolSize,
		Allocated: allocated,
		Available: rb.poolSize - allocated,
		Usage:     float64(allocated) / float64(rb.poolSize),
	}
}

// BudgetStats holds budget statistics.
type BudgetStats struct {
	PoolSize  int64   `json:"pool_size"`
	Allocated int64   `json:"allocated"`
	Available int64   `json:"available"`
	Usage     float64 `json:"usage"`
}

// monitor runs in a goroutine and checks system MemStats.
// It logs warnings if actual heap usage diverges significantly from allocations.
func (rb *ResourceBudget) monitor(interval time.Duration) {
	defer rb.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rb.checkMemStats()
		case <-rb.done:
			return
		}
	}
}

// checkMemStats reads runtime memory stats and logs warnings if needed.
func (rb *ResourceBudget) checkMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	rb.lastHeapAlloc = m.HeapAlloc
	rb.lastCheck = time.Now()

	// Check if actual heap is much larger than our tracked allocations
	// This could indicate memory leaks or underestimated allocations
	tracked := atomic.LoadInt64(&rb.allocated)
	if tracked > 0 {
		ratio := float64(m.HeapAlloc) / float64(tracked)
		if ratio > 3.0 {
			// Heap is 3x our tracked value - log warning (once)
			if atomic.CompareAndSwapInt32(&rb.warnedRecently, 0, 1) {
				debug.Log("worker", "WARNING: heap (%d MB) >> tracked worker memory (%d MB)",
					m.HeapAlloc/(1024*1024), tracked/(1024*1024))
				go func() {
					time.Sleep(5 * time.Minute)
					atomic.StoreInt32(&rb.warnedRecently, 0)
				}()
			}
		}
	}
}

// Close stops the monitor goroutine.
func (rb *ResourceBudget) Close() {
	close(rb.done)
	rb.wg.Wait()
}

// GetLastMemStats returns the last recorded heap allocation.
func (rb *ResourceBudget) GetLastMemStats() (heapAlloc uint64, lastCheck time.Time) {
	return rb.lastHeapAlloc, rb.lastCheck
}
