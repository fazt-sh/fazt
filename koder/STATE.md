# Fazt Implementation State

**Last Updated**: 2026-01-19
**Current Version**: v0.10.1 (local = zyt = source ✓)

## Status

```
State: CLEAN
Next: Update /fazt-app skill to use v0.10 local-first workflow.
```

---

## Next Up: Update /fazt-app Skill

The v0.10 infrastructure is complete but `/fazt-app` skill doesn't use it yet.

### Current Skill Problems

File: `.claude/commands/fazt-app.md`

1. **Deploys directly to zyt** - skips local testing entirely
2. **No `/_fazt/*` endpoints** - Claude can't introspect app state
3. **No local workflow** - can't test serverless before deploying

### CLI Syntax Question

The `--to`/`--from` flags are redundant with `@peer`. Consider:

```bash
# Current
fazt app deploy ./myapp --to local

# Cleaner option 1: @peer suffix
fazt app deploy ./myapp @local

# Cleaner option 2: positional (like 'list' already does)
fazt app deploy ./myapp local
```

Note: `@peer` at START = remote execution (`fazt @zyt app list`).
For deploy, files transfer FROM local, so it's different from remote exec.

### What the Skill Should Do

```
1. Build app (Vite or zero-build)
2. Start local server if not running
3. Deploy to local: fazt app deploy ./app --to local
4. Get local URL: http://app.192.168.64.3.nip.io:8080
5. Test & iterate using /_fazt/* endpoints
6. When working: fazt app deploy ./app --to zyt
```

### New Endpoints for Testing (v0.10)

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/_fazt/info` | GET | App metadata, file count, storage keys |
| `/_fazt/storage` | GET | List all KV keys with sizes |
| `/_fazt/storage/:key` | GET | Get specific KV value |
| `/_fazt/snapshot` | POST | Save current state with name |
| `/_fazt/restore/:name` | POST | Restore to named snapshot |
| `/_fazt/snapshots` | GET | List available snapshots |
| `/_fazt/logs` | GET | Recent serverless logs |
| `/_fazt/errors` | GET | Recent errors only |

These let Claude debug serverless without guessing.

### Test Case: Othelo App

`servers/zyt/othelo/` - broken locally, good test for the updated workflow.

---

## Quick Reference

```bash
# Local development
fazt server start --domain 192.168.64.3 --port 8080 --db /tmp/fazt-local.db
fazt app deploy ./myapp --to local
# http://myapp.192.168.64.3.nip.io:8080

# Test endpoints
curl http://myapp.192.168.64.3.nip.io:8080/_fazt/info
curl http://myapp.192.168.64.3.nip.io:8080/_fazt/storage
curl http://myapp.192.168.64.3.nip.io:8080/_fazt/logs

# Deploy to production
fazt app deploy ./myapp --to zyt
```

---

## v0.10 Implementation (Complete)

| Component | File |
|-----------|------|
| Agent Endpoints | `internal/handlers/agent_handler.go` |
| Aliases | `internal/handlers/aliases_handler.go` |
| Command Gateway | `internal/handlers/cmd_gateway.go` |
| CLI v2 | `cmd/server/app_v2.go` |
| Release Script | `scripts/release.sh` |
