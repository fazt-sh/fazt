# Background Jobs

## Summary

A persistent job queue for long-running tasks. Jobs are spawned, queued,
executed by workers, and can report progress. Failed jobs retry with
configurable backoff and eventually land in a dead-letter queue.

## Job Lifecycle

```
    spawn()
       │
       ▼
   ┌───────┐
   │QUEUED │ ──── waiting for worker slot
   └───┬───┘
       │
       ▼
   ┌───────┐
   │RUNNING│ ──── executing handler
   └───┬───┘
       │
   ┌───┴───┐
   │       │
   ▼       ▼
┌─────┐ ┌──────┐
│DONE │ │FAILED│
└─────┘ └──┬───┘
           │
           ▼ (if retries remain)
       ┌───────┐
       │QUEUED │
       └───────┘
           │
           ▼ (if retries exhausted)
       ┌──────────┐
       │DEAD_LETTER│
       └──────────┘
```

## Spawning Jobs

### From Serverless Handler

```javascript
// api/process.js - HTTP endpoint that spawns a job
module.exports = async (req) => {
    const job = await fazt.worker.spawn('workers/resize.js', {
        data: {
            imageUrl: req.json.url,
            sizes: [100, 200, 400, 800]
        },
        timeout: '5m',
        retry: 3
    });

    return {
        json: {
            jobId: job.id,
            status: job.status
        }
    };
};
```

### Spawn Options

```javascript
await fazt.worker.spawn(handler, {
    // Required
    data: { ... },           // Passed to handler

    // Resource budget
    memory: '32MB',          // Memory allocation (default: 32MB, max: pool size)
    timeout: '5m',           // Max runtime (default: 30m, null = indefinite)

    // Daemon mode (for long-running workers)
    daemon: false,           // Restart on crash (default: false)

    // Retry & scheduling
    retry: 3,                // Retry attempts (default: 0)
    retryDelay: '1m',        // Delay between retries (default: 1m)
    retryBackoff: 'exponential', // 'fixed' | 'exponential'
    priority: 'normal',      // 'low' | 'normal' | 'high'
    delay: '10s',            // Delay before first run
    uniqueKey: 'resize-123', // Prevent duplicate jobs
});
```

### Resource Budget

Workers draw from a shared memory pool (default: 256MB total).

```javascript
// Small job - takes 32MB from pool
await fazt.worker.spawn('workers/thumbnail.js', {
    data: { imageId: 123 },
    memory: '32MB'           // Default
});

// Large job - takes 128MB, fewer concurrent jobs possible
await fazt.worker.spawn('workers/video-encode.js', {
    data: { videoId: 456 },
    memory: '128MB'
});

// Greedy job - takes entire pool, runs alone
await fazt.worker.spawn('workers/ml-inference.js', {
    data: { model: 'large' },
    memory: '256MB'          // Other workers queue until this finishes
});
```

**Pool behavior:**
- Jobs request memory at spawn time
- If pool has capacity → starts immediately
- If pool full → queues until memory available
- Memory tracked via `runtime.MemStats` (soft limit)
- Exceeded jobs get 500ms grace period, then interrupted

### Daemon Mode

For workers that should run continuously (traffic simulators, data streams):

```javascript
// Start a daemon worker
const job = await fazt.worker.spawn('workers/traffic-sim.js', {
    data: { scenario: 'rush-hour' },
    daemon: true,            // Restart on crash
    memory: '64MB',          // Memory budget
    timeout: null            // Run indefinitely
});

// Daemon runs until explicitly stopped
await fazt.worker.cancel(job.id);
```

**Daemon lifecycle:**
```
    spawn(daemon: true)
           │
           ▼
       ┌───────┐
       │RUNNING│ ◄─────────────────┐
       └───┬───┘                   │
           │                       │
       ┌───┴───┐                   │
       │       │                   │
       ▼       ▼                   │
   ┌──────┐ ┌──────┐    restart    │
   │DONE  │ │CRASH │───────────────┘
   └──────┘ └──────┘   (with backoff)
       │
       ▼
   (explicit cancel)
```

**Restart backoff:** 1s → 2s → 4s → 8s → ... → 60s (max)

**Survival:** Daemons survive server restart via checkpoint:
```javascript
module.exports = async function(job) {
    // Restore state after restart
    let state = job.checkpoint() || { tick: 0 };

    while (!job.cancelled) {
        state.tick++;
        await doWork(state);

        // Checkpoint every 100 ticks for crash recovery
        if (state.tick % 100 === 0) {
            job.checkpoint(state);
        }

        await sleep(1000);
    }
};
```

## Worker Handler

### Handler Pattern

```javascript
// workers/resize.js
module.exports = async function(job) {
    const { imageUrl, sizes } = job.data;

    job.log('Starting image resize');
    job.progress(0);

    const results = [];
    for (let i = 0; i < sizes.length; i++) {
        const size = sizes[i];

        job.log(`Resizing to ${size}px`);
        const resized = await resizeImage(imageUrl, size);

        // Store result
        const key = `resized/${job.id}/${size}.jpg`;
        await fazt.storage.s3.put(key, resized, 'image/jpeg');
        results.push({ size, key });

        // Report progress
        job.progress(Math.round(((i + 1) / sizes.length) * 100));
    }

    job.log('Complete');

    // Return value is stored as job result
    return { images: results };
};
```

### Job Object (in handler)

```javascript
job.id              // Job ID
job.data            // Data passed to spawn()
job.attempt         // Current attempt (1, 2, 3...)
job.cancelled       // Boolean: true if cancel requested (for daemon loops)
job.memory          // Allocated memory in bytes
job.daemon          // Boolean: true if daemon mode

job.progress(n)     // Report progress 0-100
job.log(msg)        // Add log entry
job.checkpoint(state) // Save state for resume/crash recovery
```

### Checkpointing (Resume on Failure)

```javascript
module.exports = async function(job) {
    // Restore checkpoint if resuming
    let state = job.checkpoint() || { processed: 0 };

    const items = job.data.items;
    for (let i = state.processed; i < items.length; i++) {
        await processItem(items[i]);

        // Save checkpoint every 10 items
        if (i % 10 === 0) {
            job.checkpoint({ processed: i });
            job.progress(Math.round((i / items.length) * 100));
        }
    }

    return { processed: items.length };
};
```

## Querying Jobs

### Get Job Status

```javascript
const job = await fazt.worker.get(jobId);
// {
//   id: 'job_abc',
//   status: 'running',
//   progress: 45,
//   data: { ... },
//   result: null,
//   error: null,
//   logs: ['Starting...', 'Processing...'],
//   createdAt: ...,
//   startedAt: ...,
//   attempt: 1
// }
```

### List Jobs

```javascript
// All jobs
const jobs = await fazt.worker.list();

// Filtered
const running = await fazt.worker.list({ status: 'running' });
const failed = await fazt.worker.list({ status: 'failed' });
const recent = await fazt.worker.list({ limit: 10, order: 'desc' });
```

### Wait for Completion

```javascript
// Poll until done (with timeout)
const result = await fazt.worker.wait(jobId, { timeout: '10m' });
// Returns job with status 'done' or throws on timeout/failure
```

### Cancel Job

```javascript
await fazt.worker.cancel(jobId);
// Job transitions to 'cancelled' status
// If running, worker receives cancellation signal
```

## Dead-Letter Queue

Jobs that fail all retry attempts go to dead-letter:

```javascript
// List dead-letter jobs
const deadJobs = await fazt.worker.deadLetter.list();

// Inspect
const job = await fazt.worker.deadLetter.get(jobId);

// Retry manually
await fazt.worker.deadLetter.retry(jobId);

// Delete (acknowledge failure)
await fazt.worker.deadLetter.delete(jobId);
```

## Job Storage

### Schema

```sql
CREATE TABLE worker_jobs (
    id TEXT PRIMARY KEY,
    app_uuid TEXT NOT NULL,
    handler TEXT NOT NULL,
    data TEXT,              -- JSON
    status TEXT NOT NULL,   -- queued, running, done, failed, cancelled
    progress INTEGER DEFAULT 0,
    result TEXT,            -- JSON
    error TEXT,
    logs TEXT,              -- JSON array
    checkpoint TEXT,        -- JSON state for resume/crash recovery
    attempt INTEGER DEFAULT 1,
    max_attempts INTEGER DEFAULT 1,
    priority INTEGER DEFAULT 0,
    timeout_ms INTEGER,     -- NULL = indefinite
    memory_bytes INTEGER,   -- Allocated memory budget
    daemon INTEGER DEFAULT 0, -- 1 = restart on crash
    daemon_backoff_ms INTEGER, -- Current backoff delay
    created_at INTEGER,
    started_at INTEGER,
    completed_at INTEGER,
    next_retry_at INTEGER
);

CREATE INDEX idx_jobs_app_status ON worker_jobs(app_uuid, status);
CREATE INDEX idx_jobs_queue ON worker_jobs(status, priority DESC, created_at);
CREATE INDEX idx_jobs_daemon ON worker_jobs(daemon, status);
```

### Cleanup

Completed jobs auto-delete after retention period:

```javascript
// Default: 7 days
fazt config set limits.workers.resultRetentionDays 7
```

## Limits

| Limit                   | Default | Description                    |
| ----------------------- | ------- | ------------------------------ |
| `maxConcurrentTotal`    | 20      | All apps combined              |
| `maxConcurrentPerApp`   | 5       | Per app                        |
| `maxQueueDepth`         | 100     | Queued jobs per app            |
| `maxMemoryPoolMB`       | 256     | Total memory for all workers   |
| `defaultMemoryPerJobMB` | 32      | Default per-job allocation     |
| `maxMemoryPerJobMB`     | 256     | Max single job can request     |
| `defaultTimeoutMinutes` | 30      | Default timeout (null=forever) |
| `maxDaemonsPerApp`      | 2       | Max daemon workers per app     |
| `maxDataSizeKB`         | 1024    | Job data payload               |
| `resultRetentionDays`   | 7       | Keep completed jobs            |

### Limit Behavior

- **Queue full**: `spawn()` returns error
- **Concurrent limit**: Job waits in queue
- **Memory pool full**: Job waits until memory available
- **Memory exceeded**: 500ms grace period, then interrupted
- **Timeout exceeded**: Job interrupted, marked failed (unless daemon)
- **Daemon crash**: Auto-restart with backoff
- **At 80%**: Warning logged

## Concurrency Control

### Unique Keys

Prevent duplicate jobs:

```javascript
await fazt.worker.spawn('workers/sync.js', {
    data: { userId: 123 },
    uniqueKey: `sync-user-123`
});

// Second spawn with same uniqueKey returns existing job
const job2 = await fazt.worker.spawn('workers/sync.js', {
    data: { userId: 123 },
    uniqueKey: `sync-user-123`
});
// job2.id === job1.id (if still running/queued)
```

### Priority

```javascript
// High priority jumps the queue
await fazt.worker.spawn('workers/urgent.js', {
    data: { ... },
    priority: 'high'
});
```

## CLI

```bash
# List jobs
fazt worker list --app app_uuid --status running

# List daemon workers
fazt worker list --daemon

# View job
fazt worker show job_abc

# Cancel job (stops daemon permanently)
fazt worker cancel job_abc

# View resource usage
fazt worker pool
# Output:
# Memory Pool: 256 MB
# Allocated:   128 MB (50%)
# Available:   128 MB
# Active jobs: 4
#   job_abc (daemon)  64 MB  workers/traffic-sim.js
#   job_def           32 MB  workers/resize.js
#   job_ghi           16 MB  workers/notify.js
#   job_jkl           16 MB  workers/sync.js

# View dead-letter queue
fazt worker dead-letter list

# Retry dead-letter job
fazt worker dead-letter retry job_abc

# Purge old jobs
fazt worker purge --older-than 7d
```

## Events Integration

Workers can be triggered by other events:

```javascript
// On email received, spawn processing job
// api/email.js
module.exports = async (req) => {
    if (req.event === 'email') {
        await fazt.worker.spawn('workers/process-email.js', {
            data: { emailId: req.email.id }
        });
    }
};
```

## Realtime Progress

Combine with v0.17 Realtime for live progress:

```javascript
// workers/export.js
module.exports = async function(job) {
    for (let i = 0; i <= 100; i += 10) {
        await doWork();
        job.progress(i);

        // Broadcast progress to connected clients
        await fazt.realtime.broadcast(`job-${job.id}`, {
            progress: i
        });
    }
    return { url: exportUrl };
};
```

```javascript
// Client
rt.subscribe(`job-${jobId}`);
rt.on('message', (channel, data) => {
    progressBar.style.width = `${data.progress}%`;
});
```

## Example: PDF Report Generator

```javascript
// workers/generate-report.js
module.exports = async function(job) {
    const { reportType, dateRange } = job.data;

    job.log('Fetching data');
    job.progress(10);

    const data = await fazt.storage.ds.find('sales', {
        date: { $gte: dateRange.start, $lte: dateRange.end }
    });

    job.log(`Found ${data.length} records`);
    job.progress(30);

    job.log('Generating PDF');
    const pdf = await generatePDF(reportType, data);
    job.progress(80);

    job.log('Uploading');
    const key = `reports/${job.id}.pdf`;
    await fazt.storage.s3.put(key, pdf, 'application/pdf');
    job.progress(100);

    const url = await fazt.fs.ipfsUrl(key);

    job.log('Complete');
    return { url, records: data.length };
};
```

## Go Implementation

Uses [backlite](https://github.com/mikestefanello/backlite) - type-safe SQLite
task queues designed for embedding:

```go
import "github.com/mikestefanello/backlite"

// Define task type with queue config
type ResizeImage struct {
    ImageURL string   `json:"imageUrl"`
    Sizes    []int    `json:"sizes"`
}

func (t ResizeImage) Config() backlite.QueueConfig {
    return backlite.QueueConfig{
        Name:        "resize",
        MaxAttempts: 3,
        Backoff:     time.Minute,
    }
}

// Register handler
queue := backlite.NewQueue[ResizeImage](func(ctx context.Context, task ResizeImage) error {
    // Process resize...
    return nil
})
client.Register(queue)

// Spawn from serverless (via internal API)
client.Add(ResizeImage{ImageURL: url, Sizes: sizes}).
    Wait(10 * time.Second).  // Delay
    Tx(tx).                   // In transaction
    Save()
```

**Why backlite**:
- Pure Go (CGO only in tests, driver-agnostic)
- SQLite-native (fits cartridge model)
- Type-safe generics
- Transaction support (spawn in app's transaction)
- No polling (notification pattern)
- ~50KB binary impact

## Example: Traffic Simulator (Daemon)

A long-running worker that generates simulated traffic data and broadcasts
via WebSocket:

```javascript
// api/simulator.js - HTTP endpoint to start/stop simulator
module.exports = async function(req) {
    if (req.method === 'POST') {
        // Start daemon
        const job = await fazt.worker.spawn('workers/traffic-sim.js', {
            data: {
                scenario: req.json.scenario || 'normal',
                tickMs: req.json.tickMs || 1000
            },
            daemon: true,        // Restart on crash
            memory: '64MB',      // Memory budget
            timeout: null,       // Run forever
            uniqueKey: 'traffic-sim'  // Only one instance
        });
        return { json: { jobId: job.id, status: 'started' } };
    }

    if (req.method === 'DELETE') {
        // Stop daemon
        const jobs = await fazt.worker.list({
            uniqueKey: 'traffic-sim',
            status: 'running'
        });
        if (jobs.length > 0) {
            await fazt.worker.cancel(jobs[0].id);
            return { json: { status: 'stopped' } };
        }
        return { status: 404, json: { error: 'Not running' } };
    }

    // GET - check status
    const jobs = await fazt.worker.list({
        uniqueKey: 'traffic-sim'
    });
    return { json: { running: jobs.some(j => j.status === 'running') } };
};
```

```javascript
// workers/traffic-sim.js - The daemon worker
module.exports = async function(job) {
    const { scenario, tickMs } = job.data;

    // Restore state from checkpoint (survives crash/restart)
    let state = job.checkpoint() || {
        tick: 0,
        vehicles: generateInitialVehicles(scenario)
    };

    job.log(`Starting traffic simulator: ${scenario}`);

    // Main loop - runs until cancelled
    while (!job.cancelled) {
        state.tick++;

        // Simulate traffic movement
        state.vehicles = simulateStep(state.vehicles, scenario);

        // Broadcast to connected WebSocket clients
        await fazt.realtime.broadcast('traffic', {
            tick: state.tick,
            vehicles: state.vehicles.length,
            avgSpeed: calculateAvgSpeed(state.vehicles),
            congestion: calculateCongestion(state.vehicles)
        });

        // Checkpoint every 100 ticks for crash recovery
        if (state.tick % 100 === 0) {
            job.checkpoint(state);
            job.log(`Checkpoint at tick ${state.tick}`);
        }

        // Wait for next tick
        await sleep(tickMs);
    }

    job.log('Simulator stopped');
    return { finalTick: state.tick };
};

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}
```

```javascript
// Client-side: connect to traffic stream
const ws = new WebSocket('wss://my-app.domain.com/_ws');

ws.onopen = () => {
    ws.send(JSON.stringify({ type: 'subscribe', channel: 'traffic' }));
};

ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    if (msg.channel === 'traffic') {
        updateDashboard(msg.data);
    }
};
```

**Key patterns:**
- `daemon: true` - restarts on crash with backoff
- `timeout: null` - runs indefinitely
- `uniqueKey` - ensures only one instance
- `job.cancelled` - checked in loop for graceful shutdown
- `job.checkpoint()` - survives server restart
- `fazt.realtime.broadcast()` - streams to WebSocket clients
