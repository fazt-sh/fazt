# Fazt

**Sovereign compute** - Single Go binary + SQLite database that runs anywhere.

**Version**: 0.13.0 | **State**: `koder/STATE.md`

## Philosophy

- **No backward compatibility** - Rapidly iterating, single user. Break and evolve.
- **Static hosting first** - Serverless is enhancement, never blocks static deploy.
- **Apps are throw-away** - Fix fazt, not workarounds. Apps test fazt.

## Environment

| What | Value |
|------|-------|
| VM | `192.168.64.3` (headless Ubuntu) |
| Production | `zyt.app` |
| Local server | `fazt-local` systemd service |
| Apps dir | `servers/zyt/` (gitignored) |
| Binary | `~/.local/bin/fazt` |

## Essential Commands

```bash
# Build & Test
go build -o fazt ./cmd/server
go test ./...

# Deploy
fazt app deploy ./my-app --to zyt
fazt app deploy ./my-app --to local

# Local server
systemctl --user restart fazt-local
journalctl --user -u fazt-local -f
```

## Key Paths

```
cmd/server/           # Entry point, CLI
internal/
├── runtime/          # Serverless JS (Goja)
├── handlers/         # HTTP handlers
├── hosting/          # VFS, deploy logic
├── auth/             # OAuth, sessions
└── storage/          # KV, Docs, Blobs
koder/
├── STATE.md          # Current work
└── plans/            # Implementation plans
knowledge-base/
├── agent-context/    # Detailed dev context
└── skills/app/       # App development patterns
```

## Deep Context

Read these **as needed**, not every session:

| File | When to Read |
|------|--------------|
| `koder/STATE.md` | Start of session - current work |
| `knowledge-base/agent-context/setup.md` | Local server setup, SSH access |
| `knowledge-base/agent-context/architecture.md` | How fazt works, app model |
| `knowledge-base/agent-context/api.md` | API endpoints, CLI commands |
| `knowledge-base/agent-context/tooling.md` | Skills, knowledge-base, releasing |
| `knowledge-base/skills/app/` | App development patterns |

## Quick Reference

**Peers**: `zyt` (production), `local` (development)

**App structure**:
```
my-app/
├── manifest.json    # { "name": "my-app" }
├── index.html
├── api/main.js      # Serverless
└── private/         # Auth-gated files
```

**Create app**:
```bash
mkdir -p servers/zyt/my-app && cd servers/zyt/my-app
echo '{"name":"my-app"}' > manifest.json
```
