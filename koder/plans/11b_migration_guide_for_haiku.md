# Migration Guide for API Standardization (Haiku)

**Date**: December 9, 2025
**For**: Haiku Agent
**Task**: Migrate handlers 3-30 to use standardized API helpers

---

## ✅ What Sonnet Completed (Phase 1-2)

1. **Test Infrastructure**: `internal/handlers/testutil/helpers.go`
2. **Shared Test Setup**: `internal/handlers/handlers_test.go`
3. **Response Helpers**: `internal/api/response.go` (fully refactored)
4. **Reference Handler**: `auth_handlers.go` (FULLY MIGRATED)
5. **Reference Tests**: `auth_test.go` (WORKING EXAMPLE)

---

## 🎯 Your Mission

Migrate the remaining **28 handlers** using the exact pattern shown in `auth_handlers.go`.

###Keys Checklist (Start Here)

**Handlers to Migrate**:
- [ ] `deploy.go` (1 endpoint)
- [ ] `hosting.go` (5 endpoints)
- [ ] `api.go` (API keys - 3 endpoints)
- [ ] `config.go` (1 endpoint)
- [ ] `logs.go` (2 endpoints)
- [ ] `events.go` (4 endpoints)
- [ ] `system.go` (3 endpoints) - Partially done, needs cleanup
- [ ] `track.go` (3 endpoints)

**Already Compatible** (using deprecated helpers):
- [x] `redirects.go` - Uses `api.JSON`, `api.ServerError` (backward compatible)
- [x] `webhooks.go` - Uses `api.JSON`, `api.ServerError` (backward compatible)
- [x] `site_files.go` - Uses `api.JSON`, `api.ServerError` (backward compatible)

---

## 📖 Migration Pattern

### Pattern 1: Replace `http.Error()`

**BEFORE:**
```go
http.Error(w, "Unauthorized", http.StatusUnauthorized)
```

**AFTER:**
```go
api.Unauthorized(w, "Authentication required")
```

**Available helpers:**
- `api.BadRequest(w, "message")`
- `api.Unauthorized(w, "message")`
- `api.NotFound(w, "ERROR_CODE", "message")` ← Note: 2 params!
- `api.InternalError(w, err)`
- `api.RateLimitExceeded(w, "message")`
- `api.InvalidCredentials(w)` ← No message needed
- `api.SessionExpired(w)` ← No message needed
- `api.InvalidAPIKey(w)` ← No message needed

### Pattern 2: Replace raw `json.NewEncoder().Encode()`

**BEFORE (success):**
```go
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(data)
```

**AFTER:**
```go
api.Success(w, http.StatusOK, data)
```

**BEFORE (with metadata/pagination):**
```go
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(data)
```

**AFTER:**
```go
meta := map[string]interface{}{"total": count, "limit": limit, "offset": offset}
api.SuccessWithMeta(w, http.StatusOK, data, meta)
```

### Pattern 3: Replace custom error maps

**BEFORE:**
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusBadRequest)
json.NewEncoder(w).Encode(map[string]string{
    "error": "Invalid request",
})
```

**AFTER:**
```go
api.BadRequest(w, "Invalid request")
```

### Pattern 4: Replace `jsonError()` helper

**BEFORE:**
```go
jsonError(w, "Site not found", http.StatusNotFound)
```

**AFTER:**
```go
api.NotFound(w, "SITE_NOT_FOUND", "Site not found")
```

---

## 🔧 Step-by-Step Process (FOR EACH HANDLER)

### Step 1: Read the Reference

```bash
# Open these files side-by-side
cat internal/handlers/auth_handlers.go   # The GOOD example
cat internal/handlers/deploy.go          # Your current target
```

### Step 2: Add Import

At the top of the file, add:
```go
import (
    "github.com/fazt-sh/fazt/internal/api"
    // ... other imports
)
```

### Step 3: Find and Replace Patterns

Use your editor's find/replace or manually update:

1. Find: `http.Error(w, "message", http.StatusBadRequest)`
   Replace: `api.BadRequest(w, "message")`

2. Find: `http.Error(w, "message", http.StatusUnauthorized)`
   Replace: `api.Unauthorized(w, "message")`

3. Find: `http.Error(w, "message", http.StatusNotFound)`
   Replace: `api.NotFound(w, "NOT_FOUND", "message")`

4. Find: `http.Error(w, "message", http.StatusInternalServerError)`
   Replace: `api.InternalError(w, err)` ← Pass the error variable!

5. Find blocks like:
   ```go
   w.Header().Set("Content-Type", "application/json")
   json.NewEncoder(w).Encode(data)
   ```
   Replace:
   ```go
   api.Success(w, http.StatusOK, data)
   ```

### Step 4: Write Tests (Copy from `auth_test.go`)

**Template:**
```go
// Test<HandlerName>_Success tests successful operation
func Test<HandlerName>_Success(t *testing.T) {
    // Setup
    silenceTestLogs(t)
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    store, sessionID := setupTestAuth(t)
    setupTestConfig(t)
    InitAuth(store, auth.NewRateLimiter(), "v0.8.0-test")

    // TODO: Add any test data needed
    // createTestSite(t, db, "mysite")

    // Create request
    req := testutil.JSONRequest("GET", "/api/...", nil)
    req = testutil.WithSession(req, sessionID)

    rr := httptest.NewRecorder()

    // Execute
    <HandlerName>(rr, req)

    // Assert
    data := testutil.CheckSuccess(t, rr, 200)
    testutil.AssertFieldExists(t, data, "expected_field")
}

// Test<HandlerName>_ValidationError tests validation failure
func Test<HandlerName>_ValidationError(t *testing.T) {
    // Setup
    silenceTestLogs(t)
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    store, sessionID := setupTestAuth(t)
    InitAuth(store, auth.NewRateLimiter(), "v0.8.0-test")

    // Create request with invalid data
    req := testutil.JSONRequest("POST", "/api/...", map[string]string{
        "field": "", // Invalid!
    })
    req = testutil.WithSession(req, sessionID)

    rr := httptest.NewRecorder()

    // Execute
    <HandlerName>(rr, req)

    // Assert
    testutil.CheckError(t, rr, 400, "VALIDATION_FAILED")
}

// Test<HandlerName>_Unauthorized tests without auth
func Test<HandlerName>_Unauthorized(t *testing.T) {
    // Setup
    silenceTestLogs(t)
    store := auth.NewSessionStore(24 * time.Hour)
    InitAuth(store, auth.NewRateLimiter(), "v0.8.0-test")

    // Create request WITHOUT session
    req := testutil.JSONRequest("GET", "/api/...", nil)

    rr := httptest.NewRecorder()

    // Execute
    <HandlerName>(rr, req)

    // Assert
    testutil.CheckError(t, rr, 401, "UNAUTHORIZED")
}
```

### Step 5: Run Tests

```bash
go test ./internal/handlers/<file>_test.go -v
```

**If tests FAIL:**
- Check error message carefully
- Compare your handler with `auth_handlers.go`
- Verify you're using the correct error code (e.g., "VALIDATION_FAILED", not "BAD_REQUEST")
- Make sure you added `internal/api` import

**If tests PASS:**
```bash
git add internal/handlers/<file>.go internal/handlers/<file>_test.go
git commit -m "Migrate <file>.go to standardized API"
```

### Step 6: Update Checklist

Edit THIS file (`11b_migration_guide_for_haiku.md`) and mark `[x]`:
```markdown
- [x] `deploy.go` (1 endpoint) ✅ DONE
```

### Step 7: Move to Next Handler

Repeat steps 1-6 for the next file in the checklist.

---

## 🚫 Common Mistakes to Avoid

1. **Don't forget the api import:**
   ```go
   import "github.com/fazt-sh/fazt/internal/api" // ADD THIS!
   ```

2. **Don't use wrong NotFound signature:**
   ```go
   // WRONG:
   api.NotFound(w, "Site not found")

   // CORRECT:
   api.NotFound(w, "SITE_NOT_FOUND", "Site not found")
   ```

3. **Don't forget to pass error to InternalError:**
   ```go
   // WRONG:
   api.InternalError(w, nil)

   // CORRECT:
   api.InternalError(w, err) // Pass the actual error!
   ```

4. **Don't leave old response code:**
   ```go
   // WRONG (mixing old and new):
   w.Header().Set("Content-Type", "application/json")
   api.Success(w, http.StatusOK, data)

   // CORRECT:
   api.Success(w, http.StatusOK, data) // It sets headers for you!
   ```

---

## 🔍 Verification

After migrating each file, run:

```bash
./scripts/verify_api_migration.sh
```

**Expected output:**
```
=== API Migration Status ===
Handlers migrated: 5 / 13
Legacy http.Error calls remaining: 45
Legacy json.Encode calls remaining: 32

=== Running Tests ===
...test output...

🟡 Migration in progress...
```

---

## 🛑 When to Stop and Hand Back to Sonnet

Stop if ANY of these happen:

1. **You're stuck on the same file for >3 attempts**
   - Document the issue in NEXT_SESSION.md
   - Move to next file, mark current as "BLOCKED"

2. **Tests are failing and you don't know why**
   - Document which tests fail and the error message
   - Commit what you have so far
   - Mark file as "PARTIAL" in checklist

3. **You've migrated ALL files OR reached >3 blocked files**
   - Update NEXT_SESSION.md with summary
   - List completed files
   - List blocked files with details

---

## 📦 Final Deliverables

When you're done (or blocked), ensure:

1. **All passing migrations are committed**:
   ```bash
   git status # Should show committed changes
   ```

2. **Checklist is updated** in this file (mark [x])

3. **NEXT_SESSION.md is updated** with:
   - What you completed
   - What's blocked (with error messages)
   - What Sonnet needs to do next

---

## 🎓 Learning Resources

**If you're unsure about something:**

1. **Look at auth_handlers.go** - It's the perfect reference
2. **Look at auth_test.go** - It shows how to write tests
3. **Run the tests** - They'll tell you what's wrong
4. **Read the error codes** in `koder/plans/11_api-standardization.md` section 3.2

---

## ✨ Success Criteria

You'll know you're done when:

- [ ] All files in checklist marked `[x]`
- [ ] `go test ./internal/handlers/... -v` shows 80%+ passing
- [ ] `./scripts/verify_api_migration.sh` shows 0 legacy patterns
- [ ] All changes committed to git
- [ ] NEXT_SESSION.md updated for Sonnet

**Good luck! Follow the pattern, run tests often, and commit frequently.**
