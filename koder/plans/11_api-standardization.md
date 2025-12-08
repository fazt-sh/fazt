# Plan 11: API Standardization for Admin Dashboard

**Date**: December 9, 2025
**Status**: 🟢 Ready for Implementation
**Target**: v0.8.0
**Methodology**: Test-First Development

---

## 1. Executive Summary

### The Problem
Current Admin API has fragmented response formats:
- 3 different success patterns (`{"data": ...}`, `{"success": true, ...}`, raw JSON)
- 2 different error patterns (`http.Error` vs `{"error": ...}` vs `{"data": null, "error": ...}`)
- Inconsistent DELETE patterns (query params vs path params)
- No standardized error codes
- Makes SPA development painful (can't write generic fetch wrappers)

### The Solution
Implement a **minimal, HTTP-centric, SPA-friendly** API standard:

**Success Response:**
```json
{
  "data": {...}         // Single resource or array
}
```

**Success with Metadata:**
```json
{
  "data": [...],
  "meta": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

**Error Response:**
```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Site name is required",
    "details": {
      "field": "site_name",
      "constraint": "required"
    }
  }
}
```

### Why This Design?

| Principle | Rationale |
|:---|:---|
| **HTTP Status = Truth** | `200` means success, `4xx/5xx` means error. No redundant `"success": bool` field. |
| **Separation of Concerns** | Success returns `data`, errors return `error`. Never both in one response. |
| **Minimal Overhead** | ~12 bytes avg (`{"data":`) vs ~35 bytes for `{"success":true,"data":}` |
| **SPA-Friendly** | `resp.ok ? resp.json().data : resp.json().error.message` |
| **Extensible** | `meta` for pagination, `details` for field errors, `code` for i18n |
| **HTTP-Semantic** | Aligns with REST, GraphQL error spec, Stripe API, GitHub API |

---

## 2. Design Principles

### 2.1 Core Tenets

1. **HTTP Status Codes are First-Class Citizens**
   - `200 OK`: Successful GET/PUT/PATCH
   - `201 Created`: Successful POST
   - `204 No Content`: Successful DELETE (optional body)
   - `400 Bad Request`: Client error (validation, malformed)
   - `401 Unauthorized`: Missing/invalid auth
   - `404 Not Found`: Resource doesn't exist
   - `429 Too Many Requests`: Rate limit
   - `500 Internal Server Error`: Server panic/DB error

2. **JSON Structure Reflects HTTP Semantics**
   - `2xx` responses → `{"data": ...}` (ALWAYS)
   - `4xx/5xx` responses → `{"error": ...}` (ALWAYS)
   - Never mix `data` and `error` in the same envelope

3. **Error Codes are Stable Contracts**
   - Machine-readable (e.g., `SITE_NOT_FOUND`)
   - Used for client-side logic (retry, redirect, display)
   - Human messages can change, codes cannot

4. **Field-Level Errors for Forms**
   - SPA forms need to highlight specific fields
   - Use `error.details.field` for this

### 2.2 Comparison with Current Implementation

| Aspect | Current (`api/response.go`) | Proposed |
|:---|:---|:---|
| Success | `{"data": ..., "error": null}` | `{"data": ...}` |
| Error | `{"data": null, "error": {...}}` | `{"error": {...}}` |
| Overhead | ~18 bytes (null fields) | ~0 bytes |
| Clarity | Mixed (why include null?) | Clear (mutually exclusive) |

---

## 3. API Specification

### 3.1 Response Envelope Structure

#### Success Response (Single Resource)
```typescript
interface SuccessResponse<T> {
  data: T;
  meta?: Meta;  // Optional, only for lists/pagination
}
```

**Example:**
```json
GET /api/sites/blog
200 OK
{
  "data": {
    "name": "blog",
    "path": "vfs://blog",
    "file_count": 42,
    "size_bytes": 1048576,
    "mod_time": "2025-12-09T12:00:00Z"
  }
}
```

#### Success Response (List with Pagination)
```json
GET /api/events?limit=50&offset=100
200 OK
{
  "data": [
    {"id": 1, "domain": "example.com", ...},
    {"id": 2, "domain": "blog.com", ...}
  ],
  "meta": {
    "total": 1000,
    "limit": 50,
    "offset": 100
  }
}
```

#### Error Response
```typescript
interface ErrorResponse {
  error: {
    code: string;           // Machine-readable (UPPERCASE_SNAKE_CASE)
    message: string;        // Human-readable
    details?: {             // Optional, for validation errors
      field?: string;       // Which field caused the error
      constraint?: string;  // What constraint failed
      [key: string]: any;   // Extensible
    };
  };
}
```

**Example (Validation Error):**
```json
POST /api/sites
400 Bad Request
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Site name must be 3-63 characters and contain only alphanumeric, hyphens",
    "details": {
      "field": "site_name",
      "constraint": "pattern",
      "pattern": "^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$"
    }
  }
}
```

**Example (Not Found):**
```json
GET /api/sites/nonexistent
404 Not Found
{
  "error": {
    "code": "SITE_NOT_FOUND",
    "message": "Site 'nonexistent' does not exist"
  }
}
```

**Example (Server Error):**
```json
POST /api/deploy
500 Internal Server Error
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Database connection failed",
    "details": {
      "request_id": "req_abc123"  // For support/debugging
    }
  }
}
```

### 3.2 Error Code Registry

Standardized error codes across all endpoints:

| HTTP | Code | When to Use | Example |
|:---|:---|:---|:---|
| 400 | `BAD_REQUEST` | Generic malformed request | Missing Content-Type |
| 400 | `VALIDATION_FAILED` | Field validation error | Invalid site name format |
| 400 | `INVALID_JSON` | JSON parse error | Malformed JSON body |
| 400 | `MISSING_FIELD` | Required field missing | `site_name` not provided |
| 400 | `INVALID_FILE` | File upload issue | Not a valid ZIP |
| 401 | `UNAUTHORIZED` | Missing/invalid auth | No session cookie |
| 401 | `INVALID_CREDENTIALS` | Login failed | Wrong password |
| 401 | `SESSION_EXPIRED` | Session timeout | Session > 24h old |
| 401 | `INVALID_API_KEY` | API key auth failed | Wrong Bearer token |
| 403 | `FORBIDDEN` | Authenticated but no permission | Read-only key trying to deploy |
| 404 | `NOT_FOUND` | Generic resource not found | Unknown route |
| 404 | `SITE_NOT_FOUND` | Specific: site missing | `GET /api/sites/xyz` |
| 404 | `REDIRECT_NOT_FOUND` | Specific: redirect missing | `GET /api/redirects/999` |
| 404 | `WEBHOOK_NOT_FOUND` | Specific: webhook missing | `GET /api/webhooks/999` |
| 409 | `CONFLICT` | Resource already exists | Site name taken |
| 413 | `PAYLOAD_TOO_LARGE` | Upload exceeds limit | ZIP > 100MB |
| 429 | `RATE_LIMIT_EXCEEDED` | Too many requests | > 5 deploys/min |
| 500 | `INTERNAL_ERROR` | Server panic/DB error | Uncaught exception |
| 503 | `SERVICE_UNAVAILABLE` | DB locked, maintenance | SQLite busy |

### 3.3 RESTful Route Patterns

All routes follow standard REST conventions:

| Action | Method | Pattern | Example |
|:---|:---|:---|:---|
| **List** | GET | `/api/:resource` | `GET /api/sites` |
| **Create** | POST | `/api/:resource` | `POST /api/sites` |
| **Get** | GET | `/api/:resource/:id` | `GET /api/sites/blog` |
| **Update** | PUT/PATCH | `/api/:resource/:id` | `PUT /api/sites/blog` |
| **Delete** | DELETE | `/api/:resource/:id` | `DELETE /api/sites/blog` |

#### Routes Requiring Refactoring:

| Current (Wrong) | Corrected (REST) | Status |
|:---|:---|:---|
| `DELETE /api/sites?site_id=X` | `DELETE /api/sites/X` | 🔴 Migrate |
| `DELETE /api/envvars?id=X` | `DELETE /api/envvars/X` | 🔴 Migrate |
| `DELETE /api/keys?id=X` | `DELETE /api/keys/X` | 🔴 Migrate |
| `GET /api/logs?site_id=X` | `GET /api/sites/X/logs` | 🟡 Consider (breaking) |
| `GET /api/envvars?site_id=X` | `GET /api/sites/X/envvars` | 🟡 Consider (cleaner nesting) |

**Note:** Nested resources (`/api/sites/X/logs`) are optional. If it adds clarity, use them. Otherwise, query params for filtering (`/api/logs?site_id=X`) are acceptable.

---

## 4. Testing Strategy (Test-First Development)

### 4.1 Philosophy

**Test-Driven Refactoring:**
1. Write integration tests for ALL endpoints (based on new spec)
2. Tests initially FAIL (current handlers don't match spec)
3. Refactor handlers until tests PASS
4. Regression protection: tests prevent backsliding

### 4.2 Test Coverage Requirements

**Every endpoint MUST have:**
1. ✅ **Success case** (200/201)
2. ✅ **Validation error** (400) - at least one field
3. ✅ **Auth failure** (401) - missing/invalid token
4. ✅ **Not found** (404) - resource doesn't exist
5. ✅ **Server error** (500) - DB failure simulation

**Optional but recommended:**
- Rate limiting (429)
- Conflict (409) for duplicate creation
- Payload too large (413) for uploads

### 4.3 Test Organization

```
internal/handlers/
├── handlers_test.go           # Shared test utilities
├── auth_test.go               # Auth endpoints tests
├── sites_test.go              # Sites CRUD tests
├── deploy_test.go             # Deployment tests
├── redirects_test.go          # Redirects CRUD tests
├── webhooks_test.go           # Webhooks CRUD tests
├── events_test.go             # Analytics tests
├── system_test.go             # Health/config tests
└── ...
```

### 4.4 Test Helper Functions

Create `internal/handlers/testutil/helpers.go`:

```go
package testutil

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

// Standard response checkers
func CheckSuccess(t *testing.T, resp *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
    if resp.Code != expectedStatus {
        t.Fatalf("Expected %d, got %d: %s", expectedStatus, resp.Code, resp.Body.String())
    }

    var result struct {
        Data interface{} `json:"data"`
        Meta interface{} `json:"meta,omitempty"`
    }

    if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
        t.Fatalf("Invalid JSON response: %v", err)
    }

    if result.Data == nil {
        t.Fatal("Expected 'data' field in success response")
    }

    return result
}

func CheckError(t *testing.T, resp *httptest.ResponseRecorder, expectedStatus int, expectedCode string) {
    if resp.Code != expectedStatus {
        t.Fatalf("Expected %d, got %d: %s", expectedStatus, resp.Code, resp.Body.String())
    }

    var result struct {
        Error struct {
            Code    string                 `json:"code"`
            Message string                 `json:"message"`
            Details map[string]interface{} `json:"details,omitempty"`
        } `json:"error"`
    }

    if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
        t.Fatalf("Invalid JSON response: %v", err)
    }

    if result.Error.Code != expectedCode {
        t.Fatalf("Expected error code '%s', got '%s'", expectedCode, result.Error.Code)
    }
}

// Auth helpers
func WithAuth(req *http.Request, token string) *http.Request {
    req.Header.Set("Authorization", "Bearer "+token)
    return req
}

func WithSession(req *http.Request, sessionID string) *http.Request {
    req.AddCookie(&http.Cookie{
        Name:  "fazt_session",
        Value: sessionID,
    })
    return req
}

// Request builders
func JSONRequest(method, path string, body interface{}) *http.Request {
    jsonBody, _ := json.Marshal(body)
    req := httptest.NewRequest(method, path, bytes.NewReader(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    return req
}
```

### 4.5 Example Test Cases

**Example: Sites CRUD Test**

```go
// internal/handlers/sites_test.go
package handlers

import (
    "testing"
    "net/http/httptest"
    "github.com/fazt-sh/fazt/internal/handlers/testutil"
)

func TestListSites_Success(t *testing.T) {
    // Setup
    setupTestDB(t)
    defer cleanupTestDB(t)

    // Create test site
    createTestSite(t, "blog")

    // Request
    req := testutil.JSONRequest("GET", "/api/sites", nil)
    req = testutil.WithSession(req, testSession.ID)

    rr := httptest.NewRecorder()
    http.HandlerFunc(ListSitesHandler).ServeHTTP(rr, req)

    // Assert
    result := testutil.CheckSuccess(t, rr, 200)

    sites, ok := result["data"].([]interface{})
    if !ok || len(sites) == 0 {
        t.Fatal("Expected non-empty sites array")
    }
}

func TestCreateSite_ValidationError(t *testing.T) {
    // Setup
    setupTestDB(t)
    defer cleanupTestDB(t)

    // Request with invalid site name
    req := testutil.JSONRequest("POST", "/api/sites", map[string]string{
        "name": "",  // Empty name should fail
    })
    req = testutil.WithSession(req, testSession.ID)

    rr := httptest.NewRecorder()
    http.HandlerFunc(CreateSiteHandler).ServeHTTP(rr, req)

    // Assert
    testutil.CheckError(t, rr, 400, "VALIDATION_FAILED")
}

func TestGetSite_NotFound(t *testing.T) {
    // Setup
    setupTestDB(t)
    defer cleanupTestDB(t)

    // Request non-existent site
    req := testutil.JSONRequest("GET", "/api/sites/nonexistent", nil)
    req = testutil.WithSession(req, testSession.ID)

    rr := httptest.NewRecorder()
    http.HandlerFunc(GetSiteHandler).ServeHTTP(rr, req)

    // Assert
    testutil.CheckError(t, rr, 404, "SITE_NOT_FOUND")
}

func TestDeleteSite_Unauthorized(t *testing.T) {
    // Setup
    setupTestDB(t)
    defer cleanupTestDB(t)
    createTestSite(t, "blog")

    // Request without auth
    req := testutil.JSONRequest("DELETE", "/api/sites/blog", nil)
    // No session cookie

    rr := httptest.NewRecorder()
    http.HandlerFunc(DeleteSiteHandler).ServeHTTP(rr, req)

    // Assert
    testutil.CheckError(t, rr, 401, "UNAUTHORIZED")
}
```

### 4.6 Test Execution

```bash
# Run all handler tests
go test ./internal/handlers/... -v

# Run specific test file
go test ./internal/handlers/sites_test.go -v

# Run with coverage
go test ./internal/handlers/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# Target: 80%+ coverage on handlers package
```

---

## 5. Implementation Plan

### Phase 1: Foundation (Test Infrastructure)
**Goal:** Build the testing framework before touching handlers

**Tasks:**
1. ✅ Create `internal/handlers/testutil/helpers.go`
   - `CheckSuccess()`, `CheckError()`, `WithAuth()`, `WithSession()`, `JSONRequest()`
2. ✅ Create `internal/handlers/handlers_test.go`
   - `setupTestDB()`, `cleanupTestDB()`, `createTestSite()`, shared fixtures
3. ✅ Write one complete test suite as template
   - Pick `sites_test.go` as the reference implementation
   - All 5 required cases (success, validation, auth, not found, server error)
4. ✅ Verify tests FAIL against current implementation
   - Confirms tests are testing the NEW spec, not current behavior

**Acceptance Criteria:**
- `go test ./internal/handlers/sites_test.go` runs and FAILS (expected)
- Test output clearly shows: "Expected `{\"data\": ...}`, got `{\"data\": ..., \"error\": null}`"

---

### Phase 2: Response Helper Refactoring
**Goal:** Update `internal/api/response.go` to match new spec

**Current Code:**
```go
type Envelope struct {
    Data  interface{} `json:"data"`
    Meta  interface{} `json:"meta,omitempty"`
    Error *Error      `json:"error,omitempty"`
}
```

**New Code:**
```go
// Success envelope (no error field)
type SuccessEnvelope struct {
    Data interface{} `json:"data"`
    Meta interface{} `json:"meta,omitempty"`
}

// Error envelope (no data field)
type ErrorEnvelope struct {
    Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

**New Helper Functions:**
```go
// Success writes a successful response with data
func Success(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(SuccessEnvelope{Data: data})
}

// SuccessWithMeta writes a successful response with data and metadata
func SuccessWithMeta(w http.ResponseWriter, status int, data interface{}, meta interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(SuccessEnvelope{Data: data, Meta: meta})
}

// Error writes an error response
func Error(w http.ResponseWriter, status int, code, message string, details map[string]interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(ErrorEnvelope{
        Error: ErrorDetail{
            Code:    code,
            Message: message,
            Details: details,
        },
    })
}

// Common error shortcuts
func BadRequest(w http.ResponseWriter, message string) {
    Error(w, http.StatusBadRequest, "BAD_REQUEST", message, nil)
}

func ValidationError(w http.ResponseWriter, message, field, constraint string) {
    Error(w, http.StatusBadRequest, "VALIDATION_FAILED", message, map[string]interface{}{
        "field":      field,
        "constraint": constraint,
    })
}

func Unauthorized(w http.ResponseWriter, message string) {
    Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

func NotFound(w http.ResponseWriter, code, message string) {
    Error(w, http.StatusNotFound, code, message, nil)
}

func InternalError(w http.ResponseWriter, err error) {
    // Log the actual error
    log.Printf("Internal error: %v", err)
    // Return generic message to client
    Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred", nil)
}

func Conflict(w http.ResponseWriter, message string) {
    Error(w, http.StatusConflict, "CONFLICT", message, nil)
}

func RateLimitExceeded(w http.ResponseWriter, message string) {
    Error(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", message, nil)
}
```

**Tasks:**
1. ✅ Refactor `internal/api/response.go` with new structs and functions
2. ✅ Delete old `JSON()`, `ErrorResponse()` functions (breaking change OK)
3. ✅ Add comprehensive tests for `response.go` itself
4. ✅ Document usage in comments

**Acceptance Criteria:**
- `go test ./internal/api/...` passes
- All new helpers produce correct JSON structure
- No references to old `api.JSON()` remain

---

### Phase 3: Handler Migration (Batch 1: Core)
**Goal:** Migrate critical handlers first

**Handlers to migrate:**
1. `auth_handlers.go` - Login, Logout, UserMe, AuthStatus
2. `system.go` - Health, Config, Limits
3. `sites_test.go` reference tests now PASS

**Migration Pattern (per handler):**

**Before (auth_handlers.go:42):**
```go
json.NewEncoder(w).Encode(map[string]string{
    "username": session.Username,
    "version":  serverVersion,
})
```

**After:**
```go
api.Success(w, http.StatusOK, map[string]string{
    "username": session.Username,
    "version":  serverVersion,
})
```

**Before (auth_handlers.go:96):**
```go
w.WriteHeader(http.StatusUnauthorized)
json.NewEncoder(w).Encode(map[string]string{
    "error": "Invalid username or password",
})
```

**After:**
```go
api.Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid username or password", nil)
```

**Tasks:**
1. ✅ Migrate `auth_handlers.go`
2. ✅ Migrate `system.go`
3. ✅ Run tests: `go test ./internal/handlers/auth_test.go ./internal/handlers/system_test.go`
4. ✅ Manual smoke test: `go run ./cmd/server server start` and test `/api/login`

**Acceptance Criteria:**
- All tests in `auth_test.go` and `system_test.go` pass
- No `http.Error()` or `json.NewEncoder()` direct calls in migrated handlers
- Manual API calls return correct format

---

### Phase 4: Handler Migration (Batch 2: Hosting)
**Goal:** Migrate site/deployment handlers

**Handlers:**
1. `deploy.go` - DeployHandler
2. `hosting.go` - ListSitesHandler, DeleteSiteHandler, EnvVarsHandler, APIKeysHandler
3. `site_files.go` - GetSiteFilesHandler, GetFileContentHandler

**Critical: DELETE Route Changes**

**Before (`hosting.go`):**
```go
mux.HandleFunc("DELETE /api/sites", DeleteSiteHandler)
// Usage: DELETE /api/sites?site_id=blog
```

**After:**
```go
mux.HandleFunc("DELETE /api/sites/{id}", DeleteSiteHandler)
// Usage: DELETE /api/sites/blog
```

**Handler Update:**
```go
// Before
siteID := r.URL.Query().Get("site_id")

// After
siteID := r.PathValue("id")  // Go 1.22+ feature
```

**Tasks:**
1. ✅ Write tests: `deploy_test.go`, `hosting_test.go`, `site_files_test.go`
2. ✅ Migrate handlers to use `api.*` helpers
3. ✅ Update DELETE routes in `cmd/server/main.go` router
4. ✅ Update DELETE handler implementations to use `r.PathValue()`
5. ✅ Run tests: `go test ./internal/handlers/...`

**Acceptance Criteria:**
- `DELETE /api/sites/blog` works (not query param)
- `DELETE /api/envvars/123` works
- `DELETE /api/keys/456` works
- All tests pass

---

### Phase 5: Handler Migration (Batch 3: Analytics & Redirects)
**Goal:** Migrate analytics and redirect handlers

**Handlers:**
1. `redirects.go` - ListRedirects, CreateRedirect, DeleteRedirect
2. `webhooks.go` - ListWebhooks, CreateWebhook, UpdateWebhook, DeleteWebhook
3. `events.go` - GetEvents, GetStats, GetDomains, GetTags
4. `logs.go` - GetSiteLogsHandler

**Special Case: Pagination**

**Before (`events.go`):**
```go
// Returns raw array
json.NewEncoder(w).Encode(events)
```

**After:**
```go
// Return with metadata
api.SuccessWithMeta(w, http.StatusOK, events, map[string]interface{}{
    "total":  totalCount,
    "limit":  limit,
    "offset": offset,
})
```

**Tasks:**
1. ✅ Write tests for all handlers
2. ✅ Migrate to `api.*` helpers
3. ✅ Add pagination metadata where applicable (events, domains, tags)
4. ✅ Run tests: `go test ./internal/handlers/...`

**Acceptance Criteria:**
- All endpoints return `{"data": ...}` or `{"error": ...}`
- No raw JSON arrays
- Pagination metadata present on list endpoints

---

### Phase 6: Router & Integration
**Goal:** Update router, verify end-to-end

**Tasks:**
1. ✅ Review all route registrations in `cmd/server/main.go`
2. ✅ Ensure all DELETE routes use `/{id}` pattern
3. ✅ Add integration test: `test_api_standardization.sh`
   ```bash
   #!/bin/bash
   # Test all endpoints return correct format

   # Success case
   curl -s http://localhost:8080/api/system/health | jq '.data' || exit 1

   # Error case
   curl -s http://localhost:8080/api/sites/nonexistent | jq '.error.code' || exit 1

   echo "✅ All API responses follow standard format"
   ```
4. ✅ Update `probe_api.sh` to verify new format
5. ✅ Run full test suite: `go test ./...`

**Acceptance Criteria:**
- `go test ./...` passes (100% of written tests)
- `test_api_standardization.sh` passes
- No `http.Error()` or raw `json.Encode()` in handlers
- All routes use RESTful patterns

---

### Phase 7: Documentation & Cleanup
**Goal:** Update docs, remove old code

**Tasks:**
1. ✅ Update `koder/docs/admin-api/spec.md` with new format
2. ✅ Regenerate `request-response.md` using `probe_api.sh`
3. ✅ Add OpenAPI 3.0 spec (optional but recommended)
   - File: `koder/docs/admin-api/openapi.yaml`
4. ✅ Update CHANGELOG.md:
   ```markdown
   ## [0.8.0] - 2025-12-XX
   ### BREAKING CHANGES
   - API responses now use standardized envelope format
   - Success: `{"data": ...}`, Error: `{"error": {...}}`
   - DELETE endpoints now use path parameters (e.g., `/api/sites/{id}`)
   ```
5. ✅ Delete legacy helper functions from `internal/api/response.go` (if any remain)

**Acceptance Criteria:**
- All docs reflect new API
- No references to old format
- CHANGELOG clearly marks breaking change

---

## 6. Migration Checklist

**Use this to track progress:**

### Handlers (13 total)

- [ ] `auth_handlers.go` (4 endpoints)
  - [ ] POST /api/login
  - [ ] POST /api/logout
  - [ ] GET /api/user/me
  - [ ] GET /api/auth/status

- [ ] `system.go` (3 endpoints)
  - [ ] GET /api/system/health
  - [ ] GET /api/system/config
  - [ ] GET /api/system/limits

- [ ] `deploy.go` (1 endpoint)
  - [ ] POST /api/deploy

- [ ] `hosting.go` (5 endpoints)
  - [ ] GET /api/sites
  - [ ] DELETE /api/sites/{id} ⚠️ Route change
  - [ ] GET /api/envvars?site_id=X
  - [ ] POST /api/envvars
  - [ ] DELETE /api/envvars/{id} ⚠️ Route change

- [ ] `api.go` (3 endpoints - API Keys)
  - [ ] GET /api/keys
  - [ ] POST /api/keys
  - [ ] DELETE /api/keys/{id} ⚠️ Route change

- [ ] `redirects.go` (3 endpoints)
  - [ ] GET /api/redirects
  - [ ] POST /api/redirects
  - [ ] DELETE /api/redirects/{id} (already correct)

- [ ] `webhooks.go` (4 endpoints)
  - [ ] GET /api/webhooks
  - [ ] POST /api/webhooks
  - [ ] PUT /api/webhooks/{id}
  - [ ] DELETE /api/webhooks/{id} (already correct)

- [ ] `events.go` (4 endpoints)
  - [ ] GET /api/events
  - [ ] GET /api/stats
  - [ ] GET /api/domains
  - [ ] GET /api/tags

- [ ] `logs.go` (2 endpoints)
  - [ ] GET /api/logs?site_id=X
  - [ ] GET /api/deployments

- [ ] `site_files.go` (2 endpoints)
  - [ ] GET /api/sites/{id}/files
  - [ ] GET /api/sites/{id}/files/{path...}

- [ ] `track.go` (3 endpoints)
  - [ ] GET /t/p.gif (pixel tracking)
  - [ ] POST /t/event
  - [ ] GET /r/{slug} (redirect tracking)

- [ ] `config.go` (1 endpoint)
  - [ ] GET /api/config

**Total Endpoints:** ~35

### Tests

- [ ] `testutil/helpers.go` created
- [ ] `handlers_test.go` (shared setup) created
- [ ] `auth_test.go` (5 test cases)
- [ ] `system_test.go` (3 test cases)
- [ ] `deploy_test.go` (5 test cases)
- [ ] `hosting_test.go` (5 test cases)
- [ ] `redirects_test.go` (5 test cases)
- [ ] `webhooks_test.go` (5 test cases)
- [ ] `events_test.go` (5 test cases)
- [ ] `logs_test.go` (3 test cases)
- [ ] `site_files_test.go` (3 test cases)
- [ ] Integration test script: `test_api_standardization.sh`

**Target:** 80%+ coverage on `internal/handlers`

### Documentation

- [ ] `koder/docs/admin-api/spec.md` updated
- [ ] `koder/docs/admin-api/request-response.md` regenerated
- [ ] `koder/docs/admin-api/openapi.yaml` created (optional)
- [ ] `CHANGELOG.md` updated (breaking changes noted)
- [ ] `README.md` API section updated (if exists)

---

## 7. Success Criteria

### Functional

- ✅ All 35 endpoints return standardized format
- ✅ Success responses: `{"data": ...}` or `{"data": ..., "meta": ...}`
- ✅ Error responses: `{"error": {"code": "...", "message": "..."}}`
- ✅ All DELETE routes use `/{id}` pattern
- ✅ HTTP status codes align with response type (2xx = data, 4xx/5xx = error)

### Testing

- ✅ `go test ./...` passes with 0 failures
- ✅ 80%+ code coverage on `internal/handlers`
- ✅ Integration test script passes
- ✅ Manual smoke test of all endpoints

### Code Quality

- ✅ No `http.Error()` in handlers (use `api.Error()`)
- ✅ No raw `json.NewEncoder()` in handlers (use `api.Success()`)
- ✅ No `map[string]interface{}` for errors (use `api.Error()`)
- ✅ All error codes follow naming convention (UPPERCASE_SNAKE_CASE)
- ✅ Consistent field naming (snake_case in JSON responses)

### SPA Readiness

- ✅ Single fetch wrapper can handle all endpoints:
  ```javascript
  async function apiFetch(url, options) {
    const resp = await fetch(url, options);
    const json = await resp.json();

    if (resp.ok) {
      return json.data;  // Always exists for 2xx
    } else {
      throw new Error(json.error.message);  // Always exists for 4xx/5xx
    }
  }
  ```
- ✅ Error codes can drive UI behavior (e.g., `SESSION_EXPIRED` → redirect to login)
- ✅ Field-level validation errors populate form errors

---

## 8. Rollout Plan

### Development (This Session)
1. Implement Phases 1-7 sequentially
2. Run tests after each phase
3. Commit after each phase passes

### Pre-Release (v0.8.0-rc1)
1. Tag as release candidate
2. Test against local CLI client
3. Build sample SPA dashboard page (single endpoint)
4. Verify real-world usage

### Release (v0.8.0)
1. Merge to `master`
2. Tag `v0.8.0`
3. Update CHANGELOG
4. Rebuild admin UI against new API

### Post-Release
1. Monitor for edge cases
2. Add tests for any discovered issues
3. Optimize response sizes if needed

---

## 9. Risks & Mitigations

| Risk | Impact | Mitigation |
|:---|:---|:---|
| **Breaking CLI client** | High | Update CLI in same PR, test before merge |
| **Missed edge cases** | Medium | Comprehensive test suite catches most issues |
| **Performance regression** | Low | New format is smaller (no null fields) |
| **Developer confusion** | Low | Clear docs + helper functions make it obvious |
| **Incomplete migration** | High | Use checklist, grep for `http.Error` and `json.NewEncoder` |

---

## 10. References

**Standards:**
- [RFC 7807: Problem Details for HTTP APIs](https://datatracker.ietf.org/doc/html/rfc7807)
- [JSend Specification](https://github.com/omniti-labs/jsend)
- [JSON:API](https://jsonapi.org/)
- [Stripe API Design](https://stripe.com/docs/api)
- [GitHub API v3](https://docs.github.com/en/rest)

**Fazt Docs:**
- `koder/analysis/07_gemini-api-design-suggestions.md` (Original proposal)
- `koder/docs/admin-api/request-response.md` (Current ground truth)
- `koder/analysis/04_comprehensive_technical_overview.md` (Architecture)

---

## 11. Open Questions

1. **Nested Resources:** Should we use `/api/sites/{id}/logs` or `/api/logs?site_id={id}`?
   - **Recommendation:** Keep query params for now (less breaking). Revisit in v0.9.0.

2. **DELETE Response Body:** Should DELETE return `{"data": {"message": "Deleted"}}` or `204 No Content`?
   - **Recommendation:** `200 OK` with `{"data": {"message": "..."}}` for consistency.

3. **Pagination Defaults:** Should we enforce pagination on all lists?
   - **Recommendation:** Add `limit`/`offset` to events/logs, but don't enforce on small datasets (sites, redirects).

4. **Error Details Schema:** Should we formalize the `details` object structure?
   - **Recommendation:** Keep flexible for now, document common patterns (field, constraint).

---

**End of Plan**
