# Plan 15: Kernel API Specification

**Date**: December 10, 2025
**Status**: Draft / RFC
**Depends On**: Plan 14 (Admin SPA Phase 2C) ✅
**Blocks**: Phase 3 (Real Data Integration)

---

## 1. Executive Summary

This document defines the **v0.8 "Kernel" API** — a unified, RESTful interface that treats
Fazt as an application runtime rather than a web server. The key paradigm shifts:

1. **Sites → Apps**: Everything deployed is an "App" with a stable UUID
2. **Features → System Apps**: Webhooks, redirects, analytics become optional "system apps"
3. **API-First**: Every capability is accessible via API; CLI and Dashboard are clients
4. **Source Tracking**: Apps know where they came from (personal, system, marketplace)

---

## 2. The App Entity

### 2.1 Core Schema

```sql
CREATE TABLE apps (
    -- Identity (immutable)
    id          TEXT PRIMARY KEY,           -- 'app_x7k2m9p4' (stable UUID)

    -- Display
    name        TEXT NOT NULL,              -- 'my-blog' (user-facing, mutable)

    -- Source & Provenance
    source      TEXT NOT NULL DEFAULT 'personal',  -- 'system' | 'personal' | 'marketplace'
    source_url  TEXT,                       -- Git URL for marketplace apps
    version     TEXT,                       -- Semver for marketplace apps

    -- Runtime Configuration
    spa_mode    INTEGER DEFAULT 0,          -- Route all to index.html
    clean_urls  INTEGER DEFAULT 1,          -- Strip .html
    dir_listing INTEGER DEFAULT 0,          -- Show file browser

    -- Metadata
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL,
    deployed_by TEXT                        -- 'cli' | 'dashboard' | 'mcp' | app_id
);

-- Domains are pointers to apps (many-to-one)
CREATE TABLE domains (
    id          TEXT PRIMARY KEY,
    domain      TEXT NOT NULL UNIQUE,       -- 'blog.fazt.sh' or 'custom.com'
    app_id      TEXT NOT NULL REFERENCES apps(id),
    is_primary  INTEGER DEFAULT 0,
    created_at  TEXT NOT NULL
);

-- VFS keyed to app_id (not domain)
CREATE TABLE files (
    id          TEXT PRIMARY KEY,
    app_id      TEXT NOT NULL REFERENCES apps(id),
    path        TEXT NOT NULL,
    content     BLOB,
    mime_type   TEXT,
    size        INTEGER,
    created_at  TEXT NOT NULL,
    UNIQUE(app_id, path)
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

### 2.3 Source Types

| Source        | Description                              | Example                          |
|---------------|------------------------------------------|----------------------------------|
| `system`      | Built into binary, pinned in RAM         | Dashboard, Welcome, 404          |
| `personal`    | Deployed via CLI/Dashboard               | User's blog, portfolio           |
| `marketplace` | Installed from git repo                  | Community apps, templates        |

---

## 3. API Design

### 3.1 Design Principles

1. **REST + JSON**: Standard HTTP verbs, JSON request/response
2. **Envelope Format**: All responses wrapped in `{ data, meta, error }`
3. **Consistent Naming**: Plural nouns, kebab-case for URL params
4. **Pagination**: `?limit=N&offset=M` for list endpoints
5. **Filtering**: `?status=active&source=personal` for queries

### 3.2 Endpoint Map

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
POST   /api/apps                Create app (empty or from deploy)
GET    /api/apps/{id}           Get app details
PUT    /api/apps/{id}           Update app settings
DELETE /api/apps/{id}           Delete app and all files

GET    /api/apps/{id}/files     List files (tree structure)
GET    /api/apps/{id}/files/*   Get file content
POST   /api/apps/{id}/deploy    Deploy files to app
```

#### Domains
```
GET    /api/apps/{id}/domains   List domains for app
POST   /api/apps/{id}/domains   Add domain to app
DELETE /api/apps/{id}/domains/{domain}  Remove domain
PUT    /api/apps/{id}/domains/{domain}  Set as primary
```

#### Environment Variables
```
GET    /api/apps/{id}/env       List env vars (values masked)
POST   /api/apps/{id}/env       Add env var
PUT    /api/apps/{id}/env/{key} Update env var
DELETE /api/apps/{id}/env/{key} Delete env var
```

#### API Keys (per-app)
```
GET    /api/apps/{id}/keys      List API keys (tokens masked)
POST   /api/apps/{id}/keys      Create key (returns full token once)
DELETE /api/apps/{id}/keys/{kid} Revoke key
```

#### Logs
```
GET    /api/apps/{id}/logs      Get app logs
       ?level=error             Filter by level
       &since=2024-12-01        Filter by date
       &limit=100               Pagination
```

#### System
```
GET    /api/system/health       Health check, uptime, version
GET    /api/system/stats        Aggregate stats (total apps, storage)
GET    /api/system/limits       Resource thresholds
GET    /api/system/cache        VFS cache stats
GET    /api/system/config       Server configuration (sanitized)
```

#### Analytics (Global)
```
GET    /api/analytics/events    List events (paginated)
GET    /api/analytics/stats     Aggregate analytics
       ?app_id=app_x7k2m9p4     Filter by app
       ?domain=blog.fazt.sh     Filter by domain
```

#### Traffic Features (Global)
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

#### Marketplace (Future)
```
GET    /api/marketplace/repos           List configured repos
POST   /api/marketplace/repos           Add repo URL
DELETE /api/marketplace/repos/{id}      Remove repo

GET    /api/marketplace/apps            Search available apps
POST   /api/marketplace/install         Install app from repo
POST   /api/apps/{id}/update            Update marketplace app
```

### 3.3 Response Envelope

```typescript
// Success
{
  "data": { ... },              // The payload
  "meta": {                     // Optional metadata
    "total": 42,                // For paginated lists
    "limit": 10,
    "offset": 0
  }
}

// Error
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid domain format",
    "field": "domain",          // Optional: which field failed
    "details": { ... }          // Optional: additional context
  }
}
```

### 3.4 Error Codes

| Code                | HTTP | Description                        |
|---------------------|------|------------------------------------|
| `UNAUTHORIZED`      | 401  | Not logged in                      |
| `FORBIDDEN`         | 403  | Logged in but not allowed          |
| `NOT_FOUND`         | 404  | Resource doesn't exist             |
| `VALIDATION_ERROR`  | 400  | Invalid input                      |
| `CONFLICT`          | 409  | Resource already exists            |
| `RATE_LIMITED`      | 429  | Too many requests                  |
| `INTERNAL_ERROR`    | 500  | Server error                       |

---

## 4. Migration Path

### 4.1 Database Changes

```sql
-- Step 1: Add new columns to existing sites table
ALTER TABLE sites ADD COLUMN source TEXT DEFAULT 'personal';
ALTER TABLE sites ADD COLUMN source_url TEXT;
ALTER TABLE sites ADD COLUMN spa_mode INTEGER DEFAULT 0;
ALTER TABLE sites ADD COLUMN clean_urls INTEGER DEFAULT 1;
ALTER TABLE sites ADD COLUMN dir_listing INTEGER DEFAULT 0;
ALTER TABLE sites ADD COLUMN deployed_by TEXT;

-- Step 2: Generate stable IDs for existing sites
-- (Done in Go migration code)

-- Step 3: Create domains table and migrate
CREATE TABLE domains (...);
INSERT INTO domains (id, domain, app_id, is_primary, created_at)
SELECT 'dom_' || hex(randomblob(4)), domain, id, 1, created_at
FROM sites;

-- Step 4: Rename sites -> apps
ALTER TABLE sites RENAME TO apps;
```

### 4.2 Breaking Changes

| Old                        | New                              | Notes                    |
|----------------------------|----------------------------------|--------------------------|
| `GET /api/sites`           | `GET /api/apps`                  | Response shape similar   |
| `GET /api/sites/{id}`      | `GET /api/apps/{id}`             | ID format changes        |
| `site.domain`              | `app.domains[]`                  | One-to-many              |
| `POST /api/deploy`         | `POST /api/apps/{id}/deploy`     | Explicit app target      |

### 4.3 Backwards Compatibility

For v0.8, we'll maintain aliases:
- `/api/sites` → `/api/apps` (deprecated warning in response)
- Old site IDs continue to work, mapped to new app IDs

---

## 5. Implementation Phases

### Phase A: Schema Migration (Backend)
1. Add new columns to `sites` table
2. Create `domains` table
3. Write migration to generate app IDs
4. Update VFS to key by app_id

### Phase B: API Layer (Backend)
1. Create `/api/apps` handlers
2. Add `/api/apps/{id}/domains` handlers
3. Add `/api/apps/{id}/env` handlers (new feature)
4. Add `/api/apps/{id}/keys` handlers (new feature)
5. Maintain `/api/sites` as deprecated alias

### Phase C: Dashboard Integration (Frontend)
1. Update API client to use new endpoints
2. Update Sites page → Apps page
3. Wire SiteDetail tabs to real APIs
4. Add domain management UI

### Phase D: CLI Updates
1. `fazt deploy` → targets app by name or ID
2. `fazt apps list` → show all apps
3. `fazt apps create <name>` → create empty app
4. `fazt domains add <app> <domain>` → add domain

---

## 6. System Apps

### 6.1 Reserved App IDs

```go
var SystemApps = map[string]string{
    "app_sys_admin": "Dashboard SPA",
    "app_sys_root":  "Welcome Page (root domain)",
    "app_sys_404":   "Universal 404 Page",
}
```

### 6.2 System App Properties

- **Pinned**: Hydrated at boot, served from RAM
- **Protected**: Cannot be deleted via API
- **Upgradeable**: `fazt server reset-admin` replaces from binary
- **Source**: Always `system`

---

## 7. Open Questions

### 7.1 For Discussion

1. **Webhooks/Redirects as Apps?**
   - Option A: Keep as global features (current)
   - Option B: Make them "system apps" with their own UUIDs
   - Recommendation: Keep as-is for v0.8, revisit in v0.9

2. **Per-App API Keys vs Global Keys?**
   - Option A: Keys scoped to single app
   - Option B: Keys with permission sets (can deploy to X, Y, Z)
   - Recommendation: Per-app for simplicity, global keys are admin-only

3. **App Namespacing?**
   - Should `my-blog` be unique globally or per-user?
   - For single-user Fazt: globally unique is fine
   - Recommendation: Keep simple, unique names

---

## 8. Success Criteria

- [ ] All existing functionality works via new API
- [ ] Dashboard wired to real data (Phase 3 of Plan 14)
- [ ] `fazt deploy` continues to work unchanged
- [ ] No data loss during migration
- [ ] System apps survive upgrade

---

## 9. Appendix: Current vs New Endpoint Mapping

| Current Endpoint              | New Endpoint                    | Status        |
|-------------------------------|---------------------------------|---------------|
| `POST /api/login`             | `POST /api/auth/login`          | Rename        |
| `POST /api/logout`            | `POST /api/auth/logout`         | Rename        |
| `GET /api/auth/status`        | `GET /api/auth/status`          | Keep          |
| `GET /api/user/me`            | `GET /api/auth/me`              | Rename        |
| `GET /api/sites`              | `GET /api/apps`                 | Rename        |
| `GET /api/sites/{id}`         | `GET /api/apps/{id}`            | Rename        |
| `GET /api/sites/{id}/files`   | `GET /api/apps/{id}/files`      | Rename        |
| `POST /api/deploy`            | `POST /api/apps/{id}/deploy`    | **Change**    |
| `GET /api/stats`              | `GET /api/analytics/stats`      | Rename        |
| `GET /api/events`             | `GET /api/analytics/events`     | Rename        |
| `GET /api/redirects`          | `GET /api/redirects`            | Keep          |
| `GET /api/webhooks`           | `GET /api/webhooks`             | Keep          |
| `GET /api/system/health`      | `GET /api/system/health`        | Keep          |
| `GET /api/system/config`      | `GET /api/system/config`        | Keep          |
| `GET /api/keys`               | `GET /api/apps/{id}/keys`       | **Change**    |
| `GET /api/envvars`            | `GET /api/apps/{id}/env`        | **Change**    |
| `GET /api/logs`               | `GET /api/apps/{id}/logs`       | **Change**    |
| `GET /api/deployments`        | Removed (use app.updated_at)    | **Remove**    |
| `GET /api/domains`            | `GET /api/apps/{id}/domains`    | **Change**    |
| `GET /api/tags`               | Removed (use app metadata)      | **Remove**    |

---

**Plan Status**: Draft
**Next Action**: Review with user, then implement Phase A (Schema Migration)
