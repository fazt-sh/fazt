# Plan 15: Kernel API Specification

**Date**: December 10, 2025
**Status**: Draft / RFC (v2)
**Depends On**: Plan 14 (Admin SPA Phase 2C) ✅
**Blocks**: Phase 3 (Real Data Integration)

---

## 1. Executive Summary

This document defines the **v0.8 "Kernel" API** — a unified, RESTful interface that treats
Fazt as an application runtime rather than a web server. The key paradigm shifts:

1. **Sites → Apps**: Everything deployed is an "App" with a stable UUID
2. **Domains Decoupled**: Apps have UUIDs; domains are just pointers
3. **Hibernate Architecture**: Agents don't "run" — they schedule wake-ups
4. **API-First**: Every capability is accessible via API; CLI and Dashboard are clients
5. **Simple Sources**: No marketplace — just `system`, `personal`, or `git:<url>`

---

## 2. The App Entity

### 2.1 Core Schema

```sql
CREATE TABLE apps (
    -- Identity (immutable)
    id          TEXT PRIMARY KEY,           -- 'app_x7k2m9p4' (stable UUID)

    -- Display
    name        TEXT NOT NULL UNIQUE,       -- 'my-blog' (user-facing, mutable)

    -- Source & Provenance
    source      TEXT NOT NULL DEFAULT 'personal',
    -- Values: 'system' | 'personal' | 'git:<url>'

    -- Runtime Configuration
    spa_mode    INTEGER DEFAULT 0,          -- Route all to index.html
    clean_urls  INTEGER DEFAULT 1,          -- Strip .html
    dir_listing INTEGER DEFAULT 0,          -- Show file browser

    -- Metadata
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL,
    deployed_by TEXT                        -- 'cli' | 'dashboard' | 'mcp' | 'scheduler'
);

-- Domains are pointers to apps (many-to-one)
CREATE TABLE domains (
    id          TEXT PRIMARY KEY,
    domain      TEXT NOT NULL UNIQUE,       -- 'blog.fazt.sh' or 'custom.com'
    app_id      TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    is_primary  INTEGER DEFAULT 0,
    created_at  TEXT NOT NULL
);

-- VFS keyed to app_id (not domain)
CREATE TABLE files (
    id          TEXT PRIMARY KEY,
    app_id      TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    path        TEXT NOT NULL,
    content     BLOB,
    mime_type   TEXT,
    size        INTEGER,
    created_at  TEXT NOT NULL,
    UNIQUE(app_id, path)
);

-- Environment variables per app
CREATE TABLE env_vars (
    id          TEXT PRIMARY KEY,
    app_id      TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    key         TEXT NOT NULL,
    value       TEXT NOT NULL,              -- Encrypted at rest
    created_at  TEXT NOT NULL,
    UNIQUE(app_id, key)
);

-- API keys per app
CREATE TABLE api_keys (
    id          TEXT PRIMARY KEY,
    app_id      TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    token_hash  TEXT NOT NULL,              -- bcrypt hash
    prefix      TEXT NOT NULL,              -- 'fzt_' for display
    created_at  TEXT NOT NULL,
    last_used   TEXT
);
```

### 2.2 App ID Format

```
app_<nanoid-8>

Examples:
  app_x7k2m9p4    (user app)
  app_sys_admin   (system: dashboard)
  app_sys_root    (system: welcome page)
  app_sys_404     (system: 404 page)
```

### 2.3 Source Types (Simplified)

| Source          | Description                              | Example                          |
|-----------------|------------------------------------------|----------------------------------|
| `system`        | Built into binary, pinned in RAM         | Dashboard, Welcome, 404          |
| `personal`      | Deployed via CLI/Dashboard               | User's blog, portfolio           |
| `git:<url>`     | Installed from git repo                  | `git:https://github.com/x/app`   |

**No marketplace abstraction.** If you want to share apps, share the git URL.
Community can maintain awesome-lists; that's not infrastructure we build.

---

## 3. The Scheduler (Hibernate Architecture)

### 3.1 The Problem

AI agents need to "wait" — check stock prices every 5 minutes, poll APIs, retry failed tasks.
Traditional approach: `while(true) { sleep(5m); work(); }` — this blocks threads and dies on timeout.

### 3.2 The Solution: Interrupt-Driven Execution

Agents don't "run" — they **hibernate**. They do work in milliseconds, then schedule their next wake-up.

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   HTTP      │     │   Jobs DB   │     │   Ticker    │
│   Request   │     │   (sleep)   │     │  (1s loop)  │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       ▼                   │                   │
  ┌─────────┐              │                   │
  │ Run JS  │──schedule───▶│                   │
  │  <100ms │              │                   │
  └─────────┘              │                   │
       │                   │                   │
       ▼                   │    wake_at ≤ now  │
    (done)                 │◀──────────────────┤
                           │                   │
                           ▼                   │
                      ┌─────────┐              │
                      │ Run JS  │              │
                      │  <100ms │              │
                      └─────────┘              │
```

### 3.3 Jobs Table

```sql
CREATE TABLE jobs (
    id          TEXT PRIMARY KEY,
    app_id      TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    wake_at     TEXT NOT NULL,              -- ISO timestamp
    payload     TEXT,                       -- JSON state
    created_at  TEXT NOT NULL,
    status      TEXT DEFAULT 'pending',     -- pending | running | done | failed
    error       TEXT,                       -- Error message if failed

    -- Indexing for efficient polling
    CONSTRAINT idx_pending CHECK (status IN ('pending', 'running', 'done', 'failed'))
);

CREATE INDEX idx_jobs_wake ON jobs(wake_at) WHERE status = 'pending';
CREATE INDEX idx_jobs_app ON jobs(app_id);
```

### 3.4 JavaScript Runtime API

```javascript
// ═══════════════════════════════════════════════════════════════
// INJECTED GLOBALS (read-only)
// ═══════════════════════════════════════════════════════════════

process.trigger   // How this execution was triggered
                  // 'http' | 'schedule' | 'webhook' | 'cron'

process.state     // The payload from fazt.schedule(), or {} on first run

process.app       // { id, name, source }

process.env       // Environment variables (from env_vars table)


// ═══════════════════════════════════════════════════════════════
// FAZT NAMESPACE (actions)
// ═══════════════════════════════════════════════════════════════

fazt.schedule(delay, state)
// Schedule this app to run again
// delay: '30s', '5m', '1h', '1d', or ISO timestamp
// state: JSON object passed to process.state on wake
// Returns: job_id

fazt.cancel(jobId)
// Cancel a scheduled job
// Returns: boolean

fazt.jobs()
// List pending jobs for this app
// Returns: [{ id, wake_at, payload, created_at }]


// ═══════════════════════════════════════════════════════════════
// EXAMPLE: Stock Price Watcher
// ═══════════════════════════════════════════════════════════════

if (process.trigger === 'http') {
    // Initial setup via HTTP POST
    const { symbol, threshold } = request.body;
    const jobId = fazt.schedule('1m', { symbol, threshold, checks: 0 });
    return {
        message: `Watching ${symbol} for price < $${threshold}`,
        job_id: jobId
    };
}

if (process.trigger === 'schedule') {
    const { symbol, threshold, checks } = process.state;

    const res = fetch(`https://api.example.com/stock/${symbol}`);
    const price = res.json().price;

    if (price < threshold) {
        // Alert user and stop
        fetch('https://ntfy.sh/my-alerts', {
            method: 'POST',
            body: `🚨 ${symbol} dropped to $${price}!`
        });
        return { triggered: true, price, total_checks: checks };
    }

    // Keep watching (schedule next check)
    fazt.schedule('1m', { symbol, threshold, checks: checks + 1 });
    return { watching: true, price, checks: checks + 1 };
}
```

### 3.5 Backend Implementation (Go)

```go
// Ticker runs every second, checks for due jobs
func startScheduler(db *sql.DB, runtime *Runtime) {
    ticker := time.NewTicker(time.Second)
    for range ticker.C {
        jobs := getPendingJobs(db, time.Now(), 10) // Batch of 10
        for _, job := range jobs {
            go func(j Job) {
                markJobRunning(db, j.ID)

                ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
                defer cancel()

                result, err := runtime.Execute(ctx, j.AppID, ExecuteOpts{
                    Trigger: "schedule",
                    State:   j.Payload,
                })

                if err != nil {
                    markJobFailed(db, j.ID, err.Error())
                } else {
                    markJobDone(db, j.ID)
                }
            }(job)
        }
    }
}

// Expose fazt.schedule() to Goja
func (r *Runtime) registerScheduler(vm *goja.Runtime, appID string) {
    fazt := vm.NewObject()

    fazt.Set("schedule", func(call goja.FunctionCall) goja.Value {
        delay := call.Argument(0).String()
        state := call.Argument(1).Export()

        wakeAt := parseDelay(delay) // '5m' -> time.Now().Add(5*time.Minute)
        payload, _ := json.Marshal(state)

        jobID := createJob(r.db, appID, wakeAt, string(payload))
        return vm.ToValue(jobID)
    })

    fazt.Set("cancel", func(call goja.FunctionCall) goja.Value {
        jobID := call.Argument(0).String()
        ok := cancelJob(r.db, jobID, appID)
        return vm.ToValue(ok)
    })

    fazt.Set("jobs", func(call goja.FunctionCall) goja.Value {
        jobs := getAppJobs(r.db, appID)
        return vm.ToValue(jobs)
    })

    vm.Set("fazt", fazt)
}
```

### 3.6 Resource Limits

| Limit | Value | Rationale |
|-------|-------|-----------|
| Max jobs per app | 100 | Prevent runaway scheduling |
| Min delay | 10s | Prevent tight loops |
| Max delay | 30d | Reasonable horizon |
| Execution timeout | 100ms | Quick turnaround |
| Payload size | 64KB | Reasonable state |

---

## 4. API Design

### 4.1 Design Principles

1. **REST + JSON**: Standard HTTP verbs, JSON request/response
2. **Envelope Format**: All responses wrapped in `{ data, meta, error }`
3. **Consistent Naming**: Plural nouns, kebab-case paths
4. **Pagination**: `?limit=N&offset=M` for list endpoints

### 4.2 Endpoint Map

#### Authentication
```
POST   /api/auth/login          Login with username/password
POST   /api/auth/logout         Clear session
GET    /api/auth/status         Check if authenticated
GET    /api/auth/me             Current user info
```

#### Apps (Core)
```
GET    /api/apps                List all apps
POST   /api/apps                Create app (returns app_id)
GET    /api/apps/{id}           Get app details
PUT    /api/apps/{id}           Update app settings
DELETE /api/apps/{id}           Delete app and all data

POST   /api/apps/{id}/deploy    Deploy files to app
GET    /api/apps/{id}/files     List files (tree structure)
GET    /api/apps/{id}/files/*   Get file content
```

#### Domains
```
GET    /api/apps/{id}/domains           List domains
POST   /api/apps/{id}/domains           Add domain
DELETE /api/apps/{id}/domains/{domain}  Remove domain
PUT    /api/apps/{id}/domains/{domain}  Set as primary
```

#### Environment Variables
```
GET    /api/apps/{id}/env               List env vars (masked)
POST   /api/apps/{id}/env               Add env var
PUT    /api/apps/{id}/env/{key}         Update env var
DELETE /api/apps/{id}/env/{key}         Delete env var
```

#### API Keys
```
GET    /api/apps/{id}/keys              List keys (masked)
POST   /api/apps/{id}/keys              Create key (returns token once)
DELETE /api/apps/{id}/keys/{kid}        Revoke key
```

#### Logs
```
GET    /api/apps/{id}/logs              Get app logs
       ?level=error&since=...&limit=100
```

#### Jobs (Scheduler)
```
GET    /api/apps/{id}/jobs              List scheduled jobs
DELETE /api/apps/{id}/jobs/{jid}        Cancel job
```

#### System
```
GET    /api/system/health       Health, uptime, version
GET    /api/system/stats        Aggregate stats
GET    /api/system/limits       Resource thresholds
GET    /api/system/config       Server config (sanitized)
```

#### Analytics
```
GET    /api/analytics/events    List events (paginated)
GET    /api/analytics/stats     Aggregate analytics
       ?app_id=...&domain=...
```

#### Traffic (Global)
```
GET    /api/redirects           List redirects
POST   /api/redirects           Create redirect
PUT    /api/redirects/{id}      Update redirect
DELETE /api/redirects/{id}      Delete redirect

GET    /api/webhooks            List webhooks
POST   /api/webhooks            Create webhook
PUT    /api/webhooks/{id}       Update webhook
DELETE /api/webhooks/{id}       Delete webhook
```

#### Git Install (Simple)
```
POST   /api/apps/install        Install app from git URL
       { "url": "https://github.com/user/app", "name": "my-app" }
```

### 4.3 Response Envelope

```typescript
// Success
{
  "data": { ... },
  "meta": {
    "total": 42,
    "limit": 10,
    "offset": 0
  }
}

// Error
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid domain format",
    "field": "domain"
  }
}
```

---

## 5. Migration Path

### 5.1 Database Migration

```sql
-- Step 1: Add columns to sites
ALTER TABLE sites ADD COLUMN source TEXT DEFAULT 'personal';
ALTER TABLE sites ADD COLUMN spa_mode INTEGER DEFAULT 0;
ALTER TABLE sites ADD COLUMN clean_urls INTEGER DEFAULT 1;
ALTER TABLE sites ADD COLUMN dir_listing INTEGER DEFAULT 0;
ALTER TABLE sites ADD COLUMN deployed_by TEXT;

-- Step 2: Create new tables
CREATE TABLE domains (...);
CREATE TABLE env_vars (...);
CREATE TABLE api_keys (...);
CREATE TABLE jobs (...);

-- Step 3: Migrate domains from sites
INSERT INTO domains (id, domain, app_id, is_primary, created_at)
SELECT 'dom_' || hex(randomblob(4)), domain, id, 1, created_at
FROM sites;

-- Step 4: Rename sites -> apps
ALTER TABLE sites RENAME TO apps;

-- Step 5: Update files table
ALTER TABLE files ADD COLUMN app_id TEXT;
UPDATE files SET app_id = site_id;
-- Then drop site_id column
```

### 5.2 Backwards Compatibility

For v0.8 transition:
- `/api/sites` → alias to `/api/apps` (logs deprecation warning)
- Old site IDs continue to work (mapped internally)
- `POST /api/deploy` → works but logs warning to use `/api/apps/{id}/deploy`

---

## 6. Implementation Phases

### Phase A: Schema & Migration
- [ ] Write migration script
- [ ] Add new columns to sites
- [ ] Create domains, env_vars, api_keys, jobs tables
- [ ] Generate stable app IDs
- [ ] Update VFS to key by app_id

### Phase B: Core API
- [ ] `/api/apps` CRUD handlers
- [ ] `/api/apps/{id}/domains` handlers
- [ ] `/api/apps/{id}/env` handlers
- [ ] `/api/apps/{id}/keys` handlers
- [ ] `/api/apps/{id}/jobs` handlers
- [ ] Backwards compat aliases

### Phase C: Scheduler
- [ ] Jobs table implementation
- [ ] Ticker goroutine
- [ ] `fazt.schedule()` in Goja runtime
- [ ] `process.state`, `process.trigger` injection
- [ ] Resource limits enforcement

### Phase D: Dashboard Integration
- [ ] Update API client
- [ ] Wire SiteDetail tabs to real APIs
- [ ] Add Jobs tab to SiteDetail
- [ ] Update Sites → Apps naming

### Phase E: CLI Updates
- [ ] `fazt apps list`
- [ ] `fazt apps create <name>`
- [ ] `fazt apps install <git-url>`
- [ ] `fazt deploy` targets app by name/id

---

## 7. System Apps

### 7.1 Reserved IDs

```go
var SystemApps = map[string]SystemApp{
    "app_sys_admin": {Name: "Dashboard", Pinned: true},
    "app_sys_root":  {Name: "Welcome", Pinned: true},
    "app_sys_404":   {Name: "Not Found", Pinned: true},
}
```

### 7.2 Properties

- **Pinned**: Hydrated at boot, served from RAM
- **Protected**: Cannot be deleted via API
- **Upgradeable**: `fazt server reset-admin` replaces from binary

---

## 8. Success Criteria

- [ ] All existing functionality works via new API
- [ ] Dashboard wired to real data
- [ ] `fazt deploy` continues to work
- [ ] Scheduler runs 1000 jobs without issues
- [ ] No data loss during migration
- [ ] System apps survive upgrade

---

## 9. What We're NOT Building (Scope Control)

| Feature | Status | Rationale |
|---------|--------|-----------|
| Marketplace registry | Removed | Just use git URLs |
| App versioning | Deferred | Use git tags |
| App dependencies | Deferred | Too complex |
| Multi-user | Deferred | Single-admin for v0.8 |
| Cron syntax | Deferred | Start with delays only |

---

## 10. Appendix: Endpoint Migration Table

| Current                       | New                              | Change      |
|-------------------------------|----------------------------------|-------------|
| `POST /api/login`             | `POST /api/auth/login`           | Rename      |
| `GET /api/user/me`            | `GET /api/auth/me`               | Rename      |
| `GET /api/sites`              | `GET /api/apps`                  | Rename      |
| `GET /api/sites/{id}`         | `GET /api/apps/{id}`             | Rename      |
| `POST /api/deploy`            | `POST /api/apps/{id}/deploy`     | **Change**  |
| `GET /api/keys`               | `GET /api/apps/{id}/keys`        | **Change**  |
| `GET /api/envvars`            | `GET /api/apps/{id}/env`         | **Change**  |
| `GET /api/logs`               | `GET /api/apps/{id}/logs`        | **Change**  |
| `GET /api/stats`              | `GET /api/analytics/stats`       | Rename      |
| `GET /api/events`             | `GET /api/analytics/events`      | Rename      |
| `GET /api/redirects`          | `GET /api/redirects`             | Keep        |
| `GET /api/webhooks`           | `GET /api/webhooks`              | Keep        |
| `GET /api/system/*`           | `GET /api/system/*`              | Keep        |
| `GET /api/domains`            | Removed                          | Use app domains |
| `GET /api/tags`               | Removed                          | Not needed  |
| `GET /api/deployments`        | Removed                          | Use app.updated_at |
| N/A                           | `GET /api/apps/{id}/jobs`        | **New**     |
| N/A                           | `POST /api/apps/install`         | **New**     |

---

**Plan Status**: Draft v2
**Key Changes from v1**: Removed marketplace, added scheduler architecture
**Next Action**: Review, then implement Phase A (Schema Migration)
