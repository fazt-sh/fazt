# Fazt Implementation State

**Last Updated**: 2026-01-16
**Current Version**: v0.9.8 (local), v0.9.7 (zyt)

## Status

```
State: ACTIVE ISSUE
Remote upgrade auto-restart not working reliably. Binary replaces but service
doesn't restart. Requires investigation and fix.
```

---

## ACTIVE ISSUE: Remote Upgrade Auto-Restart

### Goal

1. Fresh droplet install should work without issues
2. Remote upgrade via `fazt remote upgrade <peer>` should work without SSH

### Current Behavior

When running `fazt remote upgrade zyt`:
1. Binary downloads and replaces successfully
2. API returns success: "Upgraded! Server is restarting..."
3. **BUT service does NOT restart** - still runs old version
4. Manual SSH required: `systemctl restart fazt`

### What's Been Tried

#### v0.9.6: Sudoers + Directory Permissions
- Added `/etc/sudoers.d/fazt` with NOPASSWD rules for systemctl
- Added group write permission on `/usr/local/bin`
- Changed upgrade handler to use `sudo systemctl restart fazt`
- **Result**: Sudoers works when tested manually, but auto-restart still fails

#### v0.9.8: Added 500ms Delay
- Added `time.Sleep(500 * time.Millisecond)` before restart command
- Theory: response might not flush before restart
- **Result**: Not yet tested (CI building)

### Technical Details

**Upgrade Handler** (`internal/handlers/upgrade_handler.go:160-167`):
```go
go func() {
    time.Sleep(500 * time.Millisecond)
    exec.Command("sudo", "systemctl", "restart", "fazt").Run()
}()
```

**Sudoers Rule** (`/etc/sudoers.d/fazt`):
```
fazt ALL=(ALL) NOPASSWD: /bin/systemctl restart fazt, /bin/systemctl start fazt, /bin/systemctl stop fazt
```

**Directory Permissions**:
```bash
# Binary owned by fazt user (enables atomic replacement)
chown fazt:fazt /usr/local/bin/fazt

# Directory has group write (enables staging new binary)
chgrp fazt /usr/local/bin
chmod g+w /usr/local/bin
```

**Service File** (`/etc/systemd/system/fazt.service`):
```ini
[Service]
Type=simple
User=fazt
ProtectSystem=strict
ReadWritePaths=/usr/local/bin
```

### Verified Working

- `su - fazt -c 'sudo -n /bin/systemctl restart fazt'` - WORKS
- Binary replacement via atomic rename - WORKS
- Sudoers rule allows passwordless restart - WORKS
- Directory permissions allow staging - WORKS

### Unknown/Suspect

1. **Goroutine execution**: Does the goroutine actually run before process exits?
2. **exec.Command error handling**: Errors are silently ignored
3. **PATH issues**: `systemctl` vs `/bin/systemctl` resolution
4. **Timing**: Is 500ms enough? Too much?

### Files to Investigate

| File | Purpose |
|------|---------|
| `internal/handlers/upgrade_handler.go` | Upgrade API handler |
| `install.sh` | Fresh install script |
| `internal/provision/manager.go` | `fazt service install` logic |

### Testing Plan

1. Wait for v0.9.8 CI to complete
2. Run `fazt remote upgrade zyt`
3. Check if version updates without manual restart
4. If still fails:
   - SSH in and check journalctl for errors
   - Add logging to upgrade handler
   - Test with longer delay
   - Consider synchronous restart (block until done)

### Alternative Approaches to Consider

1. **Synchronous restart with timeout**: Don't return until restart confirmed
2. **External restarter**: Spawn separate process to do the restart
3. **Systemd WatchdogSec**: Let systemd handle restart on exit
4. **Self-exec**: Binary re-execs itself instead of systemctl

---

## Recent Releases

| Version | Date | Summary |
|---------|------|---------|
| v0.9.8 | 2026-01-16 | Upgrade restart delay (testing) |
| v0.9.7 | 2026-01-16 | API routing fix for storage |
| v0.9.6 | 2026-01-16 | Sudoers + directory permissions |
| v0.9.5 | 2026-01-16 | Storage primitives + stdlib |

## Apps on zyt.app

| App | URL | Features |
|-----|-----|----------|
| home | https://zyt.app | Editorial homepage |
| tetris | https://tetris.zyt.app | 3D game with high score API |
| snake | https://snake.zyt.app | Snake game with high score API |

**High Score APIs**: Now working after v0.9.7 routing fix.

---

## Quick Reference

| Command | Purpose |
|---------|---------|
| `fazt remote status zyt` | Check health/version |
| `fazt remote upgrade zyt` | Upgrade server |
| `fazt remote deploy <dir> zyt` | Deploy app |
| `ssh root@165.227.11.46` | Direct SSH to zyt |

## Next Steps

1. Test v0.9.8 upgrade
2. If fails, add proper error logging to upgrade handler
3. Consider alternative restart mechanisms
4. Update install.sh to match whatever fix works
