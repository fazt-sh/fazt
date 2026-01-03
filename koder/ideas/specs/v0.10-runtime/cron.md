# JS-Cron: Scheduled Execution

## Summary

JS-Cron enables scheduled serverless function execution. Apps define triggers
in `app.json`, and the kernel invokes handlers at specified intervals.

## Rationale

### The Problem

Many apps need background tasks:
- Cleanup expired sessions
- Sync external data
- Send scheduled notifications
- Generate reports

Without cron, apps need external schedulers or always-on processes.

### The Solution

Declarative scheduling in `app.json`:

```json
{
  "cron": [
    { "schedule": "0 * * * *", "handler": "api/hourly.js" },
    { "schedule": "0 0 * * *", "handler": "api/daily.js" }
  ]
}
```

## Hibernate Architecture

### The Principle

Functions consume **zero RAM** when idle. They are "hibernating" until:
- HTTP request arrives
- Cron schedule triggers
- External event occurs

```
Timeline:
─────────────────────────────────────────────────────────
  idle    │ execute │    idle      │ execute │   idle
  0 RAM   │ 64 MB   │    0 RAM     │ 64 MB   │   0 RAM
─────────────────────────────────────────────────────────
```

### Benefits

- Run 100s of scheduled tasks on 1GB RAM
- No blocking threads
- No process management

## Configuration

### app.json

```json
{
  "name": "my-app",
  "cron": [
    {
      "schedule": "*/5 * * * *",
      "handler": "api/check-prices.js",
      "timeout": 60000
    },
    {
      "schedule": "0 9 * * 1",
      "handler": "api/weekly-report.js"
    }
  ]
}
```

### Schedule Syntax

Standard cron format: `minute hour day month weekday`

| Pattern       | Meaning               |
| ------------- | --------------------- |
| `* * * * *`   | Every minute          |
| `*/5 * * * *` | Every 5 minutes       |
| `0 * * * *`   | Every hour            |
| `0 0 * * *`   | Every day at midnight |
| `0 9 * * 1`   | Every Monday at 9am   |
| `0 0 1 * *`   | First of every month  |

## Handler Pattern

### Request Object

Cron handlers receive a special request:

```javascript
// api/daily.js
module.exports = async function(request) {
    // request.trigger === 'cron'
    // request.schedule === '0 0 * * *'
    // request.scheduledAt === 1704067200000

    await cleanupExpiredSessions();

    return { success: true };
};
```

### State Persistence

Use `fazt.schedule()` for stateful continuations:

```javascript
// api/sync.js
module.exports = async function(request) {
    const state = request.state || { page: 1 };

    const data = await fetchPage(state.page);
    await fazt.storage.ds.insert('items', data);

    if (data.hasMore) {
        // Schedule continuation in 1 minute
        await fazt.schedule(60000, { page: state.page + 1 });
    }

    return { processed: data.items.length };
};
```

## Execution

### Kernel Scheduler

```go
type Scheduler struct {
    db      *sql.DB
    ticker  *time.Ticker
}

func (s *Scheduler) Run() {
    for range s.ticker.C {
        jobs := s.getDueJobs()
        for _, job := range jobs {
            go s.execute(job)
        }
    }
}

func (s *Scheduler) execute(job CronJob) {
    runtime := NewRuntime(job.AppUUID)
    result := runtime.Run(job.Handler, &Request{
        Trigger:     "cron",
        Schedule:    job.Schedule,
        ScheduledAt: job.DueAt,
        State:       job.State,
    })
    s.recordResult(job, result)
}
```

### Concurrency

- Each cron job runs in its own goroutine
- Multiple apps' crons can run in parallel
- Same app's crons can overlap (be careful)

### Overlap Prevention

```json
{
  "cron": [
    {
      "schedule": "* * * * *",
      "handler": "api/sync.js",
      "skipIfRunning": true
    }
  ]
}
```

## Monitoring

### CLI

```bash
# List scheduled jobs
fazt app cron list app_x9z2k

# View execution history
fazt app cron history app_x9z2k

# Manually trigger
fazt app cron run app_x9z2k api/daily.js
```

### API

```
GET /api/apps/{uuid}/cron
GET /api/apps/{uuid}/cron/history
POST /api/apps/{uuid}/cron/trigger
```

## Resource Limits

| Limit                 | Value     | Rationale               |
| --------------------- | --------- | ----------------------- |
| Max cron jobs per app | 10        | Prevent scheduler flood |
| Min interval          | 1 minute  | Prevent tight loops     |
| Execution timeout     | 5 minutes | Prevent hung jobs       |
| Retry on failure      | 3 times   | Handle transient errors |

## Failure Handling

### Retry Policy

```json
{
  "cron": [
    {
      "schedule": "0 * * * *",
      "handler": "api/sync.js",
      "retry": {
        "attempts": 3,
        "backoff": "exponential"
      }
    }
  ]
}
```

### Dead Letter Queue

Failed jobs after retries go to a dead letter log:

```sql
SELECT * FROM cron_failures
WHERE app_uuid = 'app_x9z2k'
ORDER BY failed_at DESC;
```

## Open Questions

1. **Timezone handling**: UTC only, or per-app timezone?
2. **Miss handling**: If server was down, run missed jobs?
3. **Priority**: Can critical crons preempt others?
