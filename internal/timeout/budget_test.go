package timeout

import (
	"context"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.RequestTimeout != 10*time.Second {
		t.Errorf("expected RequestTimeout=10s, got %v", cfg.RequestTimeout)
	}
	if cfg.StorageOpTimeout != 3*time.Second {
		t.Errorf("expected StorageOpTimeout=3s, got %v", cfg.StorageOpTimeout)
	}
	if cfg.MinOperationTime != 500*time.Millisecond {
		t.Errorf("expected MinOperationTime=500ms, got %v", cfg.MinOperationTime)
	}
	if cfg.QueueAdmissionTime != 1*time.Second {
		t.Errorf("expected QueueAdmissionTime=1s, got %v", cfg.QueueAdmissionTime)
	}
}

func TestNewBudget_WithDeadline(t *testing.T) {
	cfg := DefaultConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	budget := NewBudget(ctx, cfg)

	// Should be approximately 5 seconds
	remaining := budget.Remaining()
	if remaining < 4*time.Second || remaining > 5*time.Second {
		t.Errorf("expected remaining ~5s, got %v", remaining)
	}
}

func TestNewBudget_WithoutDeadline(t *testing.T) {
	cfg := DefaultConfig()
	ctx := context.Background()

	budget := NewBudget(ctx, cfg)

	// Should use RequestTimeout (10s) as deadline
	remaining := budget.Remaining()
	if remaining < 9*time.Second || remaining > 10*time.Second {
		t.Errorf("expected remaining ~10s, got %v", remaining)
	}
}

func TestBudget_CanStartOperation(t *testing.T) {
	cfg := DefaultConfig()

	// Test with plenty of time
	ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()
	budget1 := NewBudget(ctx1, cfg)
	if !budget1.CanStartOperation() {
		t.Error("expected CanStartOperation=true with 5s remaining")
	}

	// Test with insufficient time (below MinOperationTime)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel2()
	budget2 := NewBudget(ctx2, cfg)
	if budget2.CanStartOperation() {
		t.Error("expected CanStartOperation=false with 100ms remaining")
	}
}

func TestBudget_CanEnterQueue(t *testing.T) {
	cfg := DefaultConfig()

	// Test with plenty of time
	ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()
	budget1 := NewBudget(ctx1, cfg)
	if !budget1.CanEnterQueue() {
		t.Error("expected CanEnterQueue=true with 5s remaining")
	}

	// Test with insufficient time (below QueueAdmissionTime)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel2()
	budget2 := NewBudget(ctx2, cfg)
	if budget2.CanEnterQueue() {
		t.Error("expected CanEnterQueue=false with 500ms remaining")
	}
}

func TestBudget_StorageContext(t *testing.T) {
	cfg := DefaultConfig()

	// Test with plenty of time
	ctx1, cancel1 := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel1()
	budget1 := NewBudget(ctx1, cfg)

	opCtx, opCancel, err := budget1.StorageContext(ctx1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer opCancel()

	// The timeout should be min(remaining/2, StorageOpTimeout) = min(4s, 3s) = 3s
	deadline, ok := opCtx.Deadline()
	if !ok {
		t.Fatal("expected context to have deadline")
	}
	timeout := time.Until(deadline)
	if timeout < 2900*time.Millisecond || timeout > 3100*time.Millisecond {
		t.Errorf("expected timeout ~3s, got %v", timeout)
	}
}

func TestBudget_StorageContext_UsesHalfRemaining(t *testing.T) {
	cfg := DefaultConfig()

	// With 2s remaining, should get 1s timeout (half of remaining)
	ctx1, cancel1 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel1()
	budget1 := NewBudget(ctx1, cfg)

	opCtx, opCancel, err := budget1.StorageContext(ctx1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer opCancel()

	deadline, ok := opCtx.Deadline()
	if !ok {
		t.Fatal("expected context to have deadline")
	}
	timeout := time.Until(deadline)
	// Should be ~1s (half of 2s remaining)
	if timeout < 900*time.Millisecond || timeout > 1100*time.Millisecond {
		t.Errorf("expected timeout ~1s, got %v", timeout)
	}
}

func TestBudget_StorageContext_InsufficientTime(t *testing.T) {
	cfg := DefaultConfig()

	// With 100ms remaining, should return error
	ctx1, cancel1 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel1()
	budget1 := NewBudget(ctx1, cfg)

	_, _, err := budget1.StorageContext(ctx1)
	if err != ErrInsufficientTime {
		t.Errorf("expected ErrInsufficientTime, got %v", err)
	}
}

func TestBudget_Config(t *testing.T) {
	cfg := DefaultConfig()
	ctx := context.Background()
	budget := NewBudget(ctx, cfg)

	returnedCfg := budget.Config()
	if returnedCfg != cfg {
		t.Error("Config() should return the same config")
	}
}
