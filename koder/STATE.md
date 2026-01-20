# Fazt Implementation State

**Last Updated**: 2026-01-20
**Current Version**: v0.10.5

## Status

```
State: CLEAN
Next: (none)
```

---

## v0.10.5 - LLM-Friendly CLI (Complete)

Implemented all planned CLI improvements:

1. **New templates**: `static`, `vue`, `vue-api`
2. **`.gitignore` support**: Deploy respects .gitignore patterns
3. **`fazt app validate`**: Pre-deploy validation with JS syntax checking
4. **`fazt app logs`**: SSE streaming of serverless execution logs
5. **Better Goja errors**: Line numbers and context in error messages

---

## Previous Session: Reflex App + DX Ideas

### 1. Reflex - Example App (Complete)

Built an advanced typing speed game to showcase fazt capabilities:
- **Live**: https://reflex.zyt.app
- **Source**: `servers/zyt/reflex/`

**Features demonstrated:**
- Document store (ds) for session state, user stats
- Key-value (kv) for leaderboards, daily challenges, multiplayer with TTL
- Complex queries for aggregating rankings
- Vue 3 + Vite dual-mode (static hosting + dev server)
- PWA-ready with sounds, haptics, themes

**README includes feature proposals** for what would be possible with:
- WebSockets (v0.17) - real-time multiplayer
- Cron (v0.10) - daily challenge generation
- Stdlib (v0.10) - lodash, dayjs
- Sandbox (v0.10) - user-generated challenges

---

## Next: LLM-Friendly App Development

Goal: Make it possible for cheaper models (Haiku) to build fazt apps reliably.

### Design Principle

**Use existing concepts, don't invent new ones.**

- Use `.gitignore` for deploy exclusions (not `.faztignore`)
- Use `package.json` patterns (not `app.json`)
- Copy from templates (not interpret long specs)

### Fazt CLI Additions (Priority Order)

#### 1. `fazt app create --template`

Scaffold working boilerplate so models only write business logic.

```bash
fazt app create myapp --template static   # index.html + manifest
fazt app create myapp --template vue      # Vue + Vite + lib/
fazt app create myapp --template api      # With serverless API
```

**Implementation:**
- Embed templates in binary (like admin dashboard)
- Templates at `internal/assets/templates/`
- Each template is a working app out-of-box

#### 2. Respect `.gitignore` on deploy

Don't upload `node_modules/`, `dist/`, `.git/`, etc.

**Implementation:**
- Parse `.gitignore` in deploy command
- Use existing gitignore libraries (go-gitignore)
- No new config files to learn

#### 3. `fazt app validate <dir>`

Check app before deploy, return actionable errors.

```bash
fazt app validate ./myapp
# ✓ manifest.json valid
# ✓ index.html found
# ✗ api/main.js:15 - SyntaxError: Unexpected token
```

**Implementation:**
- Check manifest.json schema
- Check required files exist
- Parse JS with Goja, report errors with line numbers

#### 4. `fazt app logs <app> [--peer]`

Stream serverless execution logs.

```bash
fazt app logs reflex --peer local
# 2026-01-20 12:00:01 [reflex] GET /api/state → 200 (12ms)
# 2026-01-20 12:00:02 [reflex] POST /api/state → 200 (8ms)
```

**Implementation:**
- Ring buffer for recent logs (already have site_logs table)
- SSE endpoint `/_fazt/logs/stream`
- CLI connects and prints

#### 5. Better Goja error messages

When serverless fails, return line number and context.

```json
{
  "error": "ReferenceError: foo is not defined",
  "file": "api/main.js",
  "line": 42,
  "context": "var x = foo.bar();"
}
```

### /fazt-app-lite Skill

After CLI improvements, create a minimal skill (~50 lines):

```markdown
# /fazt-app-lite

1. Run: `fazt app create <name> --template vue-api`
2. Edit `components/App.js` - add your UI
3. Edit `api/main.js` - add your endpoints
4. Test: `fazt app deploy <dir> --to local`
5. Ship: `fazt app deploy <dir> --to zyt`

## Storage (in api/main.js)
var ds = fazt.storage.ds;
ds.insert('items', {name: 'x'});
ds.find('items', {name: 'x'});
```

The skill is short because the CLI does the heavy lifting.

### Templates to Create

```
internal/assets/templates/
├── static/           # Just HTML/CSS/JS
│   ├── manifest.json
│   └── index.html
├── vue/              # Vue 3 + Vite
│   ├── manifest.json
│   ├── package.json
│   ├── vite.config.js
│   ├── index.html
│   ├── main.js
│   └── components/
│       └── App.js
└── vue-api/          # Vue + serverless
    ├── (all of vue/)
    ├── api/
    │   └── main.js   # CRUD template
    └── lib/
        ├── api.js    # Fetch helpers
        ├── session.js
        └── settings.js
```

---

## Quick Reference

```bash
# Current workflow
fazt app deploy servers/zyt/myapp --to local --no-build

# Future workflow (after CLI improvements)
fazt app create myapp --template vue-api
# ... edit files ...
fazt app validate servers/zyt/myapp
fazt app deploy servers/zyt/myapp --to local
fazt app logs myapp --peer local
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
| VFS (fixed) | `internal/hosting/vfs.go` |
| Apps Handler (fixed) | `internal/handlers/apps_handler.go` |
