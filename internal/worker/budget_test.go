package worker

import (
	"testing"
	"time"
)

func TestResourceBudgetBasic(t *testing.T) {
	cfg := BudgetConfig{
		PoolSize:      100 * 1024 * 1024, // 100MB
		WarnLevel:     0.8,
		CheckInterval: time.Second, // Slow interval for testing
	}
	budget := NewResourceBudget(cfg)
	defer budget.Close()

	// Initial state
	if budget.Allocated() != 0 {
		t.Errorf("Initial allocated = %d, want 0", budget.Allocated())
	}
	if budget.Available() != 100*1024*1024 {
		t.Errorf("Initial available = %d, want %d", budget.Available(), 100*1024*1024)
	}
	if budget.PoolSize() != 100*1024*1024 {
		t.Errorf("Pool size = %d, want %d", budget.PoolSize(), 100*1024*1024)
	}
}

func TestResourceBudgetAllocation(t *testing.T) {
	cfg := BudgetConfig{
		PoolSize:      100 * 1024 * 1024,
		WarnLevel:     0.8,
		CheckInterval: time.Second,
	}
	budget := NewResourceBudget(cfg)
	defer budget.Close()

	// Request 50MB
	if !budget.Request(50 * 1024 * 1024) {
		t.Error("Should be able to request 50MB")
	}
	if budget.Allocated() != 50*1024*1024 {
		t.Errorf("Allocated = %d, want %d", budget.Allocated(), 50*1024*1024)
	}

	// Request another 40MB
	if !budget.Request(40 * 1024 * 1024) {
		t.Error("Should be able to request another 40MB")
	}

	// Request 20MB should fail (only 10MB left)
	if budget.Request(20 * 1024 * 1024) {
		t.Error("Should not be able to request 20MB when only 10MB available")
	}

	// Release 40MB
	budget.Release(40 * 1024 * 1024)
	if budget.Allocated() != 50*1024*1024 {
		t.Errorf("After release, allocated = %d, want %d", budget.Allocated(), 50*1024*1024)
	}

	// Now 20MB request should succeed
	if !budget.Request(20 * 1024 * 1024) {
		t.Error("Should be able to request 20MB after release")
	}
}

func TestResourceBudgetUsage(t *testing.T) {
	cfg := BudgetConfig{
		PoolSize:      100 * 1024 * 1024,
		WarnLevel:     0.8,
		CheckInterval: time.Second,
	}
	budget := NewResourceBudget(cfg)
	defer budget.Close()

	// Initial usage
	if budget.Usage() != 0.0 {
		t.Errorf("Initial usage = %f, want 0.0", budget.Usage())
	}

	// Request 50MB - 50% usage
	budget.Request(50 * 1024 * 1024)
	if budget.Usage() != 0.5 {
		t.Errorf("After 50MB, usage = %f, want 0.5", budget.Usage())
	}

	// Request 30MB more - 80% usage
	budget.Request(30 * 1024 * 1024)
	if budget.Usage() != 0.8 {
		t.Errorf("After 80MB, usage = %f, want 0.8", budget.Usage())
	}
}

func TestResourceBudgetStats(t *testing.T) {
	cfg := BudgetConfig{
		PoolSize:      100 * 1024 * 1024,
		WarnLevel:     0.8,
		CheckInterval: time.Second,
	}
	budget := NewResourceBudget(cfg)
	defer budget.Close()

	budget.Request(30 * 1024 * 1024)

	stats := budget.Stats()
	if stats.PoolSize != 100*1024*1024 {
		t.Errorf("Stats.PoolSize = %d, want %d", stats.PoolSize, 100*1024*1024)
	}
	if stats.Allocated != 30*1024*1024 {
		t.Errorf("Stats.Allocated = %d, want %d", stats.Allocated, 30*1024*1024)
	}
	if stats.Available != 70*1024*1024 {
		t.Errorf("Stats.Available = %d, want %d", stats.Available, 70*1024*1024)
	}
	if stats.Usage != 0.3 {
		t.Errorf("Stats.Usage = %f, want 0.3", stats.Usage)
	}
}

func TestResourceBudgetConcurrent(t *testing.T) {
	cfg := BudgetConfig{
		PoolSize:      100 * 1024 * 1024,
		WarnLevel:     0.9,
		CheckInterval: time.Second,
	}
	budget := NewResourceBudget(cfg)
	defer budget.Close()

	// Run concurrent allocations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				if budget.Request(1 * 1024 * 1024) {
					// Use briefly
					budget.Release(1 * 1024 * 1024)
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// All memory should be released
	if budget.Allocated() != 0 {
		t.Errorf("After concurrent test, allocated = %d, want 0", budget.Allocated())
	}
}

func TestResourceBudgetReleaseUnderflow(t *testing.T) {
	cfg := BudgetConfig{
		PoolSize:      100 * 1024 * 1024,
		WarnLevel:     0.8,
		CheckInterval: time.Second,
	}
	budget := NewResourceBudget(cfg)
	defer budget.Close()

	// Release without allocation - should not go negative
	budget.Release(50 * 1024 * 1024)
	if budget.Allocated() != 0 {
		t.Errorf("After underflow release, allocated = %d, want 0", budget.Allocated())
	}
}

func TestResourceBudgetClose(t *testing.T) {
	cfg := BudgetConfig{
		PoolSize:      100 * 1024 * 1024,
		WarnLevel:     0.8,
		CheckInterval: 10 * time.Millisecond, // Fast for testing
	}
	budget := NewResourceBudget(cfg)

	// Let monitor run a few cycles
	time.Sleep(50 * time.Millisecond)

	// Close should not hang
	done := make(chan bool)
	go func() {
		budget.Close()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("Close should complete quickly")
	}
}
