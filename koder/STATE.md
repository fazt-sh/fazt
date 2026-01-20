# Fazt Implementation State

**Last Updated**: 2026-01-21
**Current Version**: v0.10.8

## Status

State: CLEAN

---

## Last Session (2026-01-21)

### Release: v0.10.8 - Debug Mode

Added `FAZT_DEBUG=1` environment variable for development observability.

**Features:**
- Debug mode enabled by default in development mode
- Storage operations logged with timing and row counts
- Runtime requests logged with request IDs for tracing
- VM pool state monitoring
- Warnings for common mistakes (e.g., setting `id` in insert)

**Implementation:**
- `internal/debug/debug.go` - New debug package with logging helpers
- `internal/config/config.go` - Added `DebugMode()` method
- `internal/storage/bindings.go` - Added debug logging to ds.* operations
- `internal/runtime/handler.go` - Added request ID generation and logging
- `internal/runtime/runtime.go` - Added VM pool state logging

**Usage:**
```bash
# Explicit enable
FAZT_DEBUG=1 fazt server start ...

# Automatic in dev mode (ENV=development)
fazt server start --db servers/local/data.db ...
```

**Sample output:**
```
[DEBUG] Debug mode enabled
[DEBUG runtime] req=a1b2c3 app=myapp path=/api/tasks method=GET started
[DEBUG storage] find myapp/tasks query={} rows=5 took=1.2ms
[DEBUG runtime] req=a1b2c3 app=myapp path=/api/tasks status=200 took=3.5ms
```

---

## Next Session

No specific task queued. Potential areas:
- More debug coverage (KV, S3 operations)
- SQL query logging option
- Performance profiling tools

---

## Quick Reference

```bash
# Local server (debug on by default in dev mode)
fazt server start --port 8080 --domain 192.168.64.3 --db servers/local/data.db

# Force debug in any environment
FAZT_DEBUG=1 fazt server start ...

# Disable debug in dev mode
FAZT_DEBUG=0 fazt server start ...

# Release
source .env && ./scripts/release.sh vX.Y.Z

# Upgrade zyt
fazt remote upgrade zyt
```
