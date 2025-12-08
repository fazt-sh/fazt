# Admin API Specification

> **⚠️ IMPLEMENTATION NOTICE**
> This spec describes the *intended* design. For the **actual, current behavior** of the API (including inconsistencies), refer to **[request-response.md](./request-response.md)** which contains real captured payloads. Use that file as the "Ground Truth" for client implementation until this spec and the code are reconciled.

**Date**: December 8, 2025
**Version**: v0.7.2
**Status**: Current Implementation

This document outlines the RESTful API endpoints for the Fazt Admin Dashboard.
**Design Goal**: Consistency, Completeness, Observability.

## Response Standard
Most JSON responses follow a simple structure:
```json
{
  "success": true,
  "data": { ... },
  "message": "optional message"
}
```

New endpoints using the standardized API helper follow:
```json
{
  "data": [ ... ] or { ... },
  "meta": { "total": 100, "limit": 50, "offset": 0 }, // Optional
  "error": null // or { "code": "ERR", "message": "..." }
}
```

## 1. Authentication
*Session-based for Dashboard, Bearer Token for CLI.*

| Method | Endpoint | Purpose | Notes |
|:---|:---|:---|:---|
| `POST` | `/api/login` | Login (Returns Session Cookie) | Body: `{username, password, remember_me}` |
| `POST` | `/api/logout` | Destroy Session | Clears session cookie |
| `GET` | `/api/auth/status` | Current auth status | Returns `{authenticated, username, expiresAt}` |
| `GET` | `/api/user/me` | Current User Profile | Returns `{username, version}` |

## 2. Hosting & Sites
*Primary Resource: Sites (Subdomains)*

| Method | Endpoint | Purpose | Notes |
|:---|:---|:---|:---|
| `GET` | `/api/sites` | List all sites | Returns array of `{Name, FileCount, SizeBytes, ModTime}` |
| `POST` | `/api/deploy` | Deploy Site via ZIP | Requires Bearer token, multipart with `site_name` and `file` |
| `GET` | `/api/sites/{id}` | Single Site Details | Returns site info |
| `DELETE` | `/api/sites?site_id={id}` | Delete Site | Query param: `site_id` |
| **Files** | | | |
| `GET` | `/api/sites/{id}/files` | List Files (Tree) | Returns VFS file listing for site |
| `GET` | `/api/sites/{id}/files/{path...}` | Download File | Wildcard path parameter captures full file path |
| **Config** | | | |
| `GET` | `/api/envvars?site_id={id}` | List Env Vars | Query param: `site_id`, returns `{id, name}` (values hidden) |
| `POST` | `/api/envvars` | Set Env Var | Body: `{site_id, name, value}`. Upserts env var |
| `DELETE` | `/api/envvars?id={id}` | Remove Env Var | Query param: `id` |
| **Ops** | | | |
| `GET` | `/api/logs?site_id={id}&limit={n}` | Site Runtime Logs | Query params: `site_id` (required), `limit` (default 50, max 1000) |
| `GET` | `/api/deployments` | Deployment History | Returns last 50 deployments across all sites |

## 3. Analytics & Events

| Method | Endpoint | Purpose | Notes |
|:---|:---|:---|:---|
| `GET` | `/api/stats` | Global Overview Stats | Returns events counts, top domains, top tags, timeline |
| `GET` | `/api/events` | Raw Event Log | Query params: `domain`, `tags`, `source_type`, `limit` (default 50), `offset` (default 0) |
| `GET` | `/api/domains` | Active Custom Domains | Returns list of domains with event counts |
| `GET` | `/api/tags` | Tags with usage counts | Returns aggregated tag statistics |

### Tracking Endpoints (Public)
| Method | Endpoint | Purpose | Notes |
|:---|:---|:---|:---|
| `POST` | `/track` | Track analytics event | Body: `{domain?, hostname?, event_type?, path?, referrer?, tags[]}` |
| `GET` | `/pixel.gif` | Pixel tracking | Query params for event data |
| `GET` | `/r/{slug}` | Redirect tracking | Redirects to destination, tracks click |
| `POST` | `/webhook/{endpoint}` | Webhook receiver | Requires webhook to be configured, validates HMAC signature |

## 4. Traffic Configuration

| Method | Endpoint | Purpose | Notes |
|:---|:---|:---|:---|
| `GET` | `/api/redirects` | List Redirects | Returns `{id, slug, destination, tags[], click_count, created_at}` |
| `POST` | `/api/redirects` | Create Redirect | Body: `{slug, destination, tags[]}` |
| `DELETE` | `/api/redirects/{id}` | Delete Redirect | Path param: `id` |
| `GET` | `/api/webhooks` | List Webhooks | Returns `{id, name, endpoint, has_secret, is_active, created_at}` |
| `POST` | `/api/webhooks` | Create Webhook | Body: `{name, endpoint, secret?}` |
| `PUT` | `/api/webhooks/{id}` | Update Webhook | Body: `{name?, endpoint?, secret?, is_active?}`, partial update |
| `DELETE` | `/api/webhooks/{id}` | Delete Webhook | Path param: `id` |

## 5. System & Observability

| Method | Endpoint | Purpose | Notes |
|:---|:---|:---|:---|
| `GET` | `/api/system/health` | System health & metrics | Returns `{status, uptime_seconds, version, mode, memory, database, runtime}` |
| `GET` | `/api/system/limits` | Resource Thresholds | Returns system resource limits |
| `GET` | `/api/system/cache` | VFS Cache Stats | Returns VFS cache statistics |
| `GET` | `/api/system/db` | SQLite Stats | Returns database connection stats |
| `GET` | `/api/system/config` | Server Config (Sanitized) | Returns `{version, domain, env, https, ntfy}` |
| `GET` | `/api/config` | Alias for system/config | Same as above |
| `GET` | `/health` | Simple health check | Returns "OK" if database is healthy |

## 6. Access Control (API Keys)

| Method | Endpoint | Purpose | Notes |
|:---|:---|:---|:---|
| `GET` | `/api/keys` | List Deployment Keys | Returns list of API keys (tokens hidden) |
| `POST` | `/api/keys` | Generate New Key | Body: `{name, scopes?}`. Returns token (shown once only!) |
| `DELETE` | `/api/keys?id={id}` | Revoke Key | Query param: `id` |

---

## Implementation Notes

### Authentication
- Session-based authentication for Dashboard (cookie: `session_id`)
- Bearer token authentication for CLI/API clients (`Authorization: Bearer <token>`)
- Rate limiting: 5 failed login attempts trigger 15-minute lockout per IP
- Rate limiting: 5 deploys per minute per IP

### Environment Variables
- Must be uppercase letters, numbers, and underscores
- Must start with a letter
- Max 128 characters
- Reserved system variables cannot be overridden: `PATH`, `HOME`, `USER`, `NODE_OPTIONS`, etc.

### File Uploads
- Deploy: Max 100MB ZIP file
- Tracking: Max 10KB body size
- Webhook: Max 10KB body size

### Pagination
- Events endpoint: Default `limit=50`, supports `offset` parameter
- Logs endpoint: Default `limit=50`, max `limit=1000`
- Deployments: Fixed at last 50 entries