# Admin API Specification (v0.7.0)

**Date**: December 08, 2025
**Version**: v0.7.0
**Status**: Draft

This document outlines the API endpoints required for the Fazt Admin Dashboard (Single Page Application). The dashboard allows users to manage sites, view analytics, and configure the system.

## Authentication
The Admin Dashboard uses **Session-based Authentication** via HTTP Cookies.
- **Login**: `POST /api/login`
- **Logout**: `POST /api/logout`
- **Session Check**: `GET /api/auth/status` (or `GET /api/user/me`)

*Note: Deployment tools (CLI) use Bearer Tokens, but the Admin SPA primarily uses Cookies.*

---

## 1. Authentication & User

### Login
`POST /api/login`
Authenticates the user and sets a session cookie.
- **Request**:
  ```json
  {
    "username": "admin",
    "password": "secret_password"
  }
  ```
- **Response** (200 OK):
  ```json
  {
    "success": true,
    "user": { "username": "admin", "role": "admin" }
  }
  ```

### Logout
`POST /api/logout`
Destroys the current session.
- **Response** (200 OK): `{"success": true}`

### Get Current User
`GET /api/user/me`
Returns the currently authenticated user's profile.
- **Response** (200 OK):
  ```json
  {
    "username": "admin",
    "authenticated": true
  }
  ```
- **Response** (401 Unauthorized): `{"authenticated": false}`

---

## 2. Dashboard Analytics (Home)

### Global Stats
`GET /api/stats`
Returns high-level metrics for the dashboard overview.
- **Query Params**: `range` (24h, 7d, 30d)
- **Response**:
  ```json
  {
    "total_visits": 1250,
    "unique_visitors": 800,
    "active_sites": 5,
    "total_events": 5000,
    "recent_activity": [...]
  }
  ```

### Events Log
`GET /api/events`
Returns raw event logs for the activity feed.
- **Query Params**: `limit` (default 50), `offset`
- **Response**:
  ```json
  {
    "events": [
      {
        "id": 1,
        "type": "pageview",
        "domain": "blog",
        "path": "/hello",
        "timestamp": "2023-10-27T10:00:00Z"
      }
    ]
  }
  ```

---

## 3. Hosting & Sites

### List Sites
`GET /api/sites`
Returns a summary of all deployed sites.
- **Response**:
  ```json
  {
    "success": true,
    "sites": [
      {
        "Name": "blog",
        "FileCount": 12,
        "SizeBytes": 102400,
        "ModTime": "2023-10-27T10:00:00Z",
        "Url": "https://blog.fazt.sh"
      }
    ]
  }
  ```

### Delete Site
`DELETE /api/sites?site_id={subdomain}`
Permanently deletes a site and its files.
- **Response**: `{"success": true}`

### Get Site Deployments
`GET /api/deployments?site_id={subdomain}`
Returns history of deployments for a specific site.
- **Response**:
  ```json
  {
    "deployments": [
      {
        "id": "dep_123",
        "created_at": "...",
        "file_count": 5
      }
    ]
  }
  ```

### Get Site Logs
`GET /api/logs?site_id={subdomain}`
Returns console/runtime logs for a specific site.
- **Query Params**: `limit` (default 100)
- **Response**:
  ```json
  {
    "success": true,
    "logs": [
      {
        "level": "INFO",
        "message": "Function executed",
        "created_at": "..."
      }
    ]
  }
  ```

---

## 4. Configuration & Secrets

### System Config (Read-Only)
`GET /api/config`
Returns sanitized server configuration (version, domain, environment).
- **Response**:
  ```json
  {
    "version": "v0.7.0",
    "domain": "fazt.sh",
    "environment": "production",
    "features": { "https": true, "registration": false }
  }
  ```

### API Keys (Deployment Tokens)
`GET /api/keys`
Lists active API keys.
- **Response**:
  ```json
  {
    "keys": [
      { "name": "deployment-token", "prefix": "fz_...", "created_at": "..." }
    ]
  }
  ```

`POST /api/keys`
Generates a new API key.
- **Request**: `{"name": "CI Token"}`
- **Response**: `{"token": "fz_full_token_string"}`

`DELETE /api/keys?name={name}`
Revokes an API key.

### Environment Variables
`GET /api/envvars?site_id={subdomain}`
Lists environment variables for a site (values masked).

`POST /api/envvars`
Sets a variable for a site.
- **Request**: `{"site_id": "blog", "key": "API_URL", "value": "..."}`

---

## 5. Tracking & Webhooks

### List Redirects
`GET /api/redirects`
Returns all configured shortlinks (`/r/xyz`).

### List Webhooks
`GET /api/webhooks`
Returns configured inbound webhooks.

---

## ⚠️ Missing / Gap Analysis

The current API is sufficient for **Reading** state but lacks crucial **Management** features for a full Admin Panel:

1.  **User Management**:
    - `POST /api/users` (Create User) - *Missing*
    - `PUT /api/users/{id}` (Update Password) - *Missing*
    - *Impact*: Currently limited to single-user (admin) defined at init.

2.  **File Manager**:
    - `GET /api/files?site_id={id}&path={path}` - *Missing*
    - `PUT /api/files` (Upload single file) - *Missing*
    - `DELETE /api/files` - *Missing*
    - *Impact*: Admin cannot edit/fix site files directly; must re-deploy via CLI.

3.  **System Health**:
    - `GET /api/system/health` (CPU/RAM usage) - *Existing /health is too basic*
    - *Impact*: No visibility into resource usage (critical for v0.7.0 safeguards).

4.  **Database Ops**:
    - `POST /api/system/backup` - *Missing*
    - `POST /api/system/vacuum` - *Missing*

## Recommended Sidebar Structure

1.  **Overview** (Stats, Recent Activity)
2.  **Sites** (List, Details, Logs, Env Vars)
3.  **Deployments** (History, API Keys)
4.  **Tracking** (Redirects, Webhooks, Pixels)
5.  **Settings** (User Profile, System Config)
