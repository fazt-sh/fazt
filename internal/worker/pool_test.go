package worker

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// testDB creates an in-memory SQLite database for testing.
func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create worker_jobs table
	_, err = db.Exec(`
		CREATE TABLE worker_jobs (
			id TEXT PRIMARY KEY,
			app_id TEXT NOT NULL,
			handler TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			config TEXT DEFAULT '{}',
			progress REAL DEFAULT 0.0,
			result TEXT,
			error TEXT,
			logs TEXT DEFAULT '[]',
			checkpoint TEXT,
			attempt INTEGER DEFAULT 1,
			restart_count INTEGER DEFAULT 0,
			daemon_backoff_ms INTEGER DEFAULT 0,
			created_at INTEGER,
			started_at INTEGER,
			done_at INTEGER,
			last_healthy_at INTEGER
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create worker_jobs table: %v", err)
	}

	// Create files table for handler loading
	_, err = db.Exec(`
		CREATE TABLE files (
			id INTEGER PRIMARY KEY,
			site_id TEXT NOT NULL,
			path TEXT NOT NULL,
			content TEXT,
			mime_type TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create files table: %v", err)
	}

	return db
}

func TestPoolBasicSpawn(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	cfg := DefaultPoolConfig()
	cfg.MaxConcurrentTotal = 2
	pool := NewPool(db, cfg)
	defer pool.Shutdown(context.Background())

	// Set up a simple executor that completes immediately
	var executed sync.WaitGroup
	executed.Add(1)
	pool.SetExecutor(func(ctx context.Context, job *Job, code string) (interface{}, error) {
		executed.Done()
		return map[string]string{"status": "ok"}, nil
	})

	// Insert a test handler
	_, err := db.Exec(`INSERT INTO files (site_id, path, content) VALUES (?, ?, ?)`,
		"app-1", "workers/test.js", "return { ok: true };")
	if err != nil {
		t.Fatalf("Failed to insert test handler: %v", err)
	}

	// Spawn a job
	job, err := pool.Spawn("app-1", "workers/test.js", DefaultJobConfig())
	if err != nil {
		t.Fatalf("Spawn error: %v", err)
	}

	if job.ID == "" {
		t.Error("Job ID should not be empty")
	}
	if job.AppID != "app-1" {
		t.Errorf("Job AppID = %s, want app-1", job.AppID)
	}
	if job.Handler != "workers/test.js" {
		t.Errorf("Job Handler = %s, want workers/test.js", job.Handler)
	}

	// Wait for execution
	done := make(chan struct{})
	go func() {
		executed.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("Job did not execute within timeout")
	}
}

func TestPoolUniqueKey(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	cfg := DefaultPoolConfig()
	pool := NewPool(db, cfg)
	defer pool.Shutdown(context.Background())

	// Don't set executor - jobs will stay queued

	// Insert a test handler
	db.Exec(`INSERT INTO files (site_id, path, content) VALUES (?, ?, ?)`,
		"app-1", "workers/test.js", "return true;")

	// Spawn first job with unique key
	jobCfg := DefaultJobConfig()
	jobCfg.UniqueKey = "unique-task-1"
	job1, err := pool.Spawn("app-1", "workers/test.js", jobCfg)
	if err != nil {
		t.Fatalf("First spawn error: %v", err)
	}

	// Spawn second job with same unique key - should return existing
	job2, err := pool.Spawn("app-1", "workers/test.js", jobCfg)
	if err != nil {
		t.Fatalf("Second spawn error: %v", err)
	}

	if job1.ID != job2.ID {
		t.Errorf("Jobs with same uniqueKey should have same ID: %s != %s", job1.ID, job2.ID)
	}
}

func TestPoolDaemonLimit(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	cfg := DefaultPoolConfig()
	cfg.MaxDaemonsPerApp = 2
	pool := NewPool(db, cfg)
	defer pool.Shutdown(context.Background())

	// Insert a test handler
	db.Exec(`INSERT INTO files (site_id, path, content) VALUES (?, ?, ?)`,
		"app-1", "workers/test.js", "return true;")

	// Spawn daemons up to limit
	daemonCfg := DefaultJobConfig()
	daemonCfg.Daemon = true

	_, err := pool.Spawn("app-1", "workers/test.js", daemonCfg)
	if err != nil {
		t.Fatalf("First daemon spawn error: %v", err)
	}

	_, err = pool.Spawn("app-1", "workers/test.js", daemonCfg)
	if err != nil {
		t.Fatalf("Second daemon spawn error: %v", err)
	}

	// Third daemon should fail
	_, err = pool.Spawn("app-1", "workers/test.js", daemonCfg)
	if err == nil {
		t.Error("Third daemon spawn should fail due to limit")
	}
}

func TestPoolCancel(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	cfg := DefaultPoolConfig()
	pool := NewPool(db, cfg)
	defer pool.Shutdown(context.Background())

	// Executor that waits for cancellation
	var cancelled bool
	var mu sync.Mutex
	pool.SetExecutor(func(ctx context.Context, job *Job, code string) (interface{}, error) {
		<-ctx.Done()
		mu.Lock()
		cancelled = true
		mu.Unlock()
		return nil, ctx.Err()
	})

	// Insert a test handler
	db.Exec(`INSERT INTO files (site_id, path, content) VALUES (?, ?, ?)`,
		"app-1", "workers/test.js", "return true;")

	// Spawn a job
	job, _ := pool.Spawn("app-1", "workers/test.js", DefaultJobConfig())

	// Give time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel
	err := pool.Cancel(job.ID)
	if err != nil {
		t.Fatalf("Cancel error: %v", err)
	}

	// Wait for cancellation to propagate
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	wasCancelled := cancelled
	mu.Unlock()

	if !wasCancelled {
		t.Error("Job should have been cancelled")
	}
}

func TestPoolMemoryAllocation(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	cfg := DefaultPoolConfig()
	cfg.MemoryPoolBytes = 100 * 1024 * 1024 // 100MB
	pool := NewPool(db, cfg)
	defer pool.Shutdown(context.Background())

	// Test allocation
	if !pool.allocateMemory(50 * 1024 * 1024) {
		t.Error("Should be able to allocate 50MB")
	}

	if !pool.allocateMemory(40 * 1024 * 1024) {
		t.Error("Should be able to allocate another 40MB")
	}

	// Should fail - only 10MB left
	if pool.allocateMemory(20 * 1024 * 1024) {
		t.Error("Should not be able to allocate 20MB when only 10MB available")
	}

	// Release and retry
	pool.releaseMemory(40 * 1024 * 1024)
	if !pool.allocateMemory(20 * 1024 * 1024) {
		t.Error("Should be able to allocate 20MB after release")
	}
}

func TestPoolStats(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	cfg := DefaultPoolConfig()
	cfg.MemoryPoolBytes = 256 * 1024 * 1024
	pool := NewPool(db, cfg)
	defer pool.Shutdown(context.Background())

	stats := pool.Stats()
	if stats.PoolMemory != 256*1024*1024 {
		t.Errorf("PoolMemory = %d, want %d", stats.PoolMemory, 256*1024*1024)
	}
	if stats.AllocatedMemory != 0 {
		t.Errorf("AllocatedMemory = %d, want 0", stats.AllocatedMemory)
	}
	if stats.ActiveJobs != 0 {
		t.Errorf("ActiveJobs = %d, want 0", stats.ActiveJobs)
	}
}

func TestPoolGracefulShutdown(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	cfg := DefaultPoolConfig()
	cfg.MaxConcurrentTotal = 1
	pool := NewPool(db, cfg)

	// Executor that takes some time
	pool.SetExecutor(func(ctx context.Context, job *Job, code string) (interface{}, error) {
		select {
		case <-time.After(100 * time.Millisecond):
			return "done", nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})

	// Insert a test handler
	db.Exec(`INSERT INTO files (site_id, path, content) VALUES (?, ?, ?)`,
		"app-1", "workers/test.js", "return true;")

	// Spawn a job
	pool.Spawn("app-1", "workers/test.js", DefaultJobConfig())

	// Give time to start
	time.Sleep(50 * time.Millisecond)

	// Shutdown with sufficient timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := pool.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown error: %v", err)
	}

	// Job may or may not have completed depending on timing
	// The important thing is shutdown completed without hanging
}

func TestPoolList(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	cfg := DefaultPoolConfig()
	pool := NewPool(db, cfg)
	defer pool.Shutdown(context.Background())

	// Insert a test handler
	db.Exec(`INSERT INTO files (site_id, path, content) VALUES (?, ?, ?)`,
		"app-1", "workers/test.js", "return true;")

	// Spawn some jobs
	pool.Spawn("app-1", "workers/test.js", DefaultJobConfig())
	pool.Spawn("app-1", "workers/test.js", DefaultJobConfig())

	// List jobs
	jobs, err := pool.List("app-1", nil, 10)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	if len(jobs) != 2 {
		t.Errorf("Jobs count = %d, want 2", len(jobs))
	}
}
