# Fazt Implementation State

**Last Updated**: 2026-01-16
**Current Version**: v0.9.21 (source), v0.9.18 (local binary), v0.9.21 (zyt)

## Status

```
State: ACTIVE ISSUE (paused)
Remote upgrade auto-restart not working. Binary replaces successfully but
service doesn't restart. Extensive investigation completed, root cause
identified but not yet fixed. Will revisit.
```

---

## ACTIVE ISSUE: Remote Upgrade Auto-Restart

### Goal

Remote upgrade via `fazt remote upgrade <peer>` should work without SSH:
1. Download new binary from GitHub releases
2. Replace current binary atomically
3. Restart the service automatically
4. New version running - no manual intervention

### Current Behavior

When running `fazt remote upgrade zyt`:
1. Binary downloads and replaces successfully ✓
2. API returns success: "Upgraded! Server is restarting..." ✓
3. **BUT service does NOT restart** - still runs old version ✗
4. Manual SSH required: `systemctl restart fazt`

### Root Cause Analysis

**The core issue**: Go's `exec.Command()` within an HTTP handler context
does not spawn truly independent processes. Even with shell detachment
tricks (nohup, setsid, &), the child process doesn't survive or execute
properly.

**Key observation**: The exact same shell command works perfectly when
run manually via SSH as the fazt user, but fails when called from Go's
exec.Command() in the upgrade handler.

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

## Untried Approaches

### 1. systemd-run (Most Promising)
Use `systemd-run` to spawn a transient systemd unit:
```go
exec.Command("systemd-run", "--no-block", "--",
    "/bin/systemctl", "restart", "fazt").Run()
```
This creates a completely independent unit managed by systemd.

### 2. Helper Script
Write a restart script and execute that:
```bash
# /usr/local/bin/fazt-restart.sh
#!/bin/bash
sleep 1
sudo /bin/systemctl restart fazt
```
Then: `exec.Command("sh", "-c", "/usr/local/bin/fazt-restart.sh &").Run()`

### 3. at Scheduler
Use the `at` command to schedule restart:
```go
exec.Command("sh", "-c", "echo 'sudo /bin/systemctl restart fazt' | at now + 1 minute").Run()
```
Completely decouples the restart from the current process.

### 4. syscall.ForkExec with Setsid
Use low-level Go syscall with proper session creation:
```go
syscall.ForkExec("/bin/sh", []string{"sh", "-c", "..."}, &syscall.ProcAttr{
    Sys: &syscall.SysProcAttr{Setsid: true},
})
```

### 5. Self-Termination + Systemd Restart
Instead of calling systemctl, just exit the process and let systemd's
`Restart=always` handle it:
```go
os.Exit(0)  // Systemd will restart us with new binary
```

---

## Recent Releases

| Version | Date | Summary |
|---------|------|---------|
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

## Next Steps (When Revisiting)

1. **Try systemd-run approach** - most likely to work
2. **Or try self-termination** - let Restart=always handle it
3. **Test v0.9.21 debug logging** - see what's actually happening
4. **Update install.sh** once fix is confirmed
