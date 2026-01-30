# Fazt Implementation State

**Last Updated**: 2026-01-30
**Current Version**: v0.13.0

## Status

State: **CLEAN** - Plan 29 released, all docs updated

---

## Active Directions

See `koder/THINKING_DIRECTIONS.md` for full list.

**Next up**:
- P2: Nexus App (stress test all capabilities including private/)
- E4: Plan 24 - Mock OAuth (local auth testing)
- E5: Plan 25 - SQL Command (remote DB debugging)

---

## Pending Plans

| Plan | Status | Purpose |
|------|--------|---------|
| 30: User Isolation & Analytics | Planning | User data isolation, analytics, GDPR |
| 29: Private Directory | ✅ Released v0.13.0 | `private/` with dual access |
| 24: Mock OAuth | Not started | Local auth testing |
| 25: SQL Command | Not started | Remote DB debugging |

---

## RBAC Notes (For Future Discussion)

Full RBAC was discussed but deferred. Key points:

**Pros:**
- Fine-grained permissions (`posts:create`, `posts:delete:own`)
- Custom roles per app (`editor`, `moderator`, `viewer`)
- Framework-level enforcement

**Cons:**
- Significant API complexity
- Most apps need only owner/admin/user
- Can be implemented at app level if needed
- Delays other features

**Decision:** Keep simple roles (owner/admin/user). Revisit when needed.

---

## Current Session (2026-01-30)

**Plan 30: User Isolation & Analytics - Design Discussion**

### Key Decisions

1. **API Namespace (Option C2)**
   - `fazt.app.user.*` - User's private data (auto-scoped)
   - `fazt.app.*` - App's shared data
   - `fazt.app.private.*` - Bundled private files
   - `fazt.auth.*` - Authentication
   - `fazt.admin.*` - Admin operations
   - `fazt.analytics.*` - Event tracking

2. **ID Format (Stripe-style)**
   - `fazt_usr_<12 chars>` - User
   - `fazt_app_<12 chars>` - App
   - `fazt_tok_<12 chars>` - Token
   - `fazt_ses_<12 chars>` - Session

3. **Analytics Enhancement**
   - Add `app_id`, `user_id` to events table
   - Query by app, user, or both
   - Track user journey across apps

4. **GDPR Compliance**
   - `fazt.admin.users.delete(userId)` removes all data

5. **Role Model**
   - One owner per instance
   - Multiple admins allowed
   - Simple roles: owner/admin/user

### Plan Created

`koder/plans/30_user_isolation_analytics.md`

---

## Previous Session (2026-01-30)

**Plan 29: Private Directory - Full Implementation**

### Features Delivered

1. **Dual Access Model**
   - HTTP `GET /private/*` → Auth check → stream (401 if not logged in)
   - Serverless `fazt.private.*` → Direct read for code logic
   - Large files stream without serverless overhead
   - Small data files processed by serverless

2. **Serverless API**
   - `fazt.private.read(path)` → string (undefined if missing)
   - `fazt.private.readJSON(path)` → object (null if missing)
   - `fazt.private.exists(path)` → boolean
   - `fazt.private.list()` → string[]

3. **Deploy Flag `--include-private`**
   - Warns when `private/` is gitignored but exists
   - Use flag to explicitly include gitignored private files
   - Prevents accidental deployment of sensitive data

### Files Changed

- `cmd/server/main.go` - Auth-gated serving, `createDeployZipWithOptions`
- `cmd/server/app.go` - `--include-private` flag
- `internal/runtime/private_bindings.go` - Serverless API (new)
- `internal/runtime/private_bindings_test.go` - 37 tests (new)
- `internal/runtime/handler.go` - Private injector
- `internal/config/config.go` - Version bump

### Knowledge-Base Updates

- `cli-app.md` - Added `--include-private` flag
- `deployment.md` - New "Private Directory" section
- `frontend-patterns.md` - Added `private/` to project structure
- `serverless-api.md` - Full `fazt.private.*` documentation
- `version.json` - Bumped to 0.13.0

### Future Enhancement Notes (in plan)

| Library | Enables | Use Case |
|---------|---------|----------|
| YAML parser | `readYAML()` | Config in YAML |
| sql.js | SQLite queries | `private/app.db` |
| CSV parser | `readCSV()` | Spreadsheet data |

---

## Quick Reference

```bash
# Deploy with private files
fazt app deploy ./my-app --to zyt --include-private

# Access private in serverless
var config = fazt.private.readJSON('config.json')

# HTTP access (requires auth)
curl https://my-app.zyt.app/private/data.json  # 401 if not logged in
```
