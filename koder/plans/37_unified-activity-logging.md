# Unified Activity Logging System

## Overview

Consolidate all logging (audit, analytics, system events) into a single `activity_log` table with:
- Weight-based prioritization (0-9)
- Auto-cleanup when exceeding max rows
- Analytics SDK auto-injection
- App-defined custom events
- Bulk export (JSONL, future OTEL)
- Pulse integration foundation

## User Choices
- **Architecture**: Single unified table
- **SDK Injection**: Auto-inject (opt-out via manifest)
- **Default Weight**: 0 (log everything)
- **Export Format**: JSONL now, OTEL later
- **Custom Events**: App specifies weight (flexible)

---

## 1. Database Schema

### Migration: `internal/database/migrations/018_activity_log.sql`

```sql
CREATE TABLE IF NOT EXISTS activity_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL DEFAULT (unixepoch()),

    -- Actor
    actor_type TEXT NOT NULL DEFAULT 'system',  -- user/system/api_key/anonymous/app
    actor_id TEXT,
    actor_ip TEXT,
    actor_ua TEXT,

    -- Resource
    resource_type TEXT NOT NULL,  -- app/user/session/kv/doc/page/config/custom
    resource_id TEXT,
    app_id TEXT,                  -- For app-scoped events (filtering)

    -- Action
    action TEXT NOT NULL,
    result TEXT DEFAULT 'success',

    -- Weight (0-9, higher = more important)
    weight INTEGER NOT NULL DEFAULT 2,

    -- Details (JSON)
    details TEXT
);

-- Indexes for queries
CREATE INDEX idx_activity_log_timestamp ON activity_log(timestamp);
CREATE INDEX idx_activity_log_weight ON activity_log(weight);
CREATE INDEX idx_activity_log_cleanup ON activity_log(weight, timestamp);
CREATE INDEX idx_activity_log_resource ON activity_log(resource_type, resource_id);
CREATE INDEX idx_activity_log_app ON activity_log(app_id);
CREATE INDEX idx_activity_log_actor ON activity_log(actor_type, actor_id);
```

### Weight Scale

| Weight | Category | Examples | Default Retention |
|--------|----------|----------|-------------------|
| 9 | Security | API key CRUD, role changes | 90 days |
| 8 | Auth | Login/logout, sessions | 30 days |
| 7 | Config | Alias CRUD, redirect CRUD | 30 days |
| 6 | Deployment | Site deploy/delete | 30 days |
| 5 | Data Mutation | KV set/delete, doc CRUD | 14 days |
| 4 | User Action | Form submit, add-to-cart (app) | 7 days |
| 3 | Navigation | Internal page views | 3 days |
| 2 | Analytics | Pageview, click (default) | 3 days |
| 1 | System | Health checks, backups | 1 day |
| 0 | Debug | Request timing, cache hits | 1 day |

---

## 2. Activity Logger Package

### New File: `internal/activity/logger.go`

```go
package activity

// Weight constants
const (
    WeightDebug      = 0
    WeightSystem     = 1
    WeightAnalytics  = 2
    WeightNavigation = 3
    WeightUserAction = 4
    WeightDataMutation = 5
    WeightDeployment = 6
    WeightConfig     = 7
    WeightAuth       = 8
    WeightSecurity   = 9
)

// ActorType constants
const (
    ActorSystem    = "system"
    ActorUser      = "user"
    ActorAPIKey    = "api_key"
    ActorAnonymous = "anonymous"
    ActorApp       = "app"  // For app-generated events
)

// Entry represents a log entry
type Entry struct {
    Timestamp    time.Time
    ActorType    string
    ActorID      string
    ActorIP      string
    ActorUA      string
    ResourceType string
    ResourceID   string
    AppID        string   // For app-scoped events
    Action       string
    Result       string
    Weight       int
    Details      map[string]interface{}
}

// Config
type Config struct {
    FlushInterval time.Duration  // 10s
    BatchSize     int            // 500
    MaxRows       int            // 500000 (~100MB)
    CleanupBatch  int            // 10000
    MinWeight     int            // 0 (log everything)
}
```

**Key functions:**
- `Init(db, config)` - initialize
- `Log(Entry)` - buffer an entry
- `LogSuccess(...)` / `LogFailure(...)` - helpers
- `LogFromRequest(r, user, ...)` - extract HTTP context
- `Shutdown()` - flush and stop

### New File: `internal/activity/query.go`

```go
// QueryParams for filtering
type QueryParams struct {
    MinWeight    *int
    MaxWeight    *int
    ActorType    string
    ActorID      string
    ResourceType string
    ResourceID   string
    AppID        string      // Filter by app
    Action       string
    Result       string      // success/failure
    Since        *time.Time
    Until        *time.Time
    Offset       int         // Pagination
    Limit        int         // Pagination (default 50, max 1000)
    Sample       float64     // 0.0-1.0, for random sampling
}

func Query(db *sql.DB, params QueryParams) ([]LogEntry, int, error)
func Cleanup(db *sql.DB, minWeight int, olderThan time.Duration) (int64, error)
func GetStats(db *sql.DB) (map[string]interface{}, error)
func Export(db *sql.DB, params QueryParams, w io.Writer, format string) error
```

---

## 3. CLI Commands

### `fazt logs list`

```bash
fazt logs list [options]
  --min-weight N      Minimum weight (0-9)
  --max-weight N      Maximum weight (0-9)
  --type TYPE         Filter by resource_type
  --app APP_ID        Filter by app_id
  --actor ACTOR_ID    Filter by actor_id
  --action ACTION     Filter by action
  --since DURATION    Since (e.g., 1h, 24h, 7d)
  --until DURATION    Until
  --result RESULT     success/failure
  --offset N          Skip first N results (default: 0)
  --limit N           Max results (default: 50, max: 1000)
  --sample RATE       Random sample rate (0.0-1.0)
  --format FORMAT     Output format: table, json (default: table)
```

### `fazt logs cleanup`

```bash
fazt logs cleanup [options]
  --weight-below N    Delete logs with weight < N
  --older-than DUR    Delete logs older than (e.g., 7d, 30d)
  --dry-run           Show what would be deleted
  --force             Skip confirmation
```

### `fazt logs stats`

```bash
fazt logs stats [--format json]
# Shows: total rows, size estimate, count by weight, oldest/newest
```

### `fazt logs export`

```bash
fazt logs export [options]
  --min-weight N      Minimum weight
  --since DURATION    Since
  --until DURATION    Until
  --app APP_ID        Filter by app
  --format FORMAT     jsonl (default), csv
  --output FILE       Output file (default: stdout)
  --compress          Gzip output
```

**JSONL format** (one JSON object per line):
```jsonl
{"id":1,"timestamp":1707000000,"actor_type":"user","actor_id":"usr_123","resource_type":"session","action":"login","weight":8}
{"id":2,"timestamp":1707000001,"actor_type":"anonymous","resource_type":"page","resource_id":"zyt.app/","action":"pageview","weight":2}
```

---

## 4. App-Defined Custom Events

### JS Runtime API

```javascript
// In serverless functions / app code
fazt.activity.log({
    action: 'add-to-cart',
    weight: 4,  // App specifies weight (0-9)
    resource_type: 'cart',
    resource_id: 'cart_abc123',
    details: {
        product_id: 'prod_xyz',
        quantity: 2,
        price: 29.99
    }
});

// Convenience helpers
fazt.activity.debug('cache_hit', { key: 'user:123' });        // weight: 0
fazt.activity.info('form_submit', { form: 'contact' });       // weight: 4
fazt.activity.warn('rate_limited', { ip: '1.2.3.4' });        // weight: 7
fazt.activity.error('payment_failed', { order: 'ord_123' });  // weight: 8
```

### Handler: `internal/runtime/activity.go`

```go
// Register in Goja VM
vm.Set("fazt", map[string]interface{}{
    "activity": map[string]interface{}{
        "log":   activityLog,
        "debug": func(action string, details map[string]interface{}) { ... },
        "info":  func(action string, details map[string]interface{}) { ... },
        "warn":  func(action string, details map[string]interface{}) { ... },
        "error": func(action string, details map[string]interface{}) { ... },
    },
})

func activityLog(opts map[string]interface{}) {
    activity.Log(activity.Entry{
        ActorType:    activity.ActorApp,
        ActorID:      currentAppID,
        AppID:        currentAppID,
        Action:       opts["action"].(string),
        Weight:       opts["weight"].(int),  // App controls weight
        ResourceType: opts["resource_type"].(string),
        ResourceID:   opts["resource_id"].(string),
        Details:      opts["details"].(map[string]interface{}),
    })
}
```

---

## 5. SDK Injection

### Script (minified ~250 bytes)

```javascript
(function(){
  var h=location.hostname,p=location.pathname,q=location.search;
  navigator.sendBeacon('/track',JSON.stringify({h:h,p:p,q:q,e:'pageview'}));
})();
```

### Manifest opt-out

```json
{
  "name": "my-app",
  "analytics": { "enabled": false }
}
```

### Implementation: `internal/hosting/handler.go`

In `ServeVFS()`, before serving HTML:
```go
if isHTML && !isAnalyticsDisabled(siteID) {
    data = injectAnalytics(data)
}
```

---

## 6. Pulse Integration

The activity_log table provides the "Recent Events" data for Pulse cognitive observability:

```go
// In Pulse beat collection
recentEvents := activity.Query(db, activity.QueryParams{
    MinWeight: activity.WeightDataMutation, // Only significant events
    Since:     &lastBeatTime,
    Limit:     100,
})

// Feed to Pulse analysis
pulsePrompt.RecentEvents = formatEventsForPulse(recentEvents)
```

**Pulse queries activity_log for:**
- Recent deploys (weight 6)
- Auth events (weight 8)
- Errors (result='failure')
- App-defined events (custom analytics)

---

## 7. API Endpoints

### `GET /api/logs` (admin)

```
?min_weight=5&app_id=app_123&since=24h&offset=0&limit=50&format=json
```

### `DELETE /api/logs/cleanup` (admin)

```json
{ "weight_below": 4, "older_than": "14d" }
```

### `GET /api/logs/stats` (admin)

Returns row counts, size estimate, weight distribution.

### `GET /api/logs/export` (admin)

Streams JSONL for bulk export.

---

## 8. Files Summary

### Create

| File | Purpose |
|------|---------|
| `internal/database/migrations/018_activity_log.sql` | Schema |
| `internal/activity/logger.go` | Buffer, logging |
| `internal/activity/query.go` | Query, cleanup, export |
| `internal/hosting/analytics_inject.go` | SDK injection |
| `internal/runtime/activity.go` | JS API for apps |
| `cmd/server/logs.go` | CLI commands |

### Modify

| File | Change |
|------|--------|
| `cmd/server/main.go` | Init, shutdown, CLI routing |
| `internal/hosting/handler.go` | Analytics injection |
| `internal/handlers/auth_handlers.go` | Replace audit.Log* |
| `internal/handlers/track.go` | Use activity.Log |
| `internal/hosting/deploy.go` | Log deployments |
| `internal/storage/kv.go` | Log KV ops |
| `internal/storage/ds.go` | Log doc ops |

### Remove (Phase 3)

| File | Reason |
|------|--------|
| `internal/audit/audit.go` | Replaced |

---

## 9. Implementation Order

1. **Schema** - `018_activity_log.sql`
2. **Core Package** - `internal/activity/logger.go`, `query.go`
3. **Server Init** - main.go init/shutdown
4. **CLI** - `logs list`, `logs cleanup`, `logs stats`, `logs export`
5. **API Endpoints** - `/api/logs/*`
6. **Instrumentation** - Auth → Deploy → Storage → Track
7. **JS Runtime API** - `fazt.activity.log()` for apps
8. **SDK Injection** - Auto-inject analytics
9. **Pulse Hookup** - Query activity_log in Pulse beats

---

## 10. Verification

```bash
# CLI tests
fazt logs stats
fazt logs list --min-weight 5 --app app_123 --format json
fazt logs cleanup --weight-below 2 --older-than 1h --dry-run
fazt logs export --since 24h --format jsonl > logs.jsonl

# SDK injection
curl -s http://myapp.192.168.64.3.nip.io:8080/ | grep sendBeacon

# Custom app events
# Deploy app with fazt.activity.log() call, verify in activity_log

# Database
sqlite3 ~/.fazt/data.db "SELECT weight, COUNT(*) FROM activity_log GROUP BY weight"
```

---

## Future: OTEL Support

For external log viewer integration (Grafana, Datadog), add OTEL export:

```bash
fazt logs export --format otel --output http://localhost:4317
```

This connects to OTEL collectors for enterprise observability pipelines.
