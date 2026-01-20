# Fazt Implementation State

**Last Updated**: 2026-01-20
**Current Version**: v0.10.5

## Status

```
State: PLANNING
Next: Simplify /fazt-app skill
```

---

## Task: Simplify /fazt-app

**Goal**: Make fazt app building reliable for cheaper models (Haiku-class).

### The Problem

Current `/fazt-app` is 1000+ lines. It tries to teach the LLM everything
about Vue, Vite, storage APIs, component patterns, etc. This makes it:
- Brittle (LLM can miss details)
- Expensive (lots of context)
- Inconsistent (generates from scratch each time)

### The Solution

v0.10.5 CLI now does the heavy lifting:

| Capability | What It Does |
|------------|--------------|
| `fazt app create --template vue-api` | Scaffolds complete project |
| `fazt app validate` | Pre-deploy validation with JS syntax |
| `fazt app logs -f` | Real-time debugging |
| `.gitignore` support | Auto-excludes node_modules, etc. |
| Better Goja errors | Line numbers and context |

**New approach**: Skill orchestrates CLI, doesn't duplicate knowledge.

### Target

Reduce `/fazt-app` from ~1000 lines to ~100-150 lines:

```markdown
# /fazt-app

1. Scaffold: `fazt app create <name> --template vue-api`
2. Customize:
   - UI: `src/components/App.js`
   - API: `api/items.js` (rename collection, add endpoints)
3. Validate: `fazt app validate ./<name>`
4. Test: `fazt app deploy ./<name> --to local`
5. Debug: `fazt app logs <name> --peer local -f`
6. Ship: `fazt app deploy ./<name> --to zyt`

## Storage Quick Reference
var ds = fazt.storage.ds
ds.insert('items', {name: 'x'})
ds.find('items', {})
ds.update('items', {id: '...'}, {$set: {name: 'y'}})
ds.delete('items', {id: '...'})

## Design Notes
- Apple-esque: clean, spacious, subtle
- Components in separate .js files
- Session via ?s= URL param (already in template)
```

### What to Keep

- CLI command sequence
- Storage API quick reference (ds, kv basics)
- Brief design principles
- Location behavior (servers/zyt/ vs /tmp/)

### What to Remove

- Full boilerplate code (templates have it)
- Component architecture details
- Detailed Vue/Vite patterns
- 500 lines of example code

### Implementation Steps

1. Read current `/fazt-app` skill
2. Read `vue-api` template to verify it has everything
3. Write new minimal skill
4. Test with a sample prompt
5. Delete old verbose skill

---

## Future Enhancement (v0.10.6)

Consider `fazt app dev` command:
```bash
fazt app dev ./myapp --peer local
# Validates → Deploys → Opens browser → Tails logs → Watches for changes
```

Would reduce iteration loop to single command.

---

## Quick Reference

```bash
# Create app
fazt app create myapp --template vue-api

# Validate
fazt app validate ./myapp

# Local testing
fazt app deploy ./myapp --to local
fazt app logs myapp --peer local -f

# Production
fazt app deploy ./myapp --to zyt
```
