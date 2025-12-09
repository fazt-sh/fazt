# Fazt.sh Assistant Guide

**Reference**: `koder/start.md` contains the detailed architectural map.

---

## 🧠 Core Philosophy

*   **Cartridge Architecture**: One Binary (`fazt`) + One SQLite DB (`data.db`).
*   **Zero Dependencies**: Pure Go + `modernc.org/sqlite`. NO CGO. No external runtimes.
*   **VFS**: User sites/assets live in the DB. No disk I/O for hosting.
*   **System Sites**: Reserved sites (`root`, `404`) MUST be seeded from `go:embed` assets on startup if missing. They live in VFS but originate from the binary.
*   **Safety**: `CGO_ENABLED=0` always.

---

## 🔨 Build & Test

### Test-First Methodology
*   **Philosophy**: When building a feature, implement all tests first and build to pass the tests.
*   **Build (Local)**: `go build -o fazt ./cmd/server`
*   **Test (All)**: `go test ./...`
*   **Run (Dev)**: `go run ./cmd/server server start --domain localhost --port 8080`
*   **Env Differences**: Recognize what can be tested in the coding environment and build tests accordingly.

### ⚠️ Testing Server Stability
When testing if the server runs without crashing:

*   **DO NOT** use bash backgrounding (`&`) to test server stability
    - Backgrounded processes appear to "work" but you cannot detect silent exits
*   **PREFERRED METHODS**:
    1. Use `timeout 10 go run ./cmd/server server start` to test for N seconds
    2. Run in foreground and check logs for panics/errors
    3. Use `ps aux | grep server` AFTER a delay to verify process persistence

*   **CORRECT PATTERN**:
    ```bash
    # Start server
    go run ./cmd/server server start --port 8080 &

    # Wait for startup (3-5 seconds)
    sleep 3

    # Verify it's still running
    ps aux | grep "server start" | grep -v grep

    # Test endpoints
    curl http://localhost:8080/api/system/health

    # Clean up
    pkill -f "server start"
    ```

---

## 📦 Release Workflow

**Detailed Guide**: `koder/workflows/ON_NEW_VERSION.md`

1.  **Code**: Implement feature/fix.
2.  **Test**: `go test ./...` (MUST PASS).
3.  **Changelog**: Update `CHANGELOG.md` AND `docs/changelog.json`.
4.  **Tag**: `git tag vX.Y.Z && git push origin master --tags`.
5.  **Build**: GitHub Action auto-builds release (Version injected via ldflags).

**Current Version**: v0.7.2 (released), v0.8.0-dev (in development)

---

## 📂 Structure

### Key Directories
*   `cmd/server/`: Main entry point and CLI commands.
*   `internal/api/`: **NEW** - Standardized API response helpers.
*   `internal/handlers/`: HTTP handlers for all endpoints.
*   `internal/handlers/testutil/`: **NEW** - Test utilities and helpers.
*   `internal/provision/`: Systemd, Install, User management.
*   `internal/hosting/`: VFS, Deploy, CertMagic.
*   `internal/database/`: Migrations, Query logic.
*   `internal/auth/`: Session management, rate limiting.
*   `internal/config/`: Configuration management.
*   `install.sh`: The "curl | bash" installer.

### Key Documentation
*   `koder/start.md`: Bootstrap entry point for new sessions.
*   `koder/NEXT_SESSION.md`: Current status and next steps.
*   `koder/plans/`: Implementation plans and specifications.
*   `koder/analysis/`: Technical analysis documents.
*   `koder/docs/admin-api/`: API documentation.

---

## 🌐 API Design (v0.8.0+)

### Standardized Response Format

**Success Response:**
```json
{"data": {...}}
{"data": [...], "meta": {"total": 100, "limit": 20, "offset": 0}}
```

**Error Response:**
```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Site name is required",
    "details": {"field": "site_name", "constraint": "required"}
  }
}
```

### Using API Helpers (internal/api/response.go)

**Success:**
```go
api.Success(w, http.StatusOK, data)
api.SuccessWithMeta(w, http.StatusOK, data, meta)
```

**Errors:**
```go
api.BadRequest(w, "message")
api.Unauthorized(w, "message")
api.NotFound(w, "ERROR_CODE", "message")
api.ValidationError(w, "message", "field", "constraint")
api.InternalError(w, err)
```

**See:** `koder/plans/11_api-standardization.md` for complete specification.

---

## 🧪 Testing Infrastructure

### Test Utilities (internal/handlers/testutil/)

**Response Checkers:**
```go
testutil.CheckSuccess(t, rr, 200)          // Validates {"data": ...}
testutil.CheckError(t, rr, 400, "BAD_REQUEST")  // Validates {"error": ...}
testutil.CheckSuccessArray(t, rr, 200)     // For array responses
```

**Request Helpers:**
```go
req := testutil.JSONRequest("POST", "/api/endpoint", body)
req = testutil.WithSession(req, sessionID)  // Add session cookie
req = testutil.WithAuth(req, token)         // Add Bearer token
```

### Shared Test Setup (internal/handlers/handlers_test.go)

```go
db := setupTestDB(t)                    // In-memory SQLite with full schema
defer cleanupTestDB(t, db)
store, sessionID := setupTestAuth(t)    // Session store + valid session
setupTestConfig(t)                      // Test configuration
createTestSite(t, db, "mysite")         // Create test data
```

**See:** `internal/handlers/auth_test.go` for reference test patterns.

---

## 🛠️ Environment

*   **Git**: Container has credentials. Can commit & push to `origin/master`.
*   **Container Setup**: Running inside a podman container, with no systemd.
*   **Host**: Host machine is a Mac M4.
*   **Deployment Target**: Digital Ocean Droplet, x86/Ubuntu 24.04 LTS, $6 base instance.
*   **Note**: `GEMINI.md` is symlinked to this file.

---

## 📋 Common Patterns

### Handler Pattern (Post API Standardization)
```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    // Validate
    if input == "" {
        api.ValidationError(w, "Field required", "input", "required")
        return
    }

    // Process
    result, err := doSomething(input)
    if err != nil {
        api.InternalError(w, err)
        return
    }

    // Success
    api.Success(w, http.StatusOK, result)
}
```

### Test Pattern
```go
func TestMyHandler_Success(t *testing.T) {
    // Setup
    silenceTestLogs(t)
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    store, sessionID := setupTestAuth(t)

    // Request
    req := testutil.JSONRequest("GET", "/api/endpoint", nil)
    req = testutil.WithSession(req, sessionID)
    rr := httptest.NewRecorder()

    // Execute
    MyHandler(rr, req)

    // Assert
    data := testutil.CheckSuccess(t, rr, 200)
    testutil.AssertFieldExists(t, data, "expected_field")
}
```

---

## ⚠️ Important Notes

*   **fazt.sh** is just a proposed URL; it's NOT LIVE, don't use it anywhere in code.
*   **API Format**: Always use `internal/api` helpers, never raw `http.Error()` or `json.NewEncoder()`.
*   **Test Coverage**: Write tests for all new handlers. Aim for 80%+ coverage.
*   **Error Codes**: Use standardized error codes (see `koder/plans/11_api-standardization.md` section 3.2).
*   **Session Cookies**: Use name `cc_session` (not `fazt_session`).

---

## 🚀 Quick Start for New Sessions

**Always start with:**
```
read and execute koder/start.md
```

This will:
1. Load current mission from `koder/NEXT_SESSION.md`
2. Load relevant context files
3. Verify environment
4. State readiness for first step

**Bootstrap chain ensures context is always loaded correctly.**
