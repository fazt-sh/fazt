# Admin API Specification (v0.7.1 Target)

**Date**: December 8, 2025
**Version**: v0.7.1
**Status**: Approved Target Design

This document outlines the RESTful API endpoints for the Fazt Admin Dashboard. 
**Design Goal**: Consistency, Completeness, Observability.

## Response Standard
All JSON responses follow this envelope:
```json
{
  "data": [ ... ] or { ... },
  "meta": { "total": 100, "limit": 50, "offset": 0 }, // Optional
  "error": null // or { "code": "ERR", "message": "..." }
}
```

## 1. Authentication
*Session-based for Dashboard, Bearer Token for CLI.*

| Method | Endpoint | Purpose |
|:---|:---|:---|
| `POST` | `/api/auth/login` | Login (Returns Session Cookie) |
| `POST` | `/api/auth/logout` | Destroy Session |
| `GET` | `/api/auth/me` | Current User Profile |
| `GET` | `/api/auth/limits` | Rate limits & Lockout status |

## 2. Hosting & Sites
*Primary Resource: Sites (Subdomains)*

| Method | Endpoint | Purpose |
|:---|:---|:---|
| `GET` | `/api/sites` | List all sites (Summary) |
| `POST` | `/api/sites` | Create/Deploy Site |
| `GET` | `/api/sites/{id}` | Single Site Details |
| `DELETE` | `/api/sites/{id}` | Delete Site |
| **Files** | | |
| `GET` | `/api/sites/{id}/files` | List Files (Tree) |
| `GET` | `/api/sites/{id}/files/{path}` | Download File |
| `PUT` | `/api/sites/{id}/files/{path}` | Upload/Edit File |
| `DELETE` | `/api/sites/{id}/files/{path}` | Delete File |
| **Config** | | |
| `GET` | `/api/sites/{id}/envvars` | List Env Vars |
| `POST` | `/api/sites/{id}/envvars` | Set Env Var |
| `DELETE` | `/api/sites/{id}/envvars/{key}`| Remove Env Var |
| **Ops** | | |
| `GET` | `/api/sites/{id}/logs` | Runtime Logs |
| `GET` | `/api/sites/{id}/deployments` | History |
| `POST` | `/api/sites/{id}/rollback` | Revert Deployment |

## 3. Analytics & Events

| Method | Endpoint | Purpose |
|:---|:---|:---|
| `GET` | `/api/stats` | Global Overview Stats |
| `GET` | `/api/events` | Raw Event Log |
| `GET` | `/api/events/export` | CSV/JSON Export |
| `GET` | `/api/domains` | Active Custom Domains |

## 4. Traffic Configuration

| Method | Endpoint | Purpose |
|:---|:---|:---|
| `GET` | `/api/redirects` | List Redirects |
| `POST` | `/api/redirects` | Create Redirect |
| `DELETE` | `/api/redirects/{id}` | Delete Redirect |
| `GET` | `/api/webhooks` | List Webhooks |
| `POST` | `/api/webhooks` | Create Webhook |
| `PUT` | `/api/webhooks/{id}` | Update Webhook |
| `DELETE` | `/api/webhooks/{id}` | Delete Webhook |

## 5. System & Observability
*Critical for Safeguards*

| Method | Endpoint | Purpose |
|:---|:---|:---|
| `GET` | `/api/system/health` | Status, Uptime, Version |
| `GET` | `/api/system/limits` | Resource Thresholds |
| `GET` | `/api/system/cache` | VFS Cache Stats (Hits/Size) |
| `GET` | `/api/system/db` | SQLite Stats (Size/WAL) |
| `GET` | `/api/system/config` | Server Config (Read-Only) |
| `POST` | `/api/system/maintenance` | Toggle Maintenance Mode |
| `POST` | `/api/system/vacuum` | Trigger DB Vacuum |
| `POST` | `/api/system/backup` | Trigger Snapshot |

## 6. Access Control (API Keys)

| Method | Endpoint | Purpose |
|:---|:---|:---|
| `GET` | `/api/keys` | List Deployment Keys |
| `POST` | `/api/keys` | Generate New Key |
| `DELETE` | `/api/keys/{id}` | Revoke Key |
| `POST` | `/api/keys/{id}/rotate` | Rotate Key Secret |

---

## Migration Plan

1.  **Phase 1 (Non-Breaking)**: Implement new `api/system/*` and `/api/sites/{id}` endpoints.
2.  **Phase 2 (Frontend)**: Rebuild Admin SPA to use new endpoints.
3.  **Phase 3 (Legacy)**: Maintain support for CLI-used endpoints (`/api/deploy`, `/api/logs?site_id=...`).
4.  **Phase 4 (Cleanup)**: Remove deprecated endpoints in v0.9.0.