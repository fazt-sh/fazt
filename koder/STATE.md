# Fazt Implementation State

**Last Updated**: 2026-01-21
**Current Version**: v0.10.7

## Status

State: CLEAN

---

## Last Session (2026-01-20)

### Releases

**v0.10.6** - SQLite busy_timeout
- Added `PRAGMA busy_timeout=5000` to prevent SQLITE_BUSY errors
- Failure rate: 6% → 0%

**v0.10.7** - Storage query fixes
- `findOne` now accepts query objects `{ id, session }`
- `id` field now queryable in `find()`, `update()`, `delete()`
- Better type validation with descriptive error messages

### Other Changes
- Removed localStorage from momentum (settings now in-memory)
- Moved local dev DB to `servers/local/data.db`
- Added "apps are throw-away" philosophy to CLAUDE.md
- Created `scripts/release.sh` for fast local releases
- Updated `/fazt-release` skill to use the script

---

## Next Session

### Storage Debug Mode

Add observability to catch bugs faster. The v0.10.7 bugs were hard to diagnose
because queries silently returned empty results.

**Goal:**
```bash
FAZT_STORAGE_DEBUG=1 fazt server start ...
```

**When enabled, log:**
- Every storage operation with parameters
- Actual SQL queries generated
- Row counts and timing
- Warnings for reserved fields (`id`, `_createdAt`, `_updatedAt`)

**Files:**
- `internal/storage/bindings.go` - Add debug logging
- `internal/storage/ds_query.go` - Log generated SQL
- `internal/config/config.go` - Add debug flag

---

## Quick Reference

```bash
# Local server
fazt server start --port 8080 --domain 192.168.64.3 --db servers/local/data.db

# Release
source .env && ./scripts/release.sh vX.Y.Z

# Upgrade zyt
fazt remote upgrade zyt
```
