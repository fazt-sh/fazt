# Rate Limiting

## Summary

Kernel-level rate limiting primitive using token bucket algorithm. Provides
cross-cutting rate control for APIs, jobs, notifications, and external calls.

## Why

Rate limiting is needed everywhere:
- API endpoints (per-user, per-IP)
- Background jobs (don't overwhelm external APIs)
- Notifications (don't spam users)
- Webhooks (respect target rate limits)

Without a primitive, every app builds ad-hoc solutions using KV with TTL.
These are error-prone (race conditions), repeated across apps, and not
observable at platform level.

**Philosophy alignment:**
- Single binary: ~200 lines, no deps
- Single database: SQLite-backed counters
- Events as spine: Emits `limits.exceeded` events
- JSON everywhere: Status returned as JSON

## API

### JS Runtime

```javascript
// Check if action is allowed (doesn't consume)
const status = await fazt.limits.check('api:user:123', {
  limit: 100,      // max requests
  window: '1m'     // time window
});
// { allowed: true, remaining: 73, reset: 1704307200000 }

// Check and consume in one call
const status = await fazt.limits.consume('api:user:123', {
  limit: 100,
  window: '1m',
  cost: 1          // tokens to consume (default: 1)
});
// { allowed: true, remaining: 72, reset: 1704307200000 }
// { allowed: false, remaining: 0, reset: 1704307200000, retryAfter: 23000 }

// Get current status without affecting quota
const status = await fazt.limits.status('api:user:123');

// Reset a key (admin operation)
await fazt.limits.reset('api:user:123');
```

### Key Naming Convention

```
{scope}:{resource}:{identifier}

api:user:123          # API calls by user
api:ip:192.168.1.1    # API calls by IP
notify:user:123       # Notifications to user
webhook:target:xyz    # Outbound webhook to target
job:queue:emails      # Job queue throughput
```

### Window Formats

```
'1s'   → 1 second
'30s'  → 30 seconds
'1m'   → 1 minute
'5m'   → 5 minutes
'1h'   → 1 hour
'1d'   → 1 day
```

### CLI

```bash
# Show rate limit status
fazt limits show api:user:123
# Key: api:user:123
# Limit: 100/1m
# Remaining: 73
# Reset: 2026-01-03T12:00:00Z

# List all active limits
fazt limits list [--prefix api:]

# Reset a specific key
fazt limits reset api:user:123

# Bulk reset by prefix
fazt limits reset --prefix api:user:
```

## Events

```javascript
// Emitted when limit exceeded
{
  type: 'limits.exceeded',
  data: {
    key: 'api:user:123',
    limit: 100,
    window: '1m',
    reset: 1704307200000
  }
}

// Emitted when approaching limit (configurable threshold)
{
  type: 'limits.warning',
  data: {
    key: 'api:user:123',
    remaining: 10,
    threshold: 10,  // warning at 10% remaining
    limit: 100
  }
}
```

## Common Patterns

### API Rate Limiting

```javascript
// api/main.js
module.exports = async function(request) {
  const userId = request.user?.id || request.ip;

  const { allowed, remaining, retryAfter } = await fazt.limits.consume(
    `api:user:${userId}`,
    { limit: 100, window: '1m' }
  );

  if (!allowed) {
    return {
      status: 429,
      headers: {
        'Retry-After': Math.ceil(retryAfter / 1000),
        'X-RateLimit-Remaining': 0
      },
      body: { error: 'Rate limit exceeded' }
    };
  }

  // Proceed with request...
  return {
    headers: { 'X-RateLimit-Remaining': remaining },
    body: { data: '...' }
  };
};
```

### External API Throttling

```javascript
// Respect external API limits
async function callExternalAPI(endpoint, data) {
  const { allowed, retryAfter } = await fazt.limits.consume(
    `external:${endpoint}`,
    { limit: 60, window: '1m' }  // 60 req/min
  );

  if (!allowed) {
    // Queue for later or wait
    await sleep(retryAfter);
    return callExternalAPI(endpoint, data);
  }

  return fetch(endpoint, { body: JSON.stringify(data) });
}
```

### Notification Throttling

```javascript
// Don't spam users
async function notifyUser(userId, message) {
  const { allowed } = await fazt.limits.consume(
    `notify:user:${userId}`,
    { limit: 10, window: '1h' }  // max 10/hour
  );

  if (!allowed) {
    // Queue for digest or skip
    await fazt.storage.kv.set(`notify:queued:${userId}`, message);
    return;
  }

  await fazt.services.notify.send({ to: userId, ...message });
}
```

### Sliding Window with Burst

```javascript
// Allow burst but limit sustained rate
const { allowed } = await fazt.limits.consume('api:user:123', {
  limit: 100,
  window: '1m',
  burst: 20  // allow 20 extra in first second
});
```

## Implementation

### Schema

```sql
CREATE TABLE rate_limits (
  key TEXT PRIMARY KEY,
  count INTEGER NOT NULL DEFAULT 0,
  window_start INTEGER NOT NULL,  -- Unix ms
  window_ms INTEGER NOT NULL,
  max_count INTEGER NOT NULL
);

CREATE INDEX idx_rate_limits_expiry ON rate_limits(window_start);
```

### Algorithm (Token Bucket / Fixed Window)

```go
func (r *RateLimiter) Consume(key string, limit int, windowMs int64, cost int) Status {
    now := time.Now().UnixMilli()
    windowStart := now - (now % windowMs)  // Align to window boundary

    // Atomic upsert and increment
    result := r.db.Exec(`
        INSERT INTO rate_limits (key, count, window_start, window_ms, max_count)
        VALUES (?, ?, ?, ?, ?)
        ON CONFLICT(key) DO UPDATE SET
            count = CASE
                WHEN window_start < ? THEN ?  -- New window, reset
                ELSE count + ?                 -- Same window, increment
            END,
            window_start = CASE
                WHEN window_start < ? THEN ?
                ELSE window_start
            END
        RETURNING count, window_start
    `, key, cost, windowStart, windowMs, limit,
       windowStart, cost, cost,
       windowStart, windowStart)

    remaining := limit - result.count
    allowed := remaining >= 0
    reset := result.windowStart + windowMs

    return Status{
        Allowed:    allowed,
        Remaining:  max(0, remaining),
        Reset:      reset,
        RetryAfter: reset - now,
    }
}
```

### Garbage Collection

Background goroutine cleans expired windows:

```go
func (r *RateLimiter) GC() {
    // Run every minute
    cutoff := time.Now().UnixMilli() - (24 * time.Hour).Milliseconds()
    r.db.Exec(`DELETE FROM rate_limits WHERE window_start < ?`, cutoff)
}
```

## Configuration

```bash
# Set default warning threshold (emit event at X% remaining)
fazt config set limits.warning_threshold 0.1  # 10%

# Set GC interval
fazt config set limits.gc_interval 60s
```

## Binary Impact

~200 lines of Go. No external dependencies. Uses existing SQLite connection.
