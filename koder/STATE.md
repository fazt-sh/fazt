# Fazt Implementation State

**Last Updated**: 2026-01-27
**Current Version**: v0.10.14

## Status

State: CLEAN - Worker idle timeout implemented, NEXUS uses workers for simulation

---

## Last Session

**Worker Idle Timeout + NEXUS Integration**

1. **Added `fazt.realtime` to workers** (f-73d6)
   - Workers can now broadcast to WebSocket clients
   - Enables real-time data flow from background workers

2. **Worker Idle Timeout** (f-361c)
   - New JobConfig options: `idleTimeout`, `idleChannel`
   - Worker auto-stops when no WebSocket listeners for specified duration
   - Clean stop (not failure, no daemon restart)
   - Resource-efficient: no wasted CPU/memory when nobody is watching

3. **NEXUS Mall Simulator** (f-2ebf)
   - Replaced external Bun process with fazt worker
   - `workers/mall-sim.js` generates data, broadcasts via WebSocket
   - Uses idle timeout to stop when dashboard closes
   - Removed external `server/` directory and `nexus-simulate.ts`

### API

```javascript
// Worker with idle timeout
fazt.worker.spawn('workers/sim.js', {
    daemon: true,
    idleTimeout: '1m',   // Stop if no listeners for 1 minute
    idleChannel: 'mall'  // Monitor this WebSocket channel
});
```

### Files Modified

| File | Change |
|------|--------|
| `internal/worker/job.go` | Added IdleTimeout, IdleChannel to JobConfig |
| `internal/worker/bindings.go` | Parse idleTimeout/idleChannel from JS |
| `internal/worker/pool.go` | Added watchIdleTimeout goroutine |
| `internal/worker/init.go` | Added SetListenerCountFunc |
| `cmd/server/main.go` | Wired up listener count function |

---

## Previous Session

**Worker System Release** (v0.10.14)

Released background worker system and added UX improvements.

---

## Next Up

1. **Test NEXUS with Idle Timeout**
   - Open dashboard, start simulation
   - Close all browser tabs
   - Verify worker stops after 1 minute

2. **Deploy to zyt.app** (optional)

---

## Quick Reference

```bash
# Build and install
go build -o ~/.local/bin/fazt ./cmd/server

# Run tests
go test ./...

# Local server logs
journalctl --user -u fazt-local -f

# Deploy to local
fazt app deploy <dir> --to local

# NEXUS simulation
curl -X POST http://nexus.192.168.64.3.nip.io:8080/api/simulate/start
curl http://nexus.192.168.64.3.nip.io:8080/api/simulate/status
curl -X POST http://nexus.192.168.64.3.nip.io:8080/api/simulate/stop
```

---

## Backlog

### NEXUS Dashboard Refinement

- UI/UX polish
- More widgets
- Mobile responsiveness
- Widget library expansion

Current state: Mall simulation works via fazt workers with real-time WebSocket
updates and idle timeout for resource efficiency.
