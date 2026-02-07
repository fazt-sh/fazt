# Plan 47: Codebase Cleanup

**Model**: Sonnet (mechanical refactoring, no architectural decisions)
**Estimated phases**: 4
**Validation**: `go test ./... -count=1` after every phase

---

## Phase 1: Nomenclature — `site_id` → `app_id`

Rename `site_id` to `app_id` across VFS/hosting code. 125 occurrences in 30 files.

### Scope

**Production code** (rename variable names, function params, struct fields, SQL column references):
- `internal/hosting/` — vfs.go, deploy.go, manager.go, cache_test.go, hosting_test.go
- `internal/handlers/` — hosting.go, site_files.go, logs.go, logs_stream.go, apps_handler.go, apps_handler_v2.go, agent_handler.go, cmd_gateway.go
- `internal/runtime/` — handler.go, private_bindings.go
- `internal/worker/` — pool.go
- `cmd/server/` — main.go, app_v2.go, app_logs.go

**Test code** (update to match):
- All `*_test.go` files that reference `site_id`

**Database migration** (if `site_id` is an actual column name):
- Check if `site_id` is a DB column in the `files` table
- If yes: add migration to rename column, update all SQL queries
- If it's only a Go variable name mapped to a different column: just rename the Go side

### Rules
- Rename Go identifiers: `siteID` → `appID`, `SiteID` → `AppID`, `site_id` → `app_id`
- Also rename related: `DeleteSite` → `DeleteApp` (in hosting context), `GetSiteFiles` → `GetAppFiles`, etc.
- Keep git blame useful: one commit per logical group (hosting, handlers, runtime, tests)
- Run `go test ./... -count=1` after each commit

---

## Phase 2: Dead Code Removal

Remove files and code explicitly marked as dead.

### 2a: In-memory session store (580 lines)
- Delete `internal/auth/session.go` (249 lines) — marked `LEGACY_CODE: In-memory session store is no longer used`
- Delete `internal/auth/session_test.go` (330 lines) — tests for dead code
- Verify no imports reference the deleted types (`SessionStore`, `Session` from this file)

### 2b: V1 app handlers (broken since migration 012)
- `internal/handlers/apps_handler.go` (892 lines) — queries removed columns `a.name`, `a.manifest`
- `internal/handlers/apps_handler_test.go` — tests for broken handlers
- Check if `cmd/server/main.go` still routes to V1 handlers; if so, remove those routes
- V2 handlers (`apps_handler_v2.go`) are the active replacement

### 2c: Old appid format
- `internal/appid/appid.go` — remove `app_` prefix constants, `GenerateOld()`, old format acceptance in `IsValid()`
- `internal/appid/appid_test.go` — remove tests for old format (lines marked LEGACY_CODE)
- Grep for any code still generating `app_` format IDs

### Rules
- Each sub-phase is one commit
- Run `go test ./... -count=1` after each
- If V1 handler removal causes import cycle or routing issues, leave a TODO and move on

---

## Phase 3: LEGACY_CODE Migrations

Migrate code marked with `LEGACY_CODE: Migrate to activity.Log()`.

### Scope
- `internal/handlers/auth_handlers.go` — 4 `audit.Log*()` calls → `activity.Log()`
- `internal/handlers/track.go` — analytics buffer → `activity.Log()`
- `cmd/server/main.go` — audit logging init, analytics buffer init/flush

### Approach
1. First: read `activity.Log()` API to understand the interface
2. Replace `audit.LogSuccess/LogFailure` calls with equivalent `activity.Log()` calls
3. Replace analytics buffer writes with `activity.Log()`
4. Remove `audit` package initialization if no longer used
5. Remove analytics buffer initialization if no longer used

### Rules
- If `activity.Log()` doesn't exist yet or the API is unclear: **STOP**. Leave audit calls as-is, note in STATE.md. This is an Opus task.
- Don't invent new APIs. Only migrate if the target API already exists.

---

## Phase 4: Security Hardening (Mechanical Fixes)

Fix the 4 vulnerabilities found in Phase 7 adversarial tests.

### 4a: Invite TOCTOU fix
- `internal/auth/invites.go` `RedeemInvite()` — wrap GetInvite → IsValid → CreateUser → UPDATE in a database transaction
- The existing adversarial test (`TestAdversarial`) should validate the fix

### 4b: OAuth state TOCTOU fix
- `internal/auth/oauth.go` `ValidateState()` — change SELECT → DELETE to `DELETE ... RETURNING` (single atomic statement)

### 4c: getClientIP hardening
- `internal/handlers/auth_handlers.go` `getClientIP()` — don't trust `X-Forwarded-For` unconditionally
- Only trust if request comes from known proxy (localhost/private ranges), otherwise use `r.RemoteAddr`

### 4d: Login timing side channel
- `internal/handlers/auth_handlers.go` `LoginHandler` — always run bcrypt comparison even for invalid usernames (compare against dummy hash)

### Rules
- Run adversarial tests after each fix: `go test ./cmd/server -run "TestAdversarial" -v -count=1`
- Run full suite: `go test ./... -count=1`
- If a fix requires schema changes or is more complex than expected: **STOP**, note in STATE.md

---

## Commit Strategy

One commit per sub-phase, messages like:
- `Rename site_id to app_id in hosting layer`
- `Remove dead in-memory session store`
- `Remove stale V1 app handlers`
- `Remove old app_ ID format`
- `Migrate audit logging to activity.Log()`
- `Fix invite TOCTOU with transaction`
- `Fix OAuth state TOCTOU with DELETE RETURNING`
- `Harden getClientIP against X-Forwarded-For spoofing`
- `Fix login timing side channel`

## Exit Criteria

- `go test ./... -count=1` all green
- Zero `LEGACY_CODE` markers remaining (or documented why kept)
- Zero `site_id` references in production code
- STATE.md updated with results
