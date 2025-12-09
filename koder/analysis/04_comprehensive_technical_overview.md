# Fazt.sh: Comprehensive Technical Overview

**Date**: December 9, 2025
**Version**: v0.8.0-dev (API Standardization Complete)
**Target Audience**: Senior Engineers / Architectural Reviewers

---

## 1. Executive Summary

**Fazt.sh** is a "Personal PaaS" (Platform as a Service) designed around a radical philosophy of simplicity and portability. Unlike traditional hosting platforms that rely on container orchestration (Kubernetes, Docker) or complex microservices, Fazt collapses the entire stack into a **Single Binary + Single Database** architecture.

It allows a user to deploy static sites and serverless functions to a single low-cost VPS (e.g., $6/mo DigitalOcean Droplet) with zero external dependencies.

---

## 2. Core Philosophy & Architectural Constraints

### 2.1 The "Cartridge" Concept
The architecture mimics a game console cartridge:
*   **The Console**: The `fazt` binary (stateless logic, runtime).
*   **The Cartridge**: The `data.db` SQLite file (state, filesystem, config).
*   **Result**: To migrate a server, you simply copy `data.db` to a new machine.

### 2.2 Zero External Dependencies
*   **No CGO**: The project is built with `CGO_ENABLED=0` to ensure static linking and cross-platform compatibility (Linux/amd64, Linux/arm64, macOS).
*   **No Runtime Dependencies**: It does not require Node.js, Python, Nginx, or Docker to be installed on the host.
*   **Pure Go**: All functionality, including the JS runtime and SQLite driver, is implemented in pure Go.

---

## 3. System Architecture

### 3.1 High-Level Blocks
```mermaid
graph TD
    User[User / Internet] --> |HTTPS :443| Server[Fazt Server (Go)]

    subgraph "Fazt Server Process"
        Router[Host Router]
        Dashboard[Admin Dashboard]
        Hosting[VFS Site Handler]
        Runtime[JS Runtime (Goja)]
        CertMagic[Auto-TLS (CertMagic)]
    end

    subgraph "Storage (SQLite)"
        TableFiles[table: files (VFS)]
        TableKV[table: kv_store]
        TableConfig[table: config]
        TableLogs[table: site_logs]
    end

    User --> Router
    Router --> |admin.domain| Dashboard
    Router --> |*.domain| Hosting

    Hosting --> |Static File| TableFiles
    Hosting --> |main.js| Runtime

    Runtime --> |db.get/set| TableKV
    Runtime --> |fetch| ExternalAPI[External APIs]

    Dashboard --> TableConfig
    CertMagic --> |Cert Storage| TableFiles
```

### 3.2 Key Components

#### A. Virtual Filesystem (VFS)
*   **Implementation**: Instead of storing user sites on the host's ext4/xfs filesystem, Fazt stores files as BLOBs in the `files` table in SQLite.
*   **Schema**: `files(site_id, path, content, size_bytes, hash, mime_type, updated_at)`.
*   **Performance**: Uses an in-memory LRU cache (`internal/hosting/vfs.go`) to avoid hitting the DB for hot assets (CSS/JS/HTML).
*   **Rationale**: Enables atomic backups (snapshotting the DB snapshots the filesystem) and prevents inode exhaustion on small VPSs.

#### B. Serverless Runtime
*   **Engine**: `dop251/goja` (ECMAScript 5.1(+) implementation in pure Go).
*   **Trigger**: If a site contains `main.js`, it is executed for every request.
*   **Capabilities**:
    *   `req` / `res`: Express-like API.
    *   `db`: Key-Value store access (`kv_store` table).
    *   `fetch`: HTTP client with SSRF protection.
    *   `socket`: WebSocket broadcasting.
*   **Constraints**: 100ms execution timeout, 1MB body limit.

#### C. Routing & Domain Management
*   **Library**: `caddyserver/certmagic`.
*   **Function**: Automatically manages Let's Encrypt / ZeroSSL certificates.
*   **Storage**: Certificates are stored in the SQLite DB (custom implementation of CertMagic storage interface).
*   **Logic**:
    *   `admin.<domain>` → Admin Dashboard (Authentication required).
    *   `root.<domain>` / `<domain>` → System Landing Page.
    *   `*.<domain>` → User Sites (VFS).

#### D. Database & State
*   **Library**: `modernc.org/sqlite` (CGO-free SQLite).
*   **Migrations**: SQL files embedded in binary, applied on startup (`internal/database/migrations/`).
*   **Tables**:
    *   `files`: Hosting content (VFS).
    *   `kv_store`: User-land persistence for serverless functions.
    *   `events`: Analytics (page views, redirect clicks, pixel tracking).
    *   `site_logs`: `console.log` capture from JS runtime.
    *   `config`: Server configuration (Port, Auth, Env).
    *   `api_keys`: API key management for deployments.
    *   `env_vars`: Environment variables per site.
    *   `redirects`: URL shortener/redirect service.
    *   `webhooks`: Webhook endpoints for integrations.
    *   `deployments`: Deployment history tracking.

#### E. Admin API (NEW - v0.8.0)
*   **Standardized Format**: All endpoints return `{"data": ...}` for success, `{"error": {...}}` for errors.
*   **Response Helpers**: `internal/api/response.go` provides consistent wrappers.
*   **Error Codes**: Machine-readable codes (e.g., `VALIDATION_FAILED`, `SESSION_EXPIRED`).
*   **Endpoints (~30 total)**:
    *   **Auth**: login, logout, user/me, auth/status
    *   **System**: health, config, limits
    *   **Hosting**: sites CRUD, deploy, env vars, API keys
    *   **Analytics**: events, stats, domains, tags
    *   **Redirects**: CRUD operations
    *   **Webhooks**: CRUD operations
    *   **Tracking**: pixel, redirect tracking
    *   **Logs**: site logs, deployment history

---

## 4. Technical Specifications

| Component | Technology | Reasoning |
| :--- | :--- | :--- |
| **Language** | Go 1.24+ | Concurrency, Static Binary, Tooling. |
| **Database** | SQLite (modernc) | Zero config, file-based, CGO-free. |
| **HTTP Server** | `net/http` | Standard library is robust enough. |
| **HTTPS** | `certmagic` | Best-in-class ACME implementation. |
| **JS Runtime** | `goja` | Pure Go, safe sandboxing (unlike V8/cgo). |
| **WebSockets** | `gorilla/websocket` | Standard for Go WS. |
| **Password Hashing** | `bcrypt` | Standard security practice. |
| **API Responses** | Custom (`internal/api`) | Standardized envelope format. |

---

## 5. Deployment & Operations

### 5.1 Installation
A single shell script (`install.sh`) handles:
1.  Downloading the binary.
2.  Creating a system user (`fazt`).
3.  Setting up `systemd` service.
4.  Configuring `setcap` for binding to port 443 without root.

### 5.2 Deployment Protocol (Client → Server)
1.  **Client**: Zips the target directory.
2.  **Client**: `POST /api/deploy` with Bearer token.
3.  **Server**: Unzips in memory.
4.  **Server**: Calculates SHA256 hashes of files.
5.  **Server**: Upserts into `files` table (deduplication at file path level).
6.  **Server**: Invalidates VFS cache for that site.

---

## 6. API Design (v0.8.0+)

### 6.1 Response Format

**Success:**
```json
{"data": {...}}
{"data": [...], "meta": {"total": 100, "limit": 20, "offset": 0}}
```

**Error:**
```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Site name is required",
    "details": {"field": "site_name", "constraint": "required"}
  }
}
```

### 6.2 Design Principles
*   **HTTP-centric**: Status code is the source of truth (200 = success, 4xx/5xx = error).
*   **Clean separation**: Success has `data`, errors have `error`, never both.
*   **Minimal overhead**: ~18 bytes saved per response vs mixed envelope.
*   **SPA-friendly**: Single fetch wrapper works for all endpoints.
*   **Extensible**: `meta` for pagination, `details` for field errors, `code` for i18n.

### 6.3 Error Code Registry
Standardized machine-readable codes:
*   `BAD_REQUEST`, `VALIDATION_FAILED`, `INVALID_JSON`, `MISSING_FIELD`
*   `UNAUTHORIZED`, `INVALID_CREDENTIALS`, `SESSION_EXPIRED`, `INVALID_API_KEY`
*   `FORBIDDEN`, `NOT_FOUND`, `SITE_NOT_FOUND`, `REDIRECT_NOT_FOUND`, etc.
*   `CONFLICT`, `PAYLOAD_TOO_LARGE`, `RATE_LIMIT_EXCEEDED`
*   `INTERNAL_ERROR`, `SERVICE_UNAVAILABLE`

**Full registry:** `koder/plans/11_api-standardization.md` section 3.2

---

## 7. Testing Infrastructure (NEW - v0.8.0)

### 7.1 Test Utilities
*   **Location**: `internal/handlers/testutil/helpers.go`
*   **Purpose**: Validate standardized API responses.
*   **Functions**:
    *   `CheckSuccess(t, rr, 200)` - Validates `{"data": ...}` responses
    *   `CheckError(t, rr, 400, "BAD_REQUEST")` - Validates error responses
    *   `CheckSuccessArray(t, rr, 200)` - For array data
    *   `WithSession(req, sessionID)` - Add session cookie
    *   `WithAuth(req, token)` - Add Bearer token
    *   `JSONRequest(method, path, body)` - Create JSON request

### 7.2 Shared Test Setup
*   **Location**: `internal/handlers/handlers_test.go`
*   **Functions**:
    *   `setupTestDB(t)` - In-memory SQLite with full schema
    *   `setupTestAuth(t)` - Session store + valid session
    *   `setupTestConfig(t)` - Test configuration
    *   `createTestSite()`, `createTestRedirect()`, etc. - Test data helpers

### 7.3 Test Coverage
*   **Auth handlers**: 11 comprehensive test cases (all passing)
*   **Pattern**: Success, validation error, auth failure, not found, server error
*   **Reference**: `internal/handlers/auth_test.go`

---

## 8. Current Limitations & Known Issues

1.  **No "Real" Multi-Tenancy**: Designed for a single "owner" hosting many sites.
2.  **JS Runtime Version**: ES5.1 (Goja) lacks modern JS features (async/await, ES6 modules) without transpilation.
3.  **Database Locking**: SQLite writer lock can bottleneck if high write concurrency (WAL mode enabled to mitigate).
4.  **Memory Usage**: VFS Cache and Goja VMs can consume RAM; strict limits are enforced but tuning is manual.
5.  **Admin UI**: Currently being rebuilt as vanilla SPA (see `koder/rough.md` for vision).

---

## 9. Recent Updates (v0.7.x → v0.8.0-dev)

### API Standardization (December 2025)
*   **Completed**: Full API response format standardization.
*   **Impact**: All 11 handler files migrated, ~30 endpoints standardized.
*   **Benefits**:
    *   Consistent response format across all endpoints
    *   SPA development simplified (single fetch wrapper)
    *   Better error handling (machine-readable codes)
    *   Field-level validation feedback
*   **Documentation**: `koder/plans/11_api-standardization.md`

### Test Infrastructure
*   **Added**: Comprehensive test utilities and shared setup
*   **Coverage**: High coverage on auth handlers (11 tests)
*   **Pattern**: Test-first methodology for new features

---

## 10. Future Directions (Roadmap Alignment)

### Immediate (v0.8.0)
*   **Admin SPA Rebuild**: Vanilla JS/PWA dashboard (in planning)
*   **Documentation**: Update API docs, generate OpenAPI spec

### Near-term (v0.9.0)
*   **JS Cron**: Scheduled serverless tasks
*   **Backup/Restore**: CLI commands to safely snapshot the running DB
*   **App Store**: Git-based "Click to Install" for standardized apps

### Long-term
*   **Multi-user Auth**: Proper user management (currently single admin)
*   **WebSocket API**: Real-time updates for dashboard
*   **Plugin System**: Extensibility for custom functionality

---

## 11. Key Files & Documentation

### Source Code
*   `cmd/server/main.go` - Entry point, CLI, router setup
*   `internal/api/response.go` - Standardized API response helpers
*   `internal/handlers/*.go` - HTTP handlers (all standardized)
*   `internal/hosting/vfs.go` - Virtual filesystem implementation
*   `internal/database/db.go` - Database initialization & migrations
*   `internal/auth/session.go` - Session management

### Documentation
*   `CLAUDE.md` - Assistant guide (this session's context)
*   `koder/start.md` - Bootstrap entry point
*   `koder/NEXT_SESSION.md` - Current status & next steps
*   `koder/plans/11_api-standardization.md` - API spec
*   `koder/rough.md` - Admin SPA vision
*   `koder/analysis/` - Technical analysis documents

### Testing
*   `internal/handlers/auth_test.go` - Reference test pattern
*   `internal/handlers/testutil/helpers.go` - Test utilities
*   `scripts/verify_api_migration.sh` - API migration verification

---

## 12. Development Workflow

### Starting a New Session
```bash
read and execute koder/start.md
```

This loads:
1. Current mission from `koder/NEXT_SESSION.md`
2. Relevant context files
3. Environment verification

### Common Commands
```bash
# Build
go build -o fazt ./cmd/server

# Test (all)
go test ./...

# Test (specific)
go test ./internal/handlers/auth_test.go -v

# Run (dev)
go run ./cmd/server server start --domain localhost --port 8080

# Verify API migration status
./scripts/verify_api_migration.sh
```

### Handler Development Pattern
1. Write tests first (copy from `auth_test.go`)
2. Implement handler using `api.*` helpers
3. Run tests until passing
4. Commit with descriptive message

---

**Last Updated**: December 9, 2025 (API Standardization Complete)
**Next Phase**: Admin SPA Rebuild (Planning)
