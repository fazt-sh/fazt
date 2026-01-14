# Fazt - Deep Implementation Bootstrap

> **For general context, read `CLAUDE.md` in the repo root first.**
> This file is for deep implementation work requiring a plan.

---

## When to Use This

Use this bootstrap when:
- Starting a new major feature (needs a plan)
- Implementing from a spec in `koder/ideas/specs/`
- Multi-phase work spanning sessions

For day-to-day work (bug fixes, small features, app development, managing
zyt), `CLAUDE.md` has everything needed.

---

## Implementation Protocol

### Step 1: Check State

```
Read: koder/STATE.md
```

If state is "Clean" with no active plan, create one first:
1. Create `koder/plans/XX_plan-name.md`
2. Update STATE.md with plan reference

### Step 2: Load Context

```
Read: CLAUDE.md                    # Always (environment, capabilities)
Read: koder/STATE.md               # Current progress
Read: koder/plans/<active-plan>    # Implementation details
```

### Step 3: Verify Tests Pass

Before ANY new work:

```bash
go test ./...
```

Fix failures before proceeding.

### Step 4: Execute

1. Implement code
2. Write tests
3. Run tests
4. Update STATE.md with progress

### Step 5: Update State

After EVERY significant step, update `koder/STATE.md`:
- Mark completed items
- Update current task
- Log progress

---

## Architecture

```
fazt (binary)
├── cmd/server/main.go      # Entry point, CLI
├── internal/
│   ├── handlers/           # HTTP handlers
│   ├── hosting/            # VFS, deploy logic
│   ├── database/           # SQLite init, migrations
│   ├── config/             # Server config
│   ├── remote/             # Peer management, client
│   ├── mcp/                # MCP server
│   ├── runtime/            # JS serverless runtime
│   └── api/                # Response helpers
├── admin/                  # React SPA (Vite + Tailwind)
└── ~/.config/fazt/data.db  # Client database (peers, config)
```

---

## Commands

```bash
# Build & Test
go build -o ~/.local/bin/fazt ./cmd/server
go test -v -cover ./...

# Run locally
fazt server start --port 8080

# Manage remote
fazt remote status zyt
fazt remote deploy <dir> zyt
```

---

## Specs for Future Work

Future features are spec'd in `koder/ideas/specs/`:

```
v0.9-storage/     # Blobs, documents
v0.10-runtime/    # Stdlib, sandbox
v0.11-distribution/  # Marketplace
v0.12-agentic/    # AI harness
v0.13-network/    # Domains, VPN
v0.14-security/   # RLS, notary
v0.15-identity/   # Persona
v0.16-mesh/       # P2P protocols
```

Pick a spec, create a plan, update STATE.md, execute.

---

## Critical Rules

1. **CLAUDE.md is primary** - Read it first, always
2. **Tests before commits** - Never commit failing tests
3. **Update STATE.md** - After every significant step
4. **One plan at a time** - Complete before starting next
