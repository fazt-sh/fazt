# Fazt Implementation State

**Last Updated**: 2026-01-19
**Current Version**: v0.10.1 (local = zyt = source ✓)

## Status

```
State: CLEAN
v0.10 infrastructure complete. /fazt-app skill needs update to use it.
```

---

## Context: Why v0.10?

The v0.10 changes support the **`/fazt-app` workflow** - enabling Claude to:

1. **Build** apps using Vite (or any build tool)
2. **Deploy locally** for rapid iteration
3. **Test serverless** via `/_fazt/*` endpoints (storage, snapshots, logs)
4. **Get local URL** to verify behavior
5. **Deploy to remote** once everything works

### The Problem Before v0.10

- Testing serverless locally was hard/broken
- No way for Claude to introspect app state during testing
- Jumping straight to remote deployment = slow iteration

### What v0.10 Provides

| Feature | Purpose |
|---------|---------|
| `/_fazt/info` | Claude checks app metadata |
| `/_fazt/storage` | Inspect KV state during testing |
| `/_fazt/snapshot` | Save state before experiments |
| `/_fazt/restore` | Reset to known state |
| `/_fazt/logs` | Debug serverless execution |
| `/_fazt/errors` | Find issues quickly |
| `@peer` syntax | `fazt @local ...` vs `fazt @zyt ...` |

---

## Next Up: Update /fazt-app Skill + Fix Othelo

### Problem 1: /fazt-app is Outdated

The skill at `.claude/commands/fazt-app.md` doesn't reflect v0.10:

- Jumps straight to `fazt app deploy --to zyt` (no local testing)
- Doesn't mention `/_fazt/*` endpoints
- Doesn't show local development workflow
- Uses `--to`/`--from` flags (should prefer `@peer` where appropriate)

**Fix**: Update the skill to use the local-first workflow:
```
1. Build app
2. Deploy to local: fazt app deploy ./app --to local
3. Test at http://app.192.168.64.3.nip.io:8080
4. Use /_fazt/* endpoints to debug
5. When working: fazt app deploy ./app --to zyt
```

### Problem 2: Othelo Not Working Locally

The othelo app at `servers/zyt/othelo/` doesn't work when deployed locally.
This is a good test case for the updated workflow.

**Goal**: Use updated `/fazt-app` to fix othelo locally, then deploy to zyt.

### Success Criteria

- [ ] `/fazt-app` skill updated with local-first workflow
- [ ] `/fazt-app` documents `/_fazt/*` endpoints for testing
- [ ] Othelo works locally at http://othelo.192.168.64.3.nip.io:8080
- [ ] Othelo deployed and working at https://othelo.zyt.app

---

## Quick Reference

```bash
# Start local server
fazt server start --domain 192.168.64.3 --port 8080 --db /tmp/fazt-local.db

# Deploy locally
fazt app deploy ./myapp --to local
# Test at http://myapp.192.168.64.3.nip.io:8080

# Agent endpoints (Claude uses these for testing)
curl http://myapp.192.168.64.3.nip.io:8080/_fazt/info
curl http://myapp.192.168.64.3.nip.io:8080/_fazt/storage
curl http://myapp.192.168.64.3.nip.io:8080/_fazt/logs

# Deploy to production when working
fazt app deploy ./myapp --to zyt

# @peer syntax for remote commands
fazt @local app list
fazt @zyt app list
```

---

## Apps on zyt.app

| App | ID | Status | URL |
|-----|----|----|-----|
| tetris | app_98ed8539 | ✓ | https://tetris.zyt.app |
| pomodoro | app_9c479319 | ✓ | https://pomodoro.zyt.app |
| **othelo** | app_706bd43c | **broken locally** | https://othelo.zyt.app |
| snake | app_1d28078d | ✓ | https://snake.zyt.app |
