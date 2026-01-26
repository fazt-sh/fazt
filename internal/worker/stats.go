package worker

// WorkerHealthStats provides worker pool statistics for the health endpoint.
type WorkerHealthStats struct {
	Active      int     `json:"active"`
	Queued      int     `json:"queued"`
	Total       int     `json:"total"`
	MemoryUsed  int64   `json:"memory_used_bytes"`
	MemoryPool  int64   `json:"memory_pool_bytes"`
	MemoryPct   float64 `json:"memory_used_pct"`
}

// HealthStats returns stats formatted for the health endpoint.
func HealthStats() *WorkerHealthStats {
	stats := Stats()
	if stats == nil {
		return nil
	}
	return &WorkerHealthStats{
		Active:     stats.ActiveJobs,
		Queued:     stats.QueuedJobs,
		Total:      stats.TotalJobs,
		MemoryUsed: stats.AllocatedMemory,
		MemoryPool: stats.PoolMemory,
		MemoryPct:  stats.MemoryUsedPct,
	}
}
