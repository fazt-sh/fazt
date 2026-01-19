# Fazt Implementation State

**Last Updated**: 2026-01-19
**Current Version**: v0.10.1 (local = zyt = source ✓)

## Status

```
State: CLEAN
v0.10 App Identity implemented. Next: refine /fazt-app workflow.
```

---

## Context: Why v0.10?

The v0.10 changes support the **`/fazt-app` workflow** - enabling Claude to:

1. **Build** apps using Vite (or any build tool)
2. **Deploy locally** for rapid iteration
3. **Test serverless** via `/_fazt/*` endpoints (storage, snapshots, logs)
4. **Get local URL** to verify behavior
5. **Deploy to remote** once everything works

The app identity model (`app_xxxxxxxx`) and alias system make this cleaner:
- Apps have stable IDs across local/remote
- Aliases handle routing without ID conflicts
- `@peer` execution lets Claude run commands on any peer
- Agent endpoints provide introspection for testing

---

## Completed: v0.10 - App Identity & Agent Endpoints

**Released**: v0.10.0 + v0.10.1 (bug fix)

### For /fazt-app Workflow

| Feature | Purpose in Workflow |
|---------|---------------------|
| `/_fazt/info` | Claude checks app metadata |
| `/_fazt/storage` | Inspect KV state during testing |
| `/_fazt/snapshot` | Save state before experiments |
| `/_fazt/restore` | Reset to known state |
| `/_fazt/logs` | Debug serverless execution |
| `/_fazt/errors` | Find issues quickly |
| `@peer` execution | `fazt @local app list` vs `fazt @zyt app list` |
| Aliases | Same app name locally and remotely |

### Key Files

| Component | File |
|-----------|------|
| Agent Endpoints | `internal/handlers/agent_handler.go` |
| Aliases | `internal/handlers/aliases_handler.go` |
| Command Gateway | `internal/handlers/cmd_gateway.go` |
| CLI v2 | `cmd/server/app_v2.go` |

---

## Next Up: Fix Othelo App

**Goal**: Use `/fazt-app` to fix the othelo app - a real test of the v0.10 workflow.

### The Problem

The othelo app at https://othelo.zyt.app is broken/incomplete. This is a perfect
candidate to test the new Claude-driven development workflow:

1. Pull the app locally: `fazt app pull othelo --from zyt --to ./servers/zyt/othelo`
2. Run `/fazt-app` to have Claude fix/rebuild it
3. Deploy locally, test with `/_fazt/*` endpoints
4. Deploy to zyt when working

### Success Criteria

- Othelo game is playable at https://othelo.zyt.app
- `/fazt-app` workflow proved out end-to-end
- Any workflow gaps identified and noted

---

## Quick Reference

```bash
# Local development
fazt server start --domain 192.168.64.3 --port 8080 --db /tmp/fazt-local.db
fazt app deploy ./myapp --to local
# Test at http://myapp.192.168.64.3.nip.io:8080

# Agent endpoints (for Claude testing)
curl http://myapp.192.168.64.3.nip.io:8080/_fazt/info
curl http://myapp.192.168.64.3.nip.io:8080/_fazt/storage
curl http://myapp.192.168.64.3.nip.io:8080/_fazt/logs

# Deploy to production when ready
fazt app deploy ./myapp --to zyt

# Release new fazt version
./scripts/release.sh v0.x.y "Title"
```

---

## Apps on zyt.app

| App | ID | Status | URL |
|-----|----|----|-----|
| tetris | app_98ed8539 | ✓ | https://tetris.zyt.app |
| pomodoro | app_9c479319 | ✓ | https://pomodoro.zyt.app |
| **othelo** | app_706bd43c | **broken** | https://othelo.zyt.app |
| snake | app_1d28078d | ✓ | https://snake.zyt.app |
