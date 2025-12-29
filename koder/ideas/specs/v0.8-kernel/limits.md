# System Limits

## Summary

Per-subsystem soft limits with graceful degradation. Each kernel component
has configurable thresholds that trigger warnings at 80% and enforcement
at 100%. Owners can adjust limits based on their hardware.

## Rationale

Fazt runs on constrained hardware (1GB RAM, $6 VPS). Without limits:
- One app exhausts all connections
- One job consumes all memory
- System becomes unresponsive

Limits provide:
- **Fairness**: Apps share resources
- **Stability**: System survives load spikes
- **Visibility**: Owner knows when scaling is needed

## Limit Structure

```javascript
fazt.limits = {
    global: { ... },      // System-wide
    vfs: { ... },         // Virtual filesystem
    runtime: { ... },     // JS execution
    storage: { ... },     // KV, DS, S3
    realtime: { ... },    // WebSockets (v0.17)
    email: { ... },       // Email sink (v0.18)
    workers: { ... },     // Background jobs (v0.19)
}
```

## Global Limits

```yaml
global:
  maxMemoryMB: 512          # Kernel heap limit
  maxDiskPercent: 80        # Reject writes beyond
  maxCpuPercent: 90         # Throttle at this level
```

## VFS Limits

```yaml
vfs:
  maxCacheMB: 128           # LRU cache size
  maxFileSizeMB: 100        # Single file upload
  maxAppSizeMB: 500         # Total per app
  maxTotalSizeGB: 10        # All apps combined
```

## Runtime Limits

```yaml
runtime:
  maxExecutionSeconds: 30   # Per request
  maxMemoryMB: 64           # Per invocation
  maxConcurrent: 50         # Parallel executions
```

## Storage Limits

```yaml
storage:
  kv:
    maxKeySizeBytes: 1024   # Key length
    maxValueSizeMB: 1       # Single value
    maxKeysPerApp: 100000   # Keys per app
  ds:
    maxDocumentSizeMB: 16   # Single document
    maxDocsPerCollection: 1000000
  s3:
    maxBlobSizeMB: 100      # Single blob
    maxTotalPerAppGB: 5     # Total per app
```

## Realtime Limits (v0.17)

```yaml
realtime:
  maxConnectionsTotal: 5000       # All apps
  maxConnectionsPerApp: 500       # Per app
  maxChannelsPerApp: 100          # Channels per app
  maxSubscriptionsPerClient: 20   # Per connection
  maxMessageSizeKB: 64            # Single message
  maxMessagesPerSecond: 100       # Per connection
  idleTimeoutSeconds: 300         # Disconnect idle
```

## Email Limits (v0.18)

```yaml
email:
  maxInboundPerHour: 1000         # All apps
  maxInboundPerAppPerHour: 100    # Per app
  maxMessageSizeMB: 10            # Including attachments
  maxAttachmentSizeMB: 25         # Single attachment
  maxAttachmentsPerEmail: 20
  retentionDays: 30               # Auto-delete
```

## Worker Limits (v0.19)

```yaml
workers:
  maxConcurrentTotal: 20          # All apps
  maxConcurrentPerApp: 5          # Per app
  maxQueueDepth: 100              # Queued per app
  maxRuntimeMinutes: 30           # Single job
  maxDataSizeKB: 1024             # Job payload
  resultRetentionDays: 7          # Keep completed
```

## Rate Limiting

Request rate limiting per client/app. Protects against abuse, brute force,
and ensures fair access. Uses token bucket algorithm with per-key sliding windows.

### Default Rates

```yaml
rate:
  global:
    requestsPerSecond: 1000       # All apps combined
  perApp:
    requestsPerSecond: 100        # Per app
    requestsPerMinute: 2000       # Per app
  perClient:
    requestsPerSecond: 10         # Per IP
    requestsPerMinute: 300        # Per IP
  perEndpoint:
    requestsPerSecond: 5          # Per endpoint per IP
```

### Rate Limit Behavior

When rate limit is exceeded:

```http
HTTP/1.1 429 Too Many Requests
Retry-After: 1
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1703980845
```

### Custom Rate Limits

Per-endpoint rate limiting in `app.json`:

```json
{
  "rate": {
    "api/login": {
      "perClient": { "perMinute": 5 },
      "message": "Too many login attempts"
    },
    "api/signup": {
      "perClient": { "perMinute": 3 }
    },
    "api/search": {
      "perClient": { "perSecond": 2 }
    },
    "api/webhook": {
      "perClient": { "perSecond": 50 }
    }
  }
}
```

### Rate Limiting in JS

```javascript
// Check rate limit status
const status = await fazt.limits.rate.status(clientIp);
// { remaining: 8, limit: 10, reset: 1703980845 }

// Custom rate limiting logic
const allowed = await fazt.limits.rate.check('custom-key', {
  limit: 5,
  window: 60  // seconds
});

if (!allowed) {
  return { status: 429, body: 'Rate limited' };
}

// Consume multiple tokens (for expensive operations)
const allowed = await fazt.limits.rate.consume('api-key-123', {
  limit: 100,
  window: 3600,  // per hour
  cost: 10       // this request costs 10 tokens
});
```

### Rate Limit Keys

Built-in key extraction:

| Key | Description |
|-----|-------------|
| `ip` | Client IP address (default) |
| `ip+endpoint` | IP + request path |
| `user` | Authenticated user ID |
| `apiKey` | API key from header |
| `custom` | Custom key from header/cookie |

Configure in `app.json`:

```json
{
  "rate": {
    "keyBy": "ip",              // Default
    "api/v1/*": {
      "keyBy": "apiKey",        // API key from X-API-Key header
      "perClient": { "perHour": 1000 }
    }
  }
}
```

### Whitelist/Blacklist

```json
{
  "rate": {
    "whitelist": ["10.0.0.0/8", "192.168.0.0/16"],
    "blacklist": ["1.2.3.4"]
  }
}
```

Whitelisted IPs bypass rate limits. Blacklisted IPs get immediate 403.

### Rate Limit Storage

Uses efficient in-memory sliding window:

```go
// Token bucket with sliding window
type RateLimiter struct {
    buckets sync.Map  // key -> *bucket
    cleanup *time.Ticker
}

type bucket struct {
    tokens    float64
    lastCheck time.Time
    mu        sync.Mutex
}
```

- Memory efficient: ~100 bytes per active key
- Auto-cleanup: Expired buckets removed every minute
- No persistence needed: Limits reset on restart (by design)

## Behavior at Thresholds

### At 80% (Warning)

```
[WARN] realtime: 4000/5000 connections (80%)
[WARN] storage.kv: 80000/100000 keys for app_xyz (80%)
```

- Log warning
- Notify owner (if notifications configured)
- Continue accepting requests

### At 100% (Enforcement)

| Subsystem | Action |
|-----------|--------|
| VFS | Reject upload with 413 |
| Runtime | Queue or reject with 503 |
| Realtime | Reject connection with 503 |
| Email | Reject with SMTP 452 |
| Workers | Reject spawn, return error |
| Storage | Reject write with 507 |

### At 100% (Graceful Degradation)

Some subsystems degrade instead of reject:

```
realtime:
  at100: degrade    # Disconnect oldest idle connections

vfs:
  at100: evict      # Evict LRU cache entries
```

## Configuration

### View Limits

```bash
fazt limits show

# Output:
# Subsystem       Limit                   Current    Status
# global          maxMemoryMB: 512        312        OK (61%)
# realtime        maxConnections: 5000    4200       WARN (84%)
# workers         maxConcurrent: 20       18         WARN (90%)
```

### Set Limits

```bash
# Adjust single limit
fazt config set limits.realtime.maxConnections 10000

# Dangerous mode (exceed recommended)
fazt config set limits.realtime.maxConnections 50000 --force
# Warning: Exceeding recommended limits may cause instability
```

### Per-App Limits

Apps can have individual limits (lower than system):

```json
// app.json
{
  "limits": {
    "realtime.maxConnections": 100,
    "workers.maxConcurrent": 2,
    "storage.kv.maxKeys": 10000
  }
}
```

### Reset to Defaults

```bash
fazt limits reset
fazt limits reset --subsystem realtime
```

## API

### HTTP Endpoint

```
GET /api/system/limits
```

Response:
```json
{
  "global": {
    "maxMemoryMB": { "value": 512, "current": 312, "percent": 61 }
  },
  "realtime": {
    "maxConnectionsTotal": { "value": 5000, "current": 4200, "percent": 84, "warning": true }
  }
}
```

### JS Runtime

```javascript
// Read limits (read-only in app context)
const limits = await fazt.kernel.limits();
console.log(limits.realtime.maxConnectionsTotal);
// { value: 5000, current: 4200, percent: 84 }
```

## Storage

Limits stored in kernel config:

```sql
CREATE TABLE kernel_config (
    key TEXT PRIMARY KEY,
    value TEXT,
    updated_at INTEGER
);

-- Example entries
-- key: 'limits.realtime.maxConnectionsTotal', value: '5000'
-- key: 'limits.workers.maxConcurrentPerApp', value: '5'
```

## Notifications

When thresholds are crossed:

```yaml
notifications:
  on: [warning, critical]
  channels:
    - ntfy:mytopic
    - email:admin@example.com
```

Warning at 80%, critical at 95%.

## Presets

Common configurations:

```bash
# Minimal (512MB RAM, $5 VPS)
fazt limits preset minimal

# Standard (1GB RAM, $6 VPS) - default
fazt limits preset standard

# Performance (2GB+ RAM)
fazt limits preset performance

# Custom from file
fazt limits import limits.yaml
```

### Preset: Minimal

```yaml
global.maxMemoryMB: 256
vfs.maxCacheMB: 64
realtime.maxConnectionsTotal: 1000
workers.maxConcurrentTotal: 5
```

### Preset: Standard

```yaml
global.maxMemoryMB: 512
vfs.maxCacheMB: 128
realtime.maxConnectionsTotal: 5000
workers.maxConcurrentTotal: 20
```

### Preset: Performance

```yaml
global.maxMemoryMB: 1024
vfs.maxCacheMB: 256
realtime.maxConnectionsTotal: 20000
workers.maxConcurrentTotal: 50
```

## Monitoring Dashboard

The OS shell (`os.<domain>`) displays:
- Real-time usage gauges per subsystem
- Historical usage graphs
- Alert history
- One-click limit adjustment
