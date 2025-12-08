# Admin API Analysis & Redesign Recommendations

**Date:** December 8, 2025
**Version:** v0.7.1 baseline
**Scope:** Complete API surface analysis for frontend rebuild

---

## Executive Summary

The current API grew organically alongside features. For a frontend rebuild, we need:
1. **Consistency** - Uniform patterns across all endpoints
2. **Completeness** - Full CRUD where applicable
3. **Observability** - Visibility into system health
4. **Security** - Audit trails and rate limiting visibility

This document proposes **23 new/modified endpoints** across 4 priority tiers.

---

## Part 1: Missing Endpoints to Build

### Priority 1: Critical for Frontend (Must Have)

| Endpoint | Method | Purpose | Complexity |
|----------|--------|---------|------------|
| `/api/sites/{id}` | GET | Single site details (files, size, deployments) | Low |
| `/api/sites/{id}/files` | GET | List files in a site (tree view) | Medium |
| `/api/redirects/{id}` | DELETE | Delete a redirect | Low |
| `/api/webhooks/{id}` | DELETE | Delete a webhook | Low |
| `/api/webhooks/{id}` | PUT | Toggle active/update webhook | Low |
| `/api/system/health` | GET | Detailed health (RAM, CPU, DB size, uptime) | Medium |
| `/api/version` | GET | Version info + update availability | Low |

**Rationale:** These are table-stakes for any admin dashboard. Users expect to drill into site details, delete tracking configs, and see system health.

---

### Priority 2: Important for UX (Should Have)

| Endpoint | Method | Purpose | Complexity |
|----------|--------|---------|------------|
| `/api/sites/{id}/rollback` | POST | Revert to previous deployment | Medium |
| `/api/sites/{id}/files/{path}` | GET | Download/view single file | Low |
| `/api/sites/{id}/files/{path}` | PUT | Upload/update single file | Medium |
| `/api/sites/{id}/files/{path}` | DELETE | Delete single file | Low |
| `/api/audit` | GET | View audit log entries | Low |
| `/api/system/backup` | POST | Trigger database backup | Medium |
| `/api/kv/{site_id}` | GET | View KV store for a site (serverless debug) | Low |
| `/api/kv/{site_id}/{key}` | DELETE | Clear KV entry | Low |

**Rationale:** File-level operations save users from redeploying entire sites for typo fixes. Audit visibility builds trust. Backup is essential for data safety.

---

### Priority 3: Nice to Have (Could Have)

| Endpoint | Method | Purpose | Complexity |
|----------|--------|---------|------------|
| `/api/sites/{id}/logs/stream` | GET (SSE) | Real-time log streaming | High |
| `/api/events/export` | GET | Export analytics as CSV/JSON | Medium |
| `/api/system/vacuum` | POST | Trigger DB vacuum (reclaim space) | Low |
| `/api/system/recalibrate` | POST | Re-probe hardware, update thresholds | Medium |
| `/api/certificates` | GET | View Let's Encrypt cert status | Low |
| `/api/notifications` | GET/POST | Manage ntfy notification rules | Medium |

**Rationale:** These enhance operational excellence but aren't blockers for v1 frontend.

---

### Priority 4: Future Considerations

| Endpoint | Method | Purpose | Notes |
|----------|--------|---------|-------|
| `/api/users` | CRUD | Multi-user support | Major feature |
| `/api/sites/{id}/domains` | CRUD | Custom domain mapping | Requires DNS verification |
| `/api/sites/{id}/schedule` | CRUD | Cron-like scheduled tasks | Serverless enhancement |

---

## Part 2: Security & Stability Improvements

### 2.1 Authentication Hardening

**Current State:**
- Session-based auth via cookies (dashboard)
- Bearer tokens for API (CLI/deploy)
- Rate limiting exists but not exposed

**Recommendations:**

| Issue | Recommendation | Priority |
|-------|----------------|----------|
| No token expiry visibility | Add `expires_at` to `/api/keys` response | High |
| No token rotation | Add `POST /api/keys/{id}/rotate` | Medium |
| Rate limit invisible | Add `X-RateLimit-*` headers to responses | High |
| No session list | Add `GET /api/sessions` to see active logins | Medium |
| No forced logout | Add `DELETE /api/sessions/{id}` | Medium |

**New Endpoint:**
```
GET /api/auth/limits
Response: {
  "login_attempts_remaining": 3,
  "lockout_expires": "2025-12-08T12:00:00Z",
  "rate_limit_remaining": 95,
  "rate_limit_reset": 1702036800
}
```

---

### 2.2 Input Validation Gaps

**Current Issues Found:**

| Handler | Issue | Fix |
|---------|-------|-----|
| `DeployHandler` | No file size limit in spec (only middleware) | Document 10MB default, make configurable |
| `EnvVarsHandler` | Value length unbounded | Add 4KB limit |
| `RedirectsHandler` | Destination URL not validated | Validate URL format, block `javascript:` |
| `WebhooksHandler` | Endpoint path not sanitized | Alphanumeric + hyphens only |

---

### 2.3 Stability Improvements

**Aligned with Cockpit Architecture:**

| Feature | Endpoint | Purpose |
|---------|----------|---------|
| Maintenance Mode | `GET/POST /api/system/maintenance` | Toggle 503 for user sites |
| Resource Limits | `GET /api/system/limits` | Show current thresholds |
| Cache Stats | `GET /api/system/cache` | VFS cache hit/miss, size |
| DB Stats | `GET /api/system/db` | SQLite page count, WAL size |

**Proposed Response for `/api/system/health`:**
```json
{
  "status": "healthy",
  "uptime_seconds": 86400,
  "version": "0.7.1",
  "memory": {
    "used_mb": 45,
    "limit_mb": 200,
    "vfs_cache_mb": 12
  },
  "database": {
    "path": "./data.db",
    "size_mb": 156,
    "sites_count": 12,
    "files_count": 847
  },
  "runtime": {
    "active_js_workers": 2,
    "max_js_workers": 4
  },
  "mode": "normal"  // or "maintenance"
}
```

---

## Part 3: API Design Critique & Redesign

### 3.1 Inconsistencies to Fix

| Current | Problem | Proposed |
|---------|---------|----------|
| `DELETE /api/sites?site_id=x` | Query param for identifier | `DELETE /api/sites/{id}` |
| `DELETE /api/keys?id=x` | Query param for identifier | `DELETE /api/keys/{id}` |
| `DELETE /api/envvars?id=x` | Query param for identifier | `DELETE /api/envvars/{id}` |
| `GET /api/logs?site_id=x` | Inconsistent with sites | `GET /api/sites/{id}/logs` |
| `GET /api/envvars?site_id=x` | Should be nested | `GET /api/sites/{id}/envvars` |
| `GET /api/deployments` | Returns all, no filter | `GET /api/sites/{id}/deployments` |

**Pattern to Adopt:**
```
/api/{resource}           → Collection operations
/api/{resource}/{id}      → Single resource operations
/api/{resource}/{id}/{sub} → Nested resource operations
```

---

### 3.2 Response Envelope Inconsistency

**Current (Mixed):**
```json
// Some endpoints:
{"success": true, "sites": [...]}

// Other endpoints:
[{...}, {...}]  // Raw array
```

**Proposed Standard:**
```json
{
  "data": [...],           // Always present
  "meta": {                // Optional, for pagination
    "total": 100,
    "limit": 50,
    "offset": 0
  },
  "error": null            // null on success, object on error
}
```

---

### 3.3 Error Response Standardization

**Current:** Mix of plain text and JSON errors

**Proposed:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Site name must be alphanumeric",
    "field": "site_name",
    "request_id": "req_abc123"
  }
}
```

**Error Codes:**
- `AUTH_REQUIRED` - Not logged in
- `AUTH_INVALID` - Bad credentials
- `RATE_LIMITED` - Too many requests
- `VALIDATION_ERROR` - Invalid input
- `NOT_FOUND` - Resource doesn't exist
- `CONFLICT` - Duplicate/conflict
- `INTERNAL_ERROR` - Server error

---

### 3.4 Recommended URL Structure (Full)

```
Authentication:
  POST   /api/auth/login
  POST   /api/auth/logout
  GET    /api/auth/me
  GET    /api/auth/sessions
  DELETE /api/auth/sessions/{id}

Sites (Hosting):
  GET    /api/sites
  POST   /api/sites                    (via deploy)
  GET    /api/sites/{id}
  DELETE /api/sites/{id}
  GET    /api/sites/{id}/files
  GET    /api/sites/{id}/files/{path}
  PUT    /api/sites/{id}/files/{path}
  DELETE /api/sites/{id}/files/{path}
  GET    /api/sites/{id}/logs
  GET    /api/sites/{id}/envvars
  POST   /api/sites/{id}/envvars
  DELETE /api/sites/{id}/envvars/{name}
  GET    /api/sites/{id}/deployments
  POST   /api/sites/{id}/rollback

Deploy:
  POST   /api/deploy                   (multipart)

API Keys:
  GET    /api/keys
  POST   /api/keys
  DELETE /api/keys/{id}
  POST   /api/keys/{id}/rotate

Analytics:
  GET    /api/stats
  GET    /api/events
  GET    /api/events/export
  GET    /api/domains
  GET    /api/tags

Tracking Config:
  GET    /api/redirects
  POST   /api/redirects
  DELETE /api/redirects/{id}
  GET    /api/webhooks
  POST   /api/webhooks
  PUT    /api/webhooks/{id}
  DELETE /api/webhooks/{id}

System:
  GET    /api/system/health
  GET    /api/system/config
  GET    /api/system/limits
  GET    /api/system/cache
  GET    /api/system/db
  POST   /api/system/backup
  POST   /api/system/vacuum
  POST   /api/system/recalibrate
  GET    /api/system/maintenance
  POST   /api/system/maintenance

Observability:
  GET    /api/audit
  GET    /api/certificates
  GET    /api/kv/{site_id}
  DELETE /api/kv/{site_id}/{key}

Version:
  GET    /api/version
```

---

## Part 4: Migration Strategy

### Phase 1: Non-Breaking Additions
Add new endpoints without changing existing ones:
- `/api/sites/{id}` (new)
- `/api/system/health` (new)
- `/api/version` (new)
- DELETE for redirects/webhooks (new)

### Phase 2: Deprecation Notices
Add `X-Deprecated` header to old patterns:
- `DELETE /api/sites?site_id=x` → Use `DELETE /api/sites/{id}`
- `GET /api/logs?site_id=x` → Use `GET /api/sites/{id}/logs`

### Phase 3: Frontend Rebuild
Build new frontend against new URL patterns.
Keep old endpoints working for CLI compatibility.

### Phase 4: Cleanup (v0.9.0)
Remove deprecated endpoints.
Update CLI to use new patterns.

---

## Appendix: Database Tables Not Exposed

| Table | Has API? | Recommendation |
|-------|----------|----------------|
| `events` | ✓ | - |
| `redirects` | ✓ | Add DELETE |
| `webhooks` | ✓ | Add DELETE/PUT |
| `files` | ✗ | Add file browser |
| `deployments` | ✓ | Move under sites |
| `api_keys` | ✓ | Add rotation |
| `kv_store` | ✗ | Add debug viewer |
| `env_vars` | ✓ | Restructure URL |
| `site_logs` | ✓ | Restructure URL |
| `configurations` | ✓ | via /api/config |
| `certificates` | ✗ | Add status view |
| `notifications` | ✗ | Add management |

---

## Summary

| Category | Count | Effort |
|----------|-------|--------|
| New Critical Endpoints | 7 | ~2 days |
| New Important Endpoints | 8 | ~3 days |
| URL Restructuring | 6 endpoints | ~1 day |
| Response Standardization | All endpoints | ~2 days |
| Security Additions | 5 endpoints | ~2 days |

**Recommended Approach:**
1. Implement Phase 1 (new endpoints) first
2. Build frontend against new patterns
3. Maintain backward compatibility for CLI
4. Deprecate old patterns in v0.8.0
5. Remove in v0.9.0
