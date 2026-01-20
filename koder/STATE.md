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

### Debug Mode

Add `FAZT_DEBUG=1` for development observability. Enable by default for local server.

**Why:** v0.10.7 bugs were hard to diagnose - queries silently returned empty results.

**When enabled, log:**
- Storage: operations, SQL queries, row counts, timing
- Runtime: request IDs, execution timing, VM pool state
- Warnings for common mistakes (reserved fields, type mismatches)

**Implementation:**
1. Add `FAZT_DEBUG` env var check in `internal/config/config.go`
2. Add debug logging to storage (`internal/storage/`)
3. Add debug logging to runtime (`internal/runtime/`)
4. Local server (`Environment=development`) enables debug by default

---

## Quick Reference

```bash
# Local server (debug on by default in dev mode)
fazt server start --port 8080 --domain 192.168.64.3 --db servers/local/data.db

# Force debug in any environment
FAZT_DEBUG=1 fazt server start ...

# Release
source .env && ./scripts/release.sh vX.Y.Z

# Upgrade zyt
fazt remote upgrade zyt
```
