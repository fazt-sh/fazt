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

    // Optional
    timeout: '5m',           // Max runtime (default: 5m, max: 30m)
    retry: 3,                // Retry attempts (default: 0)
    retryDelay: '1m',        // Delay between retries (default: 1m)
    retryBackoff: 'exponential', // 'fixed' | 'exponential'
    priority: 'normal',      // 'low' | 'normal' | 'high'
    delay: '10s',            // Delay before first run
    uniqueKey: 'resize-123', // Prevent duplicate jobs
});
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
job.id          // Job ID
job.data        // Data passed to spawn()
job.attempt     // Current attempt (1, 2, 3...)
job.progress(n) // Report progress 0-100
job.log(msg)    // Add log entry
job.checkpoint(state) // Save state for resume
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
    checkpoint TEXT,        -- JSON
    attempt INTEGER DEFAULT 1,
    max_attempts INTEGER DEFAULT 1,
    priority INTEGER DEFAULT 0,
    timeout_ms INTEGER,
    created_at INTEGER,
    started_at INTEGER,
    completed_at INTEGER,
    next_retry_at INTEGER
);

CREATE INDEX idx_jobs_app_status ON worker_jobs(app_uuid, status);
CREATE INDEX idx_jobs_queue ON worker_jobs(status, priority DESC, created_at);
```

### Cleanup

Completed jobs auto-delete after retention period:

```javascript
// Default: 7 days
fazt config set limits.workers.resultRetentionDays 7
```

## Limits

| Limit                 | Default | Description         |
| --------------------- | ------- | ------------------- |
| `maxConcurrentTotal`  | 20      | All apps combined   |
| `maxConcurrentPerApp` | 5       | Per app             |
| `maxQueueDepth`       | 100     | Queued jobs per app |
| `maxRuntimeMinutes`   | 30      | Single job          |
| `maxDataSizeKB`       | 1024    | Job data payload    |
| `resultRetentionDays` | 7       | Keep completed jobs |

### Limit Behavior

- **Queue full**: `spawn()` returns error
- **Concurrent limit**: Job waits in queue
- **Timeout exceeded**: Job killed, marked failed
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

# View job
fazt worker show job_abc

# Cancel job
fazt worker cancel job_abc

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
