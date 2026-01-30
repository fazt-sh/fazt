# Fazt Implementation State

**Last Updated**: 2026-01-30
**Current Version**: v0.12.0

## Status

State: **IN PROGRESS** - Plan 29 implemented with auth-gated streaming

---

## Active Directions

See `koder/THINKING_DIRECTIONS.md` for full list.

**Current**:
- E8: Plan 29 - Private Directory ✅ Complete

**Next up**:
- P2: Nexus App (stress test all capabilities)
- E4: Plan 24 - Mock OAuth (local auth testing)

---

## Pending Plans

| Plan | Status | Purpose |
|------|--------|---------|
| 29: Private Directory | ✅ Complete | `private/` with dual access modes |
| 24: Mock OAuth | Not started | Local auth testing |
| 25: SQL Command | Not started | Remote DB debugging |

---

## Current Session (2026-01-30)

**Plan 29: Private Directory (Enhanced)**

### Dual Access Model

| Access Mode | Use Case | Behavior |
|-------------|----------|----------|
| HTTP `GET /private/*` | Serve files to users | Auth check → stream (401 if not logged in) |
| Serverless `fazt.private.*` | Process data in code | Direct read for logic |

### Implementation

1. **Auth-gated HTTP** (changed from 403)
   - Unauthenticated → 401 Unauthorized
   - Authenticated → Stream file directly
   - Supports large files (video, images) without serverless overhead

2. **Serverless API** (`fazt.private.*`)
   - `read(path)` - returns file as string
   - `readJSON(path)` - parses JSON
   - `exists(path)` - returns boolean
   - `list()` - returns array of filenames

3. **Files Changed**
   - `cmd/server/main.go` - Auth-gated serving, global `siteAuthService`
   - `internal/runtime/private_bindings.go` - Serverless API
   - `internal/runtime/private_bindings_test.go` - 37 test cases
   - `internal/runtime/handler.go` - Inject private namespace
   - `knowledge-base/skills/app/references/serverless-api.md` - Docs
   - `koder/plans/29_private_directory.md` - Updated spec

4. **Testing**
   - Unit tests: All pass (37 cases including edge cases, traversal, isolation)
   - HTTP 401: Verified
   - Serverless access: Verified

### Deploy: `--include-private` Flag

| `private/` state | Flag | Behavior |
|------------------|------|----------|
| Not gitignored | - | Deploy normally |
| Gitignored | (none) | Warn + skip |
| Gitignored | `--include-private` | Info + include |

```bash
# Warning shown when private/ is gitignored
fazt app deploy ./my-app --to zyt
# Warning: private/ is gitignored but exists
#   Use --include-private to deploy private files

# Explicitly include gitignored private/
fazt app deploy ./my-app --to zyt --include-private
# Including gitignored private/ (5 files)
```

### Future: JS Library Expansion

| Library | Enables | Files |
|---------|---------|-------|
| YAML parser | `readYAML()` | Config in YAML |
| sql.js | SQLite queries | `private/app.db` |
| CSV parser | `readCSV()` | Spreadsheet data |

**Ready for v0.13.0 release.**

---

## Quick Reference

```bash
fazt remote status zyt          # Check production
fazt app deploy ./app --to zyt  # Deploy app
cat knowledge-base/version.json # Check KB version
```
