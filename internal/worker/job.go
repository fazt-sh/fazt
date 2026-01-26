// Package worker provides background job execution with resource limits.
package worker

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// JobStatus represents the current state of a job.
type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusRunning   JobStatus = "running"
	StatusDone      JobStatus = "done"
	StatusFailed    JobStatus = "failed"
	StatusCancelled JobStatus = "cancelled"
)

// JobConfig holds job execution options.
type JobConfig struct {
	// Memory budget in bytes (default: 32MB)
	MemoryBytes int64 `json:"memory_bytes"`

	// Timeout duration (nil = indefinite for daemons)
	Timeout *time.Duration `json:"timeout,omitempty"`

	// Daemon mode - restart on crash
	Daemon bool `json:"daemon"`

	// Retry configuration
	MaxAttempts int           `json:"max_attempts"`
	RetryDelay  time.Duration `json:"retry_delay"`

	// Priority: -1 (low), 0 (normal), 1 (high)
	Priority int `json:"priority"`

	// UniqueKey prevents duplicate jobs
	UniqueKey string `json:"unique_key,omitempty"`

	// Data passed to the handler
	Data map[string]interface{} `json:"data,omitempty"`

	// Idle timeout - stop if no listeners on IdleChannel for this duration
	IdleTimeout *time.Duration `json:"idle_timeout,omitempty"`
	IdleChannel string         `json:"idle_channel,omitempty"`
}

// DefaultJobConfig returns sensible defaults.
func DefaultJobConfig() JobConfig {
	timeout := 30 * time.Minute
	return JobConfig{
		MemoryBytes: 32 * 1024 * 1024, // 32MB
		Timeout:     &timeout,
		Daemon:      false,
		MaxAttempts: 1,
		RetryDelay:  time.Minute,
		Priority:    0,
	}
}

// Job represents a background job.
type Job struct {
	ID        string    `json:"id"`
	AppID     string    `json:"app_id"`
	Handler   string    `json:"handler"` // e.g., "workers/sync.js"
	Status    JobStatus `json:"status"`
	Config    JobConfig `json:"config"`
	Progress  float64   `json:"progress"` // 0.0 - 1.0
	Result    string    `json:"result,omitempty"`
	Error     string    `json:"error,omitempty"`
	Logs      []string  `json:"logs,omitempty"`
	Attempt   int       `json:"attempt"`
	CreatedAt time.Time `json:"created_at"`
	StartedAt time.Time `json:"started_at,omitempty"`
	DoneAt    time.Time `json:"done_at,omitempty"`

	// Checkpoint for crash recovery (daemon mode)
	Checkpoint string `json:"checkpoint,omitempty"`

	// Daemon restart tracking
	RestartCount   int           `json:"restart_count"`
	DaemonBackoff  time.Duration `json:"daemon_backoff"`
	LastHealthyAt  time.Time     `json:"last_healthy_at,omitempty"`

	// Runtime state (not persisted)
	mu        sync.RWMutex
	cancelled bool
	cancelFn  func()
}

// NewJob creates a new job with the given configuration.
func NewJob(id, appID, handler string, cfg JobConfig) *Job {
	return &Job{
		ID:        id,
		AppID:     appID,
		Handler:   handler,
		Status:    StatusPending,
		Config:    cfg,
		Attempt:   1,
		CreatedAt: time.Now(),
		Logs:      make([]string, 0),
	}
}

// Cancel marks the job as cancelled.
func (j *Job) Cancel() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.cancelled = true
	if j.cancelFn != nil {
		j.cancelFn()
	}
}

// IsCancelled returns true if the job was cancelled.
func (j *Job) IsCancelled() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.cancelled
}

// SetCancelFunc sets the function to call when job is cancelled.
func (j *Job) SetCancelFunc(fn func()) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.cancelFn = fn
}

// SetProgress updates the job progress (0.0 - 1.0).
func (j *Job) SetProgress(p float64) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if p < 0 {
		p = 0
	}
	if p > 1.0 {
		p = 1.0
	}
	j.Progress = p
}

// GetProgress returns the current progress.
func (j *Job) GetProgress() float64 {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.Progress
}

// AddLog appends a log entry.
func (j *Job) AddLog(msg string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Logs = append(j.Logs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg))
	// Keep only last 100 logs
	if len(j.Logs) > 100 {
		j.Logs = j.Logs[len(j.Logs)-100:]
	}
}

// SetCheckpoint saves checkpoint data for recovery.
func (j *Job) SetCheckpoint(data interface{}) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// Limit checkpoint size to 1MB
	if len(bytes) > 1024*1024 {
		return fmt.Errorf("checkpoint too large: %d bytes (max 1MB)", len(bytes))
	}
	j.Checkpoint = string(bytes)
	return nil
}

// GetCheckpoint retrieves checkpoint data.
func (j *Job) GetCheckpoint() (map[string]interface{}, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()
	if j.Checkpoint == "" {
		return nil, nil
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(j.Checkpoint), &data); err != nil {
		return nil, err
	}
	return data, nil
}

// MarkRunning transitions job to running state.
func (j *Job) MarkRunning() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = StatusRunning
	j.StartedAt = time.Now()
}

// MarkDone transitions job to done state with result.
func (j *Job) MarkDone(result interface{}) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = StatusDone
	j.DoneAt = time.Now()
	j.Progress = 1.0
	if result != nil {
		bytes, err := json.Marshal(result)
		if err != nil {
			return err
		}
		j.Result = string(bytes)
	}
	return nil
}

// MarkFailed transitions job to failed state with error.
func (j *Job) MarkFailed(err error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = StatusFailed
	j.DoneAt = time.Now()
	if err != nil {
		j.Error = err.Error()
	}
}

// MarkCancelled transitions job to cancelled state.
func (j *Job) MarkCancelled() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = StatusCancelled
	j.DoneAt = time.Now()
	j.cancelled = true
}

// ShouldRetry returns true if the job should be retried after failure.
func (j *Job) ShouldRetry() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.Attempt < j.Config.MaxAttempts
}

// IncrementAttempt increments the attempt counter.
func (j *Job) IncrementAttempt() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Attempt++
}

// ParseDuration parses a duration string like "5m", "30s", "1h".
// Returns nil for null/indefinite.
func ParseDuration(s string) (*time.Duration, error) {
	if s == "" || s == "null" || s == "0" {
		return nil, nil
	}

	// Try standard Go duration format first
	d, err := time.ParseDuration(s)
	if err == nil {
		return &d, nil
	}

	// Try custom format with just number and unit
	re := regexp.MustCompile(`^(\d+)\s*(s|sec|m|min|h|hr|d|day)?$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(s))
	if matches == nil {
		return nil, fmt.Errorf("invalid duration: %s", s)
	}

	num, _ := strconv.ParseInt(matches[1], 10, 64)
	unit := matches[2]

	var multiplier time.Duration
	switch unit {
	case "s", "sec", "":
		multiplier = time.Second
	case "m", "min":
		multiplier = time.Minute
	case "h", "hr":
		multiplier = time.Hour
	case "d", "day":
		multiplier = 24 * time.Hour
	default:
		multiplier = time.Second
	}

	d = time.Duration(num) * multiplier
	return &d, nil
}

// ParseMemory parses a memory string like "32MB", "64MB", "256MB".
func ParseMemory(s string) (int64, error) {
	if s == "" {
		return 32 * 1024 * 1024, nil // Default 32MB
	}

	s = strings.TrimSpace(strings.ToUpper(s))
	re := regexp.MustCompile(`^(\d+)\s*(B|KB|MB|GB)?$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("invalid memory: %s", s)
	}

	num, _ := strconv.ParseInt(matches[1], 10, 64)
	unit := matches[2]

	var multiplier int64
	switch unit {
	case "B", "":
		multiplier = 1
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	}

	return num * multiplier, nil
}

// JobRow holds nullable database values for constructing a Job.
type JobRow struct {
	ID              string
	AppID           string
	Handler         string
	Status          string
	ConfigJSON      string
	Progress        float64
	Result          *string
	Error           *string
	LogsJSON        *string
	Checkpoint      *string
	Attempt         int
	RestartCount    int
	DaemonBackoffMs int64
	CreatedAt       *int64
	StartedAt       *int64
	DoneAt          *int64
	LastHealthyAt   *int64
}

// JobFromRow constructs a Job from a JobRow.
func JobFromRow(row JobRow) (*Job, error) {
	var cfg JobConfig
	if row.ConfigJSON != "" {
		if err := json.Unmarshal([]byte(row.ConfigJSON), &cfg); err != nil {
			cfg = DefaultJobConfig()
		}
	} else {
		cfg = DefaultJobConfig()
	}

	var logs []string
	if row.LogsJSON != nil && *row.LogsJSON != "" {
		json.Unmarshal([]byte(*row.LogsJSON), &logs)
	}
	if logs == nil {
		logs = make([]string, 0)
	}

	j := &Job{
		ID:            row.ID,
		AppID:         row.AppID,
		Handler:       row.Handler,
		Status:        JobStatus(row.Status),
		Config:        cfg,
		Progress:      row.Progress,
		Logs:          logs,
		Attempt:       row.Attempt,
		RestartCount:  row.RestartCount,
		DaemonBackoff: time.Duration(row.DaemonBackoffMs) * time.Millisecond,
	}

	if row.Result != nil {
		j.Result = *row.Result
	}
	if row.Error != nil {
		j.Error = *row.Error
	}
	if row.Checkpoint != nil {
		j.Checkpoint = *row.Checkpoint
	}
	if row.CreatedAt != nil {
		j.CreatedAt = time.UnixMilli(*row.CreatedAt)
	}
	if row.StartedAt != nil {
		j.StartedAt = time.UnixMilli(*row.StartedAt)
	}
	if row.DoneAt != nil {
		j.DoneAt = time.UnixMilli(*row.DoneAt)
	}
	if row.LastHealthyAt != nil {
		j.LastHealthyAt = time.UnixMilli(*row.LastHealthyAt)
	}

	return j, nil
}
