# Fazt Implementation State

**Last Updated**: 2026-01-16
**Current Version**: v0.9.22 (source, pending release)

## Status

```
State: FIX IMPLEMENTED - Ready for release and testing
Remote upgrade auto-restart fix applied using systemd-run solution from
kitten project. Needs version bump, release, and verification on zyt.
```

---

## RESOLVED: Remote Upgrade Auto-Restart

### Goal

Remote upgrade via `fazt remote upgrade <peer>` should work without SSH:
1. Download new binary from GitHub releases
2. Replace current binary atomically
3. Restart the service automatically
4. New version running - no manual intervention

### Solution (v0.9.22)

**Root cause**: Go's `exec.Command()` within systemd service context doesn't
spawn truly independent processes. Even with nohup/setsid/&, child processes
remain in the same cgroup and are killed when the parent exits.

**Fix**: Use `systemd-run --scope` to spawn a transient scope unit that
exists outside our cgroup, then `os.Exit(0)` to let systemd's `Restart=always`
bring us back with the new binary.

```go
go func() {
    time.Sleep(100 * time.Millisecond) // Allow response to transmit
    exec.Command("systemd-run", "--scope", "--",
        "sudo", "/bin/systemctl", "restart", "fazt").Start()
    os.Exit(0)
}()
```

**Also fixed**: Removed `setcap` call which loses capabilities on binary
replacement. `AmbientCapabilities` in systemd unit handles this correctly.

Solution tested and verified in kitten project (`~/Projects/kitten/koder/SOLUTION.md`).

---

## What Works (Verified)

| Component | Status | Notes |
|-----------|--------|-------|
| Binary download | ✓ | Downloads from GitHub releases |
| Atomic replacement | ✓ | Uses rename() for safe swap |
| Sudoers rule | ✓ | Passwordless systemctl for fazt user |
| Manual restart | ✓ | `sudo -u fazt sudo /bin/systemctl restart fazt` |
| Shell command | ✓ | Works when run from SSH session |

**Manual test that works**:
```bash
sudo -u fazt sh -c "nohup sh -c 'sleep 1 && sudo /bin/systemctl restart fazt' >/dev/null 2>&1 &"
```

---

## Attempted Solutions (v0.9.8 - v0.9.21)

### v0.9.8: Goroutine with Delay
```go
go func() {
    time.Sleep(500 * time.Millisecond)
    exec.Command("sudo", "systemctl", "restart", "fazt").Run()
}()
```
**Result**: Failed. Goroutine never executed before handler returned.

### v0.9.9: Added setsid for Process Detachment
```go
cmd := exec.Command("setsid", "sudo", "systemctl", "restart", "fazt")
cmd.Start()
```
**Result**: Failed. Process still didn't execute.

### v0.9.10: Shell with nohup + setsid + &
```go
exec.Command("sh", "-c", "nohup setsid sudo systemctl restart fazt >/dev/null 2>&1 &").Run()
```
**Result**: Failed.

### v0.9.12: External Sleep in Shell
```go
exec.Command("sh", "-c", "sleep 1 && sudo systemctl restart fazt").Start()
```
**Result**: Failed. Shell process didn't survive.

### v0.9.13-14: Debug Logging Discovery
Added file logging to trace execution. Discovered:
- `PrivateTmp=true` in systemd service was isolating /tmp
- Debug logs were being written to private /tmp namespace
- Fixed by adding `/tmp` to `ReadWritePaths`

**Service file change**:
```ini
ReadWritePaths=/usr/local/bin /tmp /home/fazt
PrivateTmp=false
```

### v0.9.15: Full Path for Sudoers Match
Sudoers rule specifies `/bin/systemctl` but code was calling `systemctl`.
```go
exec.Command("sh", "-c", "sleep 1 && sudo /bin/systemctl restart fazt").Start()
```
**Result**: Failed. Path wasn't the issue.

### v0.9.17-18: Combined nohup + & + .Run()
```go
exec.Command("sh", "-c", "nohup sh -c 'sleep 1 && sudo /bin/systemctl restart fazt' >/dev/null 2>&1 &").Run()
```
**Result**: Failed. The outer shell completes but inner command doesn't run.

### v0.9.21: Added Comprehensive Debug Logging
```go
restartCmd := "echo 'restart-start' >> /tmp/fazt-restart.log; " +
    "nohup sh -c 'sleep 1 && echo restart-exec >> /tmp/fazt-restart.log && " +
    "sudo /bin/systemctl restart fazt' >/dev/null 2>&1 &; " +
    "echo 'restart-scheduled' >> /tmp/fazt-restart.log"
exec.Command("sh", "-c", restartCmd).Run()
```
**Status**: Deployed, awaiting test.

---

## Technical Details

### Current Code Location
`internal/handlers/upgrade_handler.go:160-167`

### Sudoers Rule (`/etc/sudoers.d/fazt`)
```
fazt ALL=(ALL) NOPASSWD: /bin/systemctl restart fazt, /bin/systemctl start fazt, /bin/systemctl stop fazt
```

### Service File (`/etc/systemd/system/fazt.service`)
```ini
[Unit]
Description=Fazt PaaS
After=network.target

[Service]
Type=simple
User=fazt
WorkingDirectory=/home/fazt/.config/fazt
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
ExecStart=/usr/local/bin/fazt server start
Restart=always
LimitNOFILE=4096
Environment=FAZT_ENV=production
ProtectSystem=strict
ReadWritePaths=/usr/local/bin /tmp /home/fazt
PrivateTmp=false

[Install]
WantedBy=multi-user.target
```

---

## Alternative Approaches (Not Needed)

These were considered before the systemd-run solution was found:

| Approach | Why Not Used |
|----------|--------------|
| Helper script | Adds deployment complexity |
| `at` scheduler | Requires at daemon, 1min minimum delay |
| syscall.ForkExec | Complex, still doesn't escape cgroup |
| Self-termination only | Works but less explicit than systemd-run |

---

## Recent Releases

| Version | Date | Summary |
|---------|------|---------|
| v0.9.22 | 2026-01-16 | **FIX**: systemd-run + os.Exit for restart |
| v0.9.21 | 2026-01-16 | Debug logging for restart tracing |
| v0.9.20 | 2026-01-16 | Test version for upgrade flow |
| v0.9.19 | 2026-01-16 | Test version |
| v0.9.18 | 2026-01-16 | nohup + & approach |
| v0.9.17 | 2026-01-16 | nohup detachment |
| v0.9.15 | 2026-01-16 | Full path /bin/systemctl |
| v0.9.13-14 | 2026-01-16 | PrivateTmp discovery |
| v0.9.12 | 2026-01-16 | External sleep approach |
| v0.9.10 | 2026-01-16 | Shell background approach |
| v0.9.9 | 2026-01-16 | setsid detachment |
| v0.9.8 | 2026-01-16 | Goroutine delay approach |
| v0.9.7 | 2026-01-16 | API routing fix for storage |
| v0.9.6 | 2026-01-16 | Sudoers + directory permissions |
| v0.9.5 | 2026-01-16 | Storage primitives + stdlib |

---

## Apps on zyt.app

| App | URL | Features |
|-----|-----|----------|
| home | https://zyt.app | Editorial homepage |
| tetris | https://tetris.zyt.app | 3D game with high score API |
| snake | https://snake.zyt.app | Snake game with high score API |

**High Score APIs**: Working after v0.9.7 routing fix.

---

## Quick Reference

| Command | Purpose |
|---------|---------|
| `fazt remote status zyt` | Check health/version |
| `fazt remote upgrade zyt` | Upgrade server (restart broken) |
| `fazt remote deploy <dir> zyt` | Deploy app |

### SSH Access (Testing Only)

```bash
ssh root@165.227.11.46   # Direct IP - domain proxied via Cloudflare
```

**IMPORTANT**: SSH is for debugging only. Goal is to eliminate SSH requirement.

---

## Next Steps

1. **Bump version to 0.9.22** in `internal/config/config.go`
2. **Release v0.9.22** via `/fazt-release` or manual workflow
3. **Test upgrade on zyt** - `fazt remote upgrade zyt`
4. **Verify auto-restart works** - service should restart without SSH
5. **Clean up debug logging** from previous versions if desired
