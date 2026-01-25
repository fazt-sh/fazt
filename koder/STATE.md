# Fazt Implementation State

**Last Updated**: 2026-01-25
**Current Version**: v0.10.13

## Status

State: CLEAN - Portable database with smart domain detection

---

## Last Session

**WebSocket Stress Tests & Portable Database Fix**

### 1. WebSocket Stress Tests (v0.10.11)

Added comprehensive stress tests in `internal/hosting/ws_stress_test.go`:

| Test | Result |
|------|--------|
| 1000 concurrent connections | ~123k/sec |
| Channel subscription throughput | ~183k/sec |
| Message fanout (500 subs) | <1µs avg |
| Memory usage | ~4KB/client |
| Rapid sub/unsub cycles | ~500k ops/sec |

Benchmarks:
- `BroadcastToChannel`: 3.5µs per broadcast to 100 subscribers
- `Subscribe`: 1.3µs per operation
- `GetSubscribers`: 1.6µs per lookup

### 2. Production Outage & Fix (v0.10.12 → v0.10.13)

**Problem**: v0.10.11 upgrade broke zyt.app (525 SSL error)

**Root cause**: Environment detection saw internal DigitalOcean IP (`10.46.0.5`)
instead of public IP, causing domain override to `10.46.0.5.nip.io`.

**v0.10.12** (hotfix): Skipped detection in production mode - worked but
required remembering env vars.

**v0.10.13** (proper fix): Smart domain detection based on domain type:

```
Real domains (zyt.app)     → Always trusted, never touched
Wildcard DNS (*.nip.io)    → Check IP, auto-update if different machine
IP addresses               → Check local, auto-update if not matching
Empty                      → Auto-detect local IP
```

**Files changed**:
- `cmd/server/main.go`: Simplified detection logic
- `internal/provision/detect.go`: Added `IsWildcardDNS`, `IsPortableDomain`
- `internal/provision/detect_test.go`: Tests for new functions

### 3. Key Design Decision

**Uniform Peers Philosophy**:
- No "dev" vs "production" distinction
- Every instance is a first-class peer
- Same binary works everywhere without env vars
- Database is portable - copy `data.db`, domain auto-adjusts

## Files Modified This Session

```
internal/hosting/ws_stress_test.go   # New - stress tests
internal/provision/detect.go         # Smart domain detection
internal/provision/detect_test.go    # New tests
cmd/server/main.go                   # Simplified detection logic
internal/config/config.go            # Version bump
CHANGELOG.md                         # Release notes
CLAUDE.md                            # Added Uniform Peers section
```

## Next Up

1. **Brainstorm /fazt-app ideas** showcasing storage + WebSockets
   - Collaborative apps (Yjs-style)
   - Real-time dashboards
   - Chat/presence demos

2. **Consider improving domain detection further**
   - Could fetch public IP from external service as fallback
   - But current "trust real domains" approach is simpler

---

## Quick Reference

```bash
# Rebuild and restart local server
go build -o ~/.local/bin/fazt ./cmd/server && \
  systemctl --user restart fazt-local

# Test WebSocket (websocat)
websocat ws://test.192.168.64.3:8080/_ws
{"type":"subscribe","channel":"chat"}

# Run stress tests
go test -v ./internal/hosting/... -run "TestStress"

# Run provision tests
go test -v ./internal/provision/...
```
