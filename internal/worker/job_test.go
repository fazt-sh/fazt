package worker

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected *time.Duration
		wantErr  bool
	}{
		{"5m", durationPtr(5 * time.Minute), false},
		{"30s", durationPtr(30 * time.Second), false},
		{"1h", durationPtr(time.Hour), false},
		{"2h30m", durationPtr(2*time.Hour + 30*time.Minute), false},
		{"100ms", durationPtr(100 * time.Millisecond), false},
		{"30", durationPtr(30 * time.Second), false},
		{"5 m", durationPtr(5 * time.Minute), false},
		{"", nil, false},
		{"null", nil, false},
		{"0", nil, false},
		{"invalid", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.expected == nil && result != nil {
				t.Errorf("ParseDuration(%q) = %v, want nil", tt.input, *result)
				return
			}
			if tt.expected != nil && result == nil {
				t.Errorf("ParseDuration(%q) = nil, want %v", tt.input, *tt.expected)
				return
			}
			if tt.expected != nil && result != nil && *result != *tt.expected {
				t.Errorf("ParseDuration(%q) = %v, want %v", tt.input, *result, *tt.expected)
			}
		})
	}
}

func TestParseMemory(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"32MB", 32 * 1024 * 1024, false},
		{"64MB", 64 * 1024 * 1024, false},
		{"256MB", 256 * 1024 * 1024, false},
		{"1GB", 1024 * 1024 * 1024, false},
		{"1024KB", 1024 * 1024, false},
		{"1024", 1024, false},
		{"", 32 * 1024 * 1024, false}, // Default
		{"32 MB", 32 * 1024 * 1024, false},
		{"invalid", 0, true},
		{"MB", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseMemory(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMemory(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("ParseMemory(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestJobLifecycle(t *testing.T) {
	cfg := DefaultJobConfig()
	job := NewJob("test-1", "app-1", "workers/test.js", cfg)

	// Initial state
	if job.Status != StatusPending {
		t.Errorf("Initial status = %s, want %s", job.Status, StatusPending)
	}
	if job.Attempt != 1 {
		t.Errorf("Initial attempt = %d, want 1", job.Attempt)
	}

	// Mark running
	job.MarkRunning()
	if job.Status != StatusRunning {
		t.Errorf("After MarkRunning, status = %s, want %s", job.Status, StatusRunning)
	}
	if job.StartedAt.IsZero() {
		t.Error("StartedAt should be set after MarkRunning")
	}

	// Mark done
	job.MarkDone(map[string]string{"result": "success"})
	if job.Status != StatusDone {
		t.Errorf("After MarkDone, status = %s, want %s", job.Status, StatusDone)
	}
	if job.Progress != 1.0 {
		t.Errorf("After MarkDone, progress = %f, want 1.0", job.Progress)
	}
	if job.Result == "" {
		t.Error("Result should be set after MarkDone")
	}
}

func TestJobCancellation(t *testing.T) {
	cfg := DefaultJobConfig()
	job := NewJob("test-2", "app-1", "workers/test.js", cfg)

	// Set cancel function
	cancelled := false
	job.SetCancelFunc(func() {
		cancelled = true
	})

	// Initially not cancelled
	if job.IsCancelled() {
		t.Error("Job should not be cancelled initially")
	}

	// Cancel
	job.Cancel()
	if !job.IsCancelled() {
		t.Error("Job should be cancelled after Cancel()")
	}
	if !cancelled {
		t.Error("Cancel function should have been called")
	}
}

func TestJobProgress(t *testing.T) {
	cfg := DefaultJobConfig()
	job := NewJob("test-3", "app-1", "workers/test.js", cfg)

	// Set progress
	job.SetProgress(0.5)
	if job.GetProgress() != 0.5 {
		t.Errorf("Progress = %f, want 0.5", job.GetProgress())
	}

	// Clamp to 0
	job.SetProgress(-0.5)
	if job.GetProgress() != 0 {
		t.Errorf("Progress = %f, want 0 (clamped)", job.GetProgress())
	}

	// Clamp to 1
	job.SetProgress(1.5)
	if job.GetProgress() != 1.0 {
		t.Errorf("Progress = %f, want 1.0 (clamped)", job.GetProgress())
	}
}

func TestJobCheckpoint(t *testing.T) {
	cfg := DefaultJobConfig()
	job := NewJob("test-4", "app-1", "workers/test.js", cfg)

	// No checkpoint initially
	cp, err := job.GetCheckpoint()
	if err != nil {
		t.Fatalf("GetCheckpoint error: %v", err)
	}
	if cp != nil {
		t.Error("Checkpoint should be nil initially")
	}

	// Set checkpoint
	err = job.SetCheckpoint(map[string]interface{}{
		"cursor": 100,
		"state":  "processing",
	})
	if err != nil {
		t.Fatalf("SetCheckpoint error: %v", err)
	}

	// Get checkpoint
	cp, err = job.GetCheckpoint()
	if err != nil {
		t.Fatalf("GetCheckpoint error: %v", err)
	}
	if cp == nil {
		t.Fatal("Checkpoint should not be nil")
	}
	if cp["cursor"] != float64(100) { // JSON numbers are float64
		t.Errorf("Checkpoint cursor = %v, want 100", cp["cursor"])
	}
}

func TestJobLogs(t *testing.T) {
	cfg := DefaultJobConfig()
	job := NewJob("test-5", "app-1", "workers/test.js", cfg)

	// Add logs
	job.AddLog("Starting...")
	job.AddLog("Processing...")
	job.AddLog("Done")

	if len(job.Logs) != 3 {
		t.Errorf("Logs count = %d, want 3", len(job.Logs))
	}

	// Logs should have timestamp prefix
	for _, log := range job.Logs {
		if len(log) < 10 || log[0] != '[' {
			t.Errorf("Log entry should have timestamp prefix: %s", log)
		}
	}
}

func TestJobRetry(t *testing.T) {
	cfg := DefaultJobConfig()
	cfg.MaxAttempts = 3
	job := NewJob("test-6", "app-1", "workers/test.js", cfg)

	// Initial attempt=1, max=3, should retry
	if !job.ShouldRetry() {
		t.Error("Job should be retryable with attempts remaining")
	}

	// After first increment: attempt=2, max=3, should retry
	job.IncrementAttempt()
	if !job.ShouldRetry() {
		t.Error("Job should be retryable after 1 increment (attempt 2 of 3)")
	}

	// After second increment: attempt=3, max=3, should NOT retry
	job.IncrementAttempt()
	if job.ShouldRetry() {
		t.Error("Job should not be retryable at max attempts")
	}
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}
