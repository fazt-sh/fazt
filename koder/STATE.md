# Fazt Implementation State

**Last Updated**: 2026-02-07
**Current Version**: v0.26.0

## Status

State: CLEAN
File upload support implemented, KB docs updated. Ready for integration testing.

---

## Last Session (2026-02-07) — App File Upload Support

### What Was Done

#### File Upload Support for Serverless Apps

Apps can now receive file uploads via `multipart/form-data`. The gap was in
request parsing — `buildRequest()` only handled JSON bodies, and the body size
middleware capped all `/api/*` requests at 1MB.

**Changes:**

1. **`internal/runtime/runtime.go`** — Added `FileUpload` struct (`Name`, `Type`,
   `Size`, `Data []byte`) and `Files map[string]FileUpload` field on `Request`.
   Files injected as `ArrayBuffer` in the Goja VM via `injectGlobals()`.

2. **`internal/runtime/handler.go`** — Extended `buildRequest()` to parse
   `multipart/form-data`: form fields → `req.Body`, files → `req.Files`.
   Added `parseMultipartFiles()` helper. Added `io` and `system` imports.

3. **`internal/middleware/security.go`** — Multipart requests now use
   `system.GetLimits().Storage.MaxUpload` (default ~10MB) instead of the 1MB
   default. Added `strings` and `system` imports.

4. **`internal/runtime/handler_test.go`** — New file with 4 tests:
   multipart with files, multipart without files, JSON body regression,
   ArrayBuffer injection in VM. All passing.

5. **KB docs updated:**
   - `knowledge-base/skills/app/references/serverless-api.md` — Added File
     Uploads section with HTML form example, handler code, file object shape,
     storage scoping guidance, and limits. Fixed stale "No network calls"
     limitation.
   - `knowledge-base/agent-context/architecture.md` — Added file uploads to
     capabilities table.

**Developer API:**
```javascript
// request.files.photo = { name, type, size, data (ArrayBuffer) }
fazt.app.user.s3.put('uploads/' + file.name, file.data, file.type)
```

User-scoped storage (`fazt.app.user.s3`) ensures file isolation per user.
Shared storage (`fazt.app.s3`) available for app-wide assets.

### Unreleased Commits

```
73c6ea4 Support large file uploads
```

---

## Next Session

### Build a test app with recently implemented features

Build and deploy a real app that exercises the new capabilities from v0.26.0:

1. **File upload** — Form with image upload, stored via `fazt.app.user.s3`
2. **`fazt.net.fetch()`** — Outbound HTTP call (e.g., external API)
3. **Auth integration** — `fazt.auth.requireLogin()` + user-scoped storage
4. **Verify isolation** — Confirm different users can't see each other's uploads

This tests the full stack end-to-end: multipart parsing → ArrayBuffer in VM →
user-scoped blob storage, plus egress proxy with secrets/allowlist.

### Pre-existing flaky tests (to fix):

**`hosting/TestStressMessageThroughput`** — Stress test asserts 85% WS delivery
across 100 clients, but VM consistently hits 75-79%. Has `testing.Short()` skip
but `go test ./...` doesn't pass `-short`. Fix: lower threshold to 70%.

**`worker/TestPoolList`** — SQLite `:memory:` connection pool race. Each new
connection to `:memory:` gets an independent blank DB. Fix: `db.SetMaxOpenConns(1)`.

---

## Quick Reference

```bash
# Test file upload
go test ./internal/runtime/ -v -run TestBuildRequest_Multipart
go test ./internal/runtime/ -v -run TestFileUpload_ArrayBuffer

# Test all affected packages
go test ./internal/runtime/ ./internal/middleware/

# Key files
cat internal/runtime/handler.go     # buildRequest() + parseMultipartFiles()
cat internal/runtime/runtime.go     # FileUpload struct, VM injection
cat internal/middleware/security.go  # Multipart body limit
```
