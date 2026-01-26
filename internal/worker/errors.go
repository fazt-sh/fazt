package worker

import "errors"

// Common errors
var (
	ErrPoolNotInitialized = errors.New("worker pool not initialized")
	ErrJobNotFound        = errors.New("job not found")
	ErrQueueFull          = errors.New("job queue full")
	ErrDaemonLimitReached = errors.New("max daemon workers reached")
	ErrMemoryPoolFull     = errors.New("memory pool exhausted")
)
