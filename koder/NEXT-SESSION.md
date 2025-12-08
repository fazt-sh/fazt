# Next Session Handoff: Server Crash Investigation

**Date**: December 8, 2025
**Status**: 🔴 **CRITICAL REGRESSION**: Server process exits immediately after startup.

## 🛑 Immediate Action Required
1.  **Debug Server Crash**: The server binary (`./fazt server start`) starts, logs "Server starting on :8080", and then the process disappears immediately.
    *   **Symptoms**: `server.log` looks normal (ends at "Dashboard: https://fazt.sh"). Process is not in `ps aux`. Port 8080 is closed.
    *   **Suspects**: Recent changes in `internal/middleware/auth.go` or `internal/analytics/buffer.go` (Safeguards Phase 1).
    *   **Action**: Run `go run ./cmd/server server start --port 8080` in foreground to catch panics/exit codes directly.
2.  **Verify Client**: Once server is stable, run `./fazt client sites`. It *should* work (Authentication fix applied in `internal/middleware/auth.go`).

## ⚠️ Known Issues
*   **Server Stability**: The `fazt` binary is unstable. It was verified earlier in the session but now fails to stay persistent.
*   **Client Connection**: Fails with `dial tcp [::1]:8080: connect: connection refused` (due to server being down).

## ✅ Accomplished (Previous Session)
1.  **Safeguards Phase 1 (Write Optimization)**:
    *   **Buffered Stats**: Events (`/pixel.gif`, `/track`) are now buffered in RAM.
    *   **Code**: `internal/analytics/buffer.go`.

2.  **Safeguards Phase 2 (Resource Awareness)**:
    *   **Probe**: `internal/system/probe.go` detects Host/Cgroup RAM limits.

3.  **New Admin API (v0.7.1)**:
    *   **Foundation**: `internal/api/response.go` (Envelope).
    *   **Refactor**: `SitesHandler` (`/api/sites`) updated to use Envelope.

4.  **Authentication Fix (Code Only)**:
    *   **Middleware**: Updated `internal/middleware/auth.go` to accept `Authorization: Bearer <token>`.

## 📋 Next Steps (The Plan)

### 1. Fix Server Crash
*   Identify why `fazt` is exiting. Check for silent panics or logic errors in the new `Analytics` buffer shutdown or initialization.

### 2. Verify & Close Phase 1
*   Once `fazt client sites` works, the Auth/API refactor is validated.

### 3. Safeguards Phase 3 (VFS Modernization)
*   **Goal**: Protect Admin Panel from OOM.
*   **Task**: Update `internal/hosting/vfs.go` to implement Byte-Weighted LRU.

### 4. Admin SPA
*   Frontend team can now build against the new API (`/api/system/*`, `/api/sites/{id}`).
