# v0.19 - Workers

**Theme**: Long-running background jobs with queue semantics.

## Summary

v0.19 adds a job queue system for tasks that exceed normal request timeouts.
Workers can run for minutes, report progress, retry on failure, and be
monitored from the dashboard.

## Goals

1. **Long-Running Tasks**: Up to 30 minutes (configurable)
2. **Job Queue**: Persistent, survives restarts
3. **Progress Tracking**: Report progress to clients
4. **Retry & Dead-Letter**: Handle failures gracefully
5. **Isolation**: Apps can't see each other's jobs

## Key Capabilities

| Capability | Description |
|------------|-------------|
| `fazt.worker.spawn()` | Create background job |
| Progress reporting | `job.progress(percent)` |
| Job queue | SQLite-backed, persistent |
| Retry policies | Configurable per job |
| Dead-letter queue | Failed jobs for inspection |

## Documents

- `jobs.md` - Job lifecycle, API, and patterns

## Dependencies

- v0.10 (Runtime): JS execution engine
- v0.9 (Storage): Job state persistence

## Difference from Cron

| Cron | Workers |
|------|---------|
| Scheduled, recurring | On-demand |
| Short-lived (30s default) | Long-running (30m default) |
| Fire and forget | Track progress, get result |
| No queue | Queued, persistent |
