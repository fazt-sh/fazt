package worker

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/fazt-sh/fazt/internal/debug"
)

// Default limits
const (
	DefaultMaxConcurrentTotal  = 20
	DefaultMaxConcurrentPerApp = 5
	DefaultMaxQueueDepth       = 100
	DefaultMemoryPoolBytes     = 256 * 1024 * 1024 // 256MB
	DefaultMemoryPerJobBytes   = 32 * 1024 * 1024  // 32MB
	DefaultTimeoutMinutes      = 30
	DefaultMaxDaemonsPerApp    = 2
)

// PoolConfig configures the worker pool.
type PoolConfig struct {
	MaxConcurrentTotal  int   // All apps combined
	MaxConcurrentPerApp int   // Per app
	MaxQueueDepth       int   // Queued jobs per app
	MemoryPoolBytes     int64 // Total memory for all workers
	MaxDaemonsPerApp    int   // Max daemon workers per app
}

// DefaultPoolConfig returns sensible defaults.
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxConcurrentTotal:  DefaultMaxConcurrentTotal,
		MaxConcurrentPerApp: DefaultMaxConcurrentPerApp,
		MaxQueueDepth:       DefaultMaxQueueDepth,
		MemoryPoolBytes:     DefaultMemoryPoolBytes,
		MaxDaemonsPerApp:    DefaultMaxDaemonsPerApp,
	}
}

// Pool manages background job execution.
type Pool struct {
	config PoolConfig
	db     *sql.DB

	// Active jobs by ID
	jobs   map[string]*Job
	jobsMu sync.RWMutex

	// Queue for pending jobs
	queue chan *Job

	// Per-app tracking
	appJobs   map[string]int // count of running jobs per app
	appJobsMu sync.RWMutex

	// Resource budget
	allocatedMemory int64
	memoryMu        sync.RWMutex

	// Lifecycle
	done   chan struct{}
	wg     sync.WaitGroup
	closed bool
	mu     sync.Mutex

	// Executor function (set externally, executes JS code)
	executor JobExecutor

	// Listener count function (for idle timeout checking)
	listenerCountFn ListenerCountFunc
}

// JobExecutor executes a job and returns the result.
type JobExecutor func(ctx context.Context, job *Job, code string) (interface{}, error)

// ListenerCountFunc returns the number of WebSocket listeners for a channel.
type ListenerCountFunc func(appID, channel string) int

// NewPool creates a new worker pool.
func NewPool(db *sql.DB, cfg PoolConfig) *Pool {
	if cfg.MaxConcurrentTotal <= 0 {
		cfg.MaxConcurrentTotal = DefaultMaxConcurrentTotal
	}
	if cfg.MaxConcurrentPerApp <= 0 {
		cfg.MaxConcurrentPerApp = DefaultMaxConcurrentPerApp
	}
	if cfg.MaxQueueDepth <= 0 {
		cfg.MaxQueueDepth = DefaultMaxQueueDepth
	}
	if cfg.MemoryPoolBytes <= 0 {
		cfg.MemoryPoolBytes = DefaultMemoryPoolBytes
	}
	if cfg.MaxDaemonsPerApp <= 0 {
		cfg.MaxDaemonsPerApp = DefaultMaxDaemonsPerApp
	}

	p := &Pool{
		config:  cfg,
		db:      db,
		jobs:    make(map[string]*Job),
		appJobs: make(map[string]int),
		queue:   make(chan *Job, cfg.MaxQueueDepth*10), // Allow some queue depth
		done:    make(chan struct{}),
	}

	// Start worker goroutines
	for i := 0; i < cfg.MaxConcurrentTotal; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	debug.Log("worker", "pool started: %d workers, %dMB memory pool",
		cfg.MaxConcurrentTotal, cfg.MemoryPoolBytes/(1024*1024))

	return p
}

// SetExecutor sets the job executor function.
func (p *Pool) SetExecutor(exec JobExecutor) {
	p.executor = exec
}

// SetListenerCountFunc sets the function to check WebSocket listener count.
func (p *Pool) SetListenerCountFunc(fn ListenerCountFunc) {
	p.listenerCountFn = fn
}

// worker is a goroutine that processes jobs from the queue.
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case job := <-p.queue:
			if job == nil {
				continue
			}
			p.executeJob(job)

		case <-p.done:
			return
		}
	}
}

// Spawn creates and queues a new job.
func (p *Pool) Spawn(appID, handler string, cfg JobConfig) (*Job, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, fmt.Errorf("pool is closed")
	}
	p.mu.Unlock()

	// Check unique key
	if cfg.UniqueKey != "" {
		if existing := p.findByUniqueKey(appID, cfg.UniqueKey); existing != nil {
			return existing, nil
		}
	}

	// Check per-app limits
	p.appJobsMu.RLock()
	appCount := p.appJobs[appID]
	p.appJobsMu.RUnlock()

	if appCount >= p.config.MaxConcurrentPerApp {
		// Check queue depth
		queuedCount := p.queuedCountForApp(appID)
		if queuedCount >= p.config.MaxQueueDepth {
			return nil, fmt.Errorf("queue full for app %s", appID)
		}
	}

	// Check daemon limit
	if cfg.Daemon {
		daemonCount := p.daemonCountForApp(appID)
		if daemonCount >= p.config.MaxDaemonsPerApp {
			return nil, fmt.Errorf("max daemons (%d) reached for app %s",
				p.config.MaxDaemonsPerApp, appID)
		}
	}

	// Check memory budget
	if cfg.MemoryBytes <= 0 {
		cfg.MemoryBytes = DefaultMemoryPerJobBytes
	}

	// Generate job ID
	id := generateJobID()

	job := NewJob(id, appID, handler, cfg)

	// Persist to database
	if err := p.persistJob(job); err != nil {
		return nil, fmt.Errorf("failed to persist job: %w", err)
	}

	// Track the job
	p.jobsMu.Lock()
	p.jobs[job.ID] = job
	p.jobsMu.Unlock()

	// Queue for execution
	select {
	case p.queue <- job:
		debug.Log("worker", "job %s queued: handler=%s app=%s", job.ID, handler, appID)
	default:
		// Queue is full, but we already checked, so this shouldn't happen
		return nil, fmt.Errorf("failed to queue job")
	}

	return job, nil
}

// executeJob runs a single job.
func (p *Pool) executeJob(job *Job) {
	// Try to allocate memory
	if !p.allocateMemory(job.Config.MemoryBytes) {
		// Requeue if pool full (wait for memory)
		time.Sleep(100 * time.Millisecond)
		select {
		case p.queue <- job:
		case <-p.done:
			return
		}
		return
	}
	defer p.releaseMemory(job.Config.MemoryBytes)

	// Increment app job count
	p.appJobsMu.Lock()
	p.appJobs[job.AppID]++
	p.appJobsMu.Unlock()
	defer func() {
		p.appJobsMu.Lock()
		p.appJobs[job.AppID]--
		if p.appJobs[job.AppID] <= 0 {
			delete(p.appJobs, job.AppID)
		}
		p.appJobsMu.Unlock()
	}()

	// Mark as running
	job.MarkRunning()
	p.updateJobStatus(job)

	// Set up context with timeout
	var ctx context.Context
	var cancel context.CancelFunc

	if job.Config.Timeout != nil && *job.Config.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), *job.Config.Timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	// Allow job cancellation
	job.SetCancelFunc(cancel)

	// Start idle watcher if configured
	var idleReason string
	if job.Config.IdleTimeout != nil && job.Config.IdleChannel != "" && p.listenerCountFn != nil {
		go p.watchIdleTimeout(ctx, cancel, job, &idleReason)
	}

	// Execute the job
	debug.Log("worker", "job %s started: handler=%s", job.ID, job.Handler)

	if p.executor == nil {
		job.MarkFailed(fmt.Errorf("no executor configured"))
		p.updateJobStatus(job)
		return
	}

	// Load the handler code from VFS
	code, err := p.loadHandlerCode(job.AppID, job.Handler)
	if err != nil {
		job.MarkFailed(fmt.Errorf("failed to load handler: %w", err))
		p.updateJobStatus(job)
		p.handleJobComplete(job)
		return
	}

	// Execute
	result, err := p.executor(ctx, job, code)

	// Handle result
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			job.AddLog("Job timed out")
			job.MarkFailed(fmt.Errorf("timeout exceeded"))
		} else if idleReason != "" {
			// Stopped due to idle timeout - this is a clean stop, not failure
			job.AddLog(idleReason)
			job.MarkDone(map[string]interface{}{"reason": "idle_timeout"})
			// Disable daemon restart for idle stop
			job.Config.Daemon = false
		} else if ctx.Err() == context.Canceled || job.IsCancelled() {
			job.AddLog("Job cancelled")
			job.MarkCancelled()
		} else {
			job.AddLog(fmt.Sprintf("Error: %v", err))
			job.MarkFailed(err)
		}
	} else {
		if err := job.MarkDone(result); err != nil {
			job.MarkFailed(fmt.Errorf("failed to serialize result: %w", err))
		}
	}

	p.updateJobStatus(job)
	p.handleJobComplete(job)

	debug.Log("worker", "job %s completed: status=%s duration=%v",
		job.ID, job.Status, time.Since(job.StartedAt))
}

// handleJobComplete handles post-completion logic (retry, daemon restart).
func (p *Pool) handleJobComplete(job *Job) {
	// Remove from active jobs if not a daemon that needs restart
	shouldRemove := true

	if job.Status == StatusFailed {
		// Check retry
		if job.ShouldRetry() {
			job.IncrementAttempt()
			job.Status = StatusPending
			p.updateJobStatus(job)

			// Delay before retry
			time.AfterFunc(job.Config.RetryDelay, func() {
				select {
				case p.queue <- job:
				case <-p.done:
				}
			})
			shouldRemove = false
		} else if job.Config.Daemon {
			// Daemon restart with backoff
			p.scheduleDaemonRestart(job)
			shouldRemove = false
		}
	}

	if shouldRemove {
		p.jobsMu.Lock()
		delete(p.jobs, job.ID)
		p.jobsMu.Unlock()
	}
}

// scheduleDaemonRestart schedules a daemon job for restart with backoff.
func (p *Pool) scheduleDaemonRestart(job *Job) {
	// Calculate backoff: 1s, 2s, 4s, 8s, ... max 60s
	backoff := time.Second * time.Duration(1<<job.RestartCount)
	if backoff > 60*time.Second {
		backoff = 60 * time.Second
	}

	// Reset backoff if healthy for 5 minutes
	if !job.LastHealthyAt.IsZero() && time.Since(job.LastHealthyAt) > 5*time.Minute {
		backoff = time.Second
		job.RestartCount = 0
	}

	job.RestartCount++
	job.DaemonBackoff = backoff
	job.Status = StatusPending
	job.cancelled = false
	p.updateJobStatus(job)

	debug.Log("worker", "daemon %s will restart in %v (attempt %d)",
		job.ID, backoff, job.RestartCount)

	time.AfterFunc(backoff, func() {
		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			return
		}
		p.mu.Unlock()

		select {
		case p.queue <- job:
		case <-p.done:
		}
	})
}

// Cancel cancels a job by ID.
func (p *Pool) Cancel(jobID string) error {
	p.jobsMu.RLock()
	job, ok := p.jobs[jobID]
	p.jobsMu.RUnlock()

	if !ok {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Cancel()

	// For daemons, mark as cancelled to prevent restart
	if job.Config.Daemon {
		job.Config.Daemon = false // Prevent restart
	}

	return nil
}

// Get returns a job by ID.
func (p *Pool) Get(jobID string) (*Job, error) {
	p.jobsMu.RLock()
	job, ok := p.jobs[jobID]
	p.jobsMu.RUnlock()

	if ok {
		return job, nil
	}

	// Try to load from database
	return p.loadJob(jobID)
}

// List returns jobs matching the filter.
func (p *Pool) List(appID string, status *JobStatus, limit int) ([]*Job, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, app_id, handler, status, config, progress,
		       result, error, logs, checkpoint, attempt, restart_count,
		       daemon_backoff_ms, created_at, started_at, done_at, last_healthy_at
		FROM worker_jobs
		WHERE 1=1
	`
	args := make([]interface{}, 0)

	if appID != "" {
		query += " AND app_id = ?"
		args = append(args, appID)
	}
	if status != nil {
		query += " AND status = ?"
		args = append(args, string(*status))
	}

	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]*Job, 0)
	for rows.Next() {
		var row JobRow
		var result, errorStr, logsJSON, checkpoint sql.NullString
		var createdAt, startedAt, doneAt, lastHealthyAt sql.NullInt64

		err := rows.Scan(
			&row.ID, &row.AppID, &row.Handler, &row.Status, &row.ConfigJSON,
			&row.Progress, &result, &errorStr, &logsJSON, &checkpoint,
			&row.Attempt, &row.RestartCount, &row.DaemonBackoffMs,
			&createdAt, &startedAt, &doneAt, &lastHealthyAt,
		)
		if err != nil {
			continue
		}

		if result.Valid {
			row.Result = &result.String
		}
		if errorStr.Valid {
			row.Error = &errorStr.String
		}
		if logsJSON.Valid {
			row.LogsJSON = &logsJSON.String
		}
		if checkpoint.Valid {
			row.Checkpoint = &checkpoint.String
		}
		if createdAt.Valid {
			row.CreatedAt = &createdAt.Int64
		}
		if startedAt.Valid {
			row.StartedAt = &startedAt.Int64
		}
		if doneAt.Valid {
			row.DoneAt = &doneAt.Int64
		}
		if lastHealthyAt.Valid {
			row.LastHealthyAt = &lastHealthyAt.Int64
		}

		job, err := JobFromRow(row)
		if err != nil {
			continue
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// Shutdown gracefully shuts down the pool.
func (p *Pool) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.mu.Unlock()

	debug.Log("worker", "shutting down pool...")

	// Signal workers to stop
	close(p.done)

	// Cancel all running jobs
	p.jobsMu.RLock()
	for _, job := range p.jobs {
		if job.Status == StatusRunning {
			job.Cancel()
		}
	}
	p.jobsMu.RUnlock()

	// Wait for workers with timeout
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		debug.Log("worker", "pool shutdown complete")
		return nil
	case <-ctx.Done():
		debug.Log("worker", "pool shutdown timed out")
		return ctx.Err()
	}
}

// RestoreDaemons restores daemon jobs that were running when server stopped.
func (p *Pool) RestoreDaemons() error {
	rows, err := p.db.Query(`
		SELECT id, app_id, handler, status, config, progress,
		       result, error, logs, checkpoint, attempt, restart_count,
		       daemon_backoff_ms, created_at, started_at, done_at, last_healthy_at
		FROM worker_jobs
		WHERE json_extract(config, '$.daemon') = 1
		  AND status IN ('running', 'pending')
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	restored := 0
	for rows.Next() {
		var row JobRow
		var result, errorStr, logsJSON, checkpoint sql.NullString
		var createdAt, startedAt, doneAt, lastHealthyAt sql.NullInt64

		err := rows.Scan(
			&row.ID, &row.AppID, &row.Handler, &row.Status, &row.ConfigJSON,
			&row.Progress, &result, &errorStr, &logsJSON, &checkpoint,
			&row.Attempt, &row.RestartCount, &row.DaemonBackoffMs,
			&createdAt, &startedAt, &doneAt, &lastHealthyAt,
		)
		if err != nil {
			continue
		}

		if result.Valid {
			row.Result = &result.String
		}
		if errorStr.Valid {
			row.Error = &errorStr.String
		}
		if logsJSON.Valid {
			row.LogsJSON = &logsJSON.String
		}
		if checkpoint.Valid {
			row.Checkpoint = &checkpoint.String
		}
		if createdAt.Valid {
			row.CreatedAt = &createdAt.Int64
		}
		if startedAt.Valid {
			row.StartedAt = &startedAt.Int64
		}
		if doneAt.Valid {
			row.DoneAt = &doneAt.Int64
		}
		if lastHealthyAt.Valid {
			row.LastHealthyAt = &lastHealthyAt.Int64
		}

		job, err := JobFromRow(row)
		if err != nil {
			continue
		}

		// Add to active jobs and queue
		p.jobsMu.Lock()
		p.jobs[job.ID] = job
		p.jobsMu.Unlock()

		select {
		case p.queue <- job:
			restored++
			debug.Log("worker", "restored daemon %s: handler=%s", job.ID, job.Handler)
		default:
		}
	}

	if restored > 0 {
		debug.Log("worker", "restored %d daemon jobs", restored)
	}

	return nil
}

// Stats returns current pool statistics.
func (p *Pool) Stats() PoolStats {
	p.jobsMu.RLock()
	activeCount := 0
	queuedCount := 0
	for _, job := range p.jobs {
		if job.Status == StatusRunning {
			activeCount++
		} else if job.Status == StatusPending {
			queuedCount++
		}
	}
	p.jobsMu.RUnlock()

	p.memoryMu.RLock()
	allocated := p.allocatedMemory
	p.memoryMu.RUnlock()

	return PoolStats{
		ActiveJobs:      activeCount,
		QueuedJobs:      queuedCount,
		TotalJobs:       len(p.jobs),
		AllocatedMemory: allocated,
		PoolMemory:      p.config.MemoryPoolBytes,
		MemoryUsedPct:   float64(allocated) / float64(p.config.MemoryPoolBytes),
	}
}

// PoolStats holds pool statistics.
type PoolStats struct {
	ActiveJobs      int     `json:"active_jobs"`
	QueuedJobs      int     `json:"queued_jobs"`
	TotalJobs       int     `json:"total_jobs"`
	AllocatedMemory int64   `json:"allocated_memory"`
	PoolMemory      int64   `json:"pool_memory"`
	MemoryUsedPct   float64 `json:"memory_used_pct"`
}

// Helper functions

func (p *Pool) allocateMemory(bytes int64) bool {
	p.memoryMu.Lock()
	defer p.memoryMu.Unlock()

	if p.allocatedMemory+bytes > p.config.MemoryPoolBytes {
		return false
	}
	p.allocatedMemory += bytes
	return true
}

func (p *Pool) releaseMemory(bytes int64) {
	p.memoryMu.Lock()
	defer p.memoryMu.Unlock()
	p.allocatedMemory -= bytes
	if p.allocatedMemory < 0 {
		p.allocatedMemory = 0
	}
}

func (p *Pool) findByUniqueKey(appID, key string) *Job {
	p.jobsMu.RLock()
	defer p.jobsMu.RUnlock()

	for _, job := range p.jobs {
		if job.AppID == appID && job.Config.UniqueKey == key {
			if job.Status == StatusPending || job.Status == StatusRunning {
				return job
			}
		}
	}
	return nil
}

func (p *Pool) queuedCountForApp(appID string) int {
	p.jobsMu.RLock()
	defer p.jobsMu.RUnlock()

	count := 0
	for _, job := range p.jobs {
		if job.AppID == appID && job.Status == StatusPending {
			count++
		}
	}
	return count
}

func (p *Pool) daemonCountForApp(appID string) int {
	p.jobsMu.RLock()
	defer p.jobsMu.RUnlock()

	count := 0
	for _, job := range p.jobs {
		if job.AppID == appID && job.Config.Daemon {
			if job.Status == StatusPending || job.Status == StatusRunning {
				count++
			}
		}
	}
	return count
}

func (p *Pool) persistJob(job *Job) error {
	configJSON, _ := json.Marshal(job.Config)
	logsJSON, _ := json.Marshal(job.Logs)

	_, err := p.db.Exec(`
		INSERT INTO worker_jobs (
			id, app_id, handler, status, config, progress,
			result, error, logs, checkpoint, attempt, restart_count,
			daemon_backoff_ms, created_at, started_at, done_at, last_healthy_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		job.ID, job.AppID, job.Handler, string(job.Status), string(configJSON),
		job.Progress, nullString(job.Result), nullString(job.Error),
		string(logsJSON), nullString(job.Checkpoint), job.Attempt,
		job.RestartCount, int64(job.DaemonBackoff/time.Millisecond),
		nullTime(job.CreatedAt), nullTime(job.StartedAt),
		nullTime(job.DoneAt), nullTime(job.LastHealthyAt),
	)
	return err
}

func (p *Pool) updateJobStatus(job *Job) {
	configJSON, _ := json.Marshal(job.Config)
	logsJSON, _ := json.Marshal(job.Logs)

	p.db.Exec(`
		UPDATE worker_jobs SET
			status = ?, config = ?, progress = ?, result = ?, error = ?,
			logs = ?, checkpoint = ?, attempt = ?, restart_count = ?,
			daemon_backoff_ms = ?, started_at = ?, done_at = ?, last_healthy_at = ?
		WHERE id = ?
	`,
		string(job.Status), string(configJSON), job.Progress,
		nullString(job.Result), nullString(job.Error),
		string(logsJSON), nullString(job.Checkpoint), job.Attempt,
		job.RestartCount, int64(job.DaemonBackoff/time.Millisecond),
		nullTime(job.StartedAt), nullTime(job.DoneAt), nullTime(job.LastHealthyAt),
		job.ID,
	)
}

func (p *Pool) loadJob(id string) (*Job, error) {
	var row JobRow
	var result, errorStr, logsJSON, checkpoint sql.NullString
	var createdAt, startedAt, doneAt, lastHealthyAt sql.NullInt64

	err := p.db.QueryRow(`
		SELECT id, app_id, handler, status, config, progress,
		       result, error, logs, checkpoint, attempt, restart_count,
		       daemon_backoff_ms, created_at, started_at, done_at, last_healthy_at
		FROM worker_jobs WHERE id = ?
	`, id).Scan(
		&row.ID, &row.AppID, &row.Handler, &row.Status, &row.ConfigJSON,
		&row.Progress, &result, &errorStr, &logsJSON, &checkpoint,
		&row.Attempt, &row.RestartCount, &row.DaemonBackoffMs,
		&createdAt, &startedAt, &doneAt, &lastHealthyAt,
	)
	if err != nil {
		return nil, err
	}

	if result.Valid {
		row.Result = &result.String
	}
	if errorStr.Valid {
		row.Error = &errorStr.String
	}
	if logsJSON.Valid {
		row.LogsJSON = &logsJSON.String
	}
	if checkpoint.Valid {
		row.Checkpoint = &checkpoint.String
	}
	if createdAt.Valid {
		row.CreatedAt = &createdAt.Int64
	}
	if startedAt.Valid {
		row.StartedAt = &startedAt.Int64
	}
	if doneAt.Valid {
		row.DoneAt = &doneAt.Int64
	}
	if lastHealthyAt.Valid {
		row.LastHealthyAt = &lastHealthyAt.Int64
	}

	return JobFromRow(row)
}

func (p *Pool) loadHandlerCode(appID, handler string) (string, error) {
	var content string
	err := p.db.QueryRow(`
		SELECT content FROM files
		WHERE site_id = ? AND path = ?
	`, appID, handler).Scan(&content)
	if err != nil {
		return "", err
	}
	return content, nil
}

func generateJobID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "job_" + hex.EncodeToString(b)
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t.UnixMilli()
}

// watchIdleTimeout monitors listener count and cancels the job if idle for too long.
func (p *Pool) watchIdleTimeout(ctx context.Context, cancel context.CancelFunc, job *Job, reason *string) {
	idleTimeout := *job.Config.IdleTimeout
	channel := job.Config.IdleChannel
	checkInterval := 5 * time.Second

	// Use shorter check interval if timeout is short
	if idleTimeout < 30*time.Second {
		checkInterval = idleTimeout / 6
		if checkInterval < time.Second {
			checkInterval = time.Second
		}
	}

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	var idleSince *time.Time

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count := p.listenerCountFn(job.AppID, channel)

			if count == 0 {
				// No listeners
				if idleSince == nil {
					now := time.Now()
					idleSince = &now
					debug.Log("worker", "job %s: no listeners on channel '%s', will stop in %v",
						job.ID, channel, idleTimeout)
				} else if time.Since(*idleSince) >= idleTimeout {
					// Idle timeout reached
					*reason = fmt.Sprintf("No listeners on channel '%s' for %v, stopping",
						channel, idleTimeout)
					debug.Log("worker", "job %s: %s", job.ID, *reason)
					cancel()
					return
				}
			} else {
				// Has listeners, reset idle timer
				if idleSince != nil {
					debug.Log("worker", "job %s: listeners returned (%d), resetting idle timer",
						job.ID, count)
					idleSince = nil
				}
			}
		}
	}
}
