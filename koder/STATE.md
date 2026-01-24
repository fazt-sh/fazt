# Fazt Implementation State

**Last Updated**: 2026-01-25
**Current Version**: v0.10.10

## Status

State: CLEAN - Unified install script complete, local server persistent

---

## Last Session

**Persistent Local Server + Unified Install Script**

1. **Created persistent local server** (Ticket pending)
   - User systemd service: `~/.config/systemd/user/fazt-local.service`
   - Auto-starts on boot via linger
   - No sudo required for local development

2. **Updated `install.sh`** with three options:
   - Production Server (system service, real domain, HTTPS)
   - Local Development (user service, auto-start, IP-based)
   - CLI Only (just binary, connect to remotes)

3. **Added environment detection** (`internal/provision/detect.go`):
   - `GetLocalIPs()` - Get all IPs of current machine
   - `DetectEnvironment()` - Check if stored domain matches machine
   - Server auto-detects and warns on domain mismatch

4. **Extended systemd support** (`internal/provision/systemd.go`):
   - `InstallUserService()` - Create user-level services
   - `EnableLinger()` - Persist services across reboots
   - `UserSystemctl()` - Manage user services

5. **Updated `/fazt-start` skill**:
   - Checks both local and zyt server health
   - Shows version matrix (source/binary/local/zyt)
   - Quick commands for common operations

## Architecture

**Install locations**:

| Type | Binary | Service | DB |
|------|--------|---------|-----|
| Production | `/usr/local/bin/fazt` | System | `~/.config/fazt/data.db` |
| Local Dev | `~/.local/bin/fazt` | User | `~/.config/fazt/data.db` |

**Domain detection on startup**:
- If CLI `--domain` flag → use it (highest priority)
- If domain in DB matches current machine's IP → use it
- If mismatch → warn and use detected local IP

## Next Up

1. **Wire WebSocket endpoint** - Foundation exists in `internal/hosting/ws.go`
   - Add `/_ws` route
   - Add channels/pub-sub
   - Add `fazt.realtime` to serverless runtime

2. **Build stress-test app** - After WebSockets working
   - "Antfarm" browser visualization
   - "Siege" Go CLI for real load
   - Live dashboard with polling/WebSocket updates

3. **Consider**: Agent-interface idea in `koder/scratch.md`

---

## Quick Reference

```bash
# Local server management
systemctl --user status fazt-local
systemctl --user restart fazt-local
journalctl --user -u fazt-local -f

# Rebuild and restart
go build -o ~/.local/bin/fazt ./cmd/server && \
  systemctl --user restart fazt-local

# Run install script
./install.sh  # Select option 2 for local dev
```
