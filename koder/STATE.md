# Fazt Implementation State

**Last Updated**: 2026-01-24
**Current Version**: v0.10.10

## Status

State: BLOCKED - Storage layer needs fixes before further app development

---

## Active Plan

**Plan 21: Storage Layer Performance & Concurrency Fixes**
See `koder/plans/21_storage_layer_fixes.md`

### Problem

Stress testing revealed storage layer can't handle concurrent load:

| Test | Success Rate | Error |
|------|--------------|-------|
| Sequential writes (20) | 70% | TimeoutError |
| Concurrent writes (10) | 60% | SQLITE_BUSY |
| Sequential reads (50) | 48% | TimeoutError |

### Root Causes

1. **Context not propagated** - `bindings.go` uses `context.Background()`
2. **Timeout race** - Runtime timeout (5s) = SQLite busy_timeout (5s)
3. **No retry logic** - SQLITE_BUSY treated as hard failure
4. **Missing indexes** - JSON session queries do full table scans
5. **Pool too large** - 25 connections causes contention

### Fixes Required

| Priority | Fix | File |
|----------|-----|------|
| P0 | Pass context to storage ops | `internal/storage/bindings.go` |
| P0 | Reduce busy_timeout to 2s | `internal/database/db.go` |
| P1 | Add retry + backoff for SQLITE_BUSY | `internal/storage/storage.go` |
| P1 | Reduce MaxOpenConns to 10 | `internal/database/db.go` |
| P2 | Add session_id index | New migration |

---

## Next Session

1. Implement Plan 21 fixes
2. Re-run stress tests (target: 95%+ success)
3. Then continue with app development

---

## Quick Reference

```bash
# Deploy app
fazt app deploy servers/zyt/cashflow --to zyt

# Force restart
curl -X POST "https://admin.zyt.app/api/upgrade?force=true" \
  -H "Authorization: Bearer $TOKEN"

# Release
source .env && ./scripts/release.sh vX.Y.Z
fazt remote upgrade zyt

# Stress test storage
curl -s "https://cashflow.zyt.app/api/categories?session=test"
```
