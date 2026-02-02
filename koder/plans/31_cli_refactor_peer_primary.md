# Plan 31: CLI Refactor - @peer Primary Pattern

**Status**: Ready to Implement
**Created**: 2026-02-02
**Related**: E1 (THINKING_DIRECTIONS.md)

## Summary

Refactor Fazt CLI to make `@peer` the primary pattern for remote operations, replacing inconsistent `--to`/`--from`/`--on` flags. Rename `remote` command to `peer`. Create unified documentation in `knowledge-base/cli/` that serves both web docs and binary `--help`.

## Motivation

**Current problems:**
- Five different peer targeting patterns (`--to`, `--from`, `--on`, positional, `@peer`)
- Semantic confusion (`--from` means "source" in pull, "target" in remove)
- No local-first defaults (deploy requires `--to` even with one peer)
- Identifier ambiguity (some commands require flags, others accept positional)

**With @peer-primary:**
```bash
# Local by default
fazt app list
fazt app deploy ./my-app

# Remote explicit with @peer prefix
fazt @zyt app list
fazt @zyt app deploy ./my-app

# Clean, consistent, one rule
```

## Design

### Core Principle: Local-First, Explicit Remote

| Pattern | Meaning |
|---------|---------|
| `fazt <command>` | Operates on LOCAL instance (implicit) |
| `fazt @local <command>` | Operates on LOCAL instance (explicit) |
| `fazt @peer <command>` | Operates on REMOTE peer (explicit) |

**Key addition:** `@local` is optional but improves documentation clarity and script readability. Both `fazt app list` and `fazt @local app list` do the same thing.

### Command Categories

**Single-Peer Operations** (query/modify):
```bash
# Local (implicit or explicit)
fazt app list              # local (implicit)
fazt @local app list       # local (explicit)

# Remote (always explicit)
fazt @zyt app list         # zyt

# Remove examples
fazt @local app remove myapp    # local (explicit)
fazt @zyt app remove myapp      # zyt
```

**Data Transfer** (two peers):
```bash
# Pull from remote to local
fazt @zyt app pull myapp --to ./local

# Pull from local to directory (explicit local)
fazt @local app pull myapp --to ./backup

# Fork across peers
fazt @zyt app fork myapp --to prod

# Fork from local to remote (explicit)
fazt @local app fork myapp --to prod
```

`@peer` (or `@local`) = source, `--to` = optional destination

### Changes Required

1. **Rename `remote` → `peer`**
   ```bash
   # Old
   fazt remote add zyt --url ... --token ...
   fazt remote list

   # New
   fazt peer add zyt --url ... --token ...
   fazt peer list
   ```

2. **Remove peer flags from app commands**
   - Remove `--to`, `--from`, `--on` flags
   - Make `@peer` the only way to target remotes
   - Local is default (no flag needed)

3. **Add `@local` support**
   ```go
   // Handle @local explicitly
   if strings.HasPrefix(command, "@") {
       peerName := command[1:]
       if peerName == "local" {
           handleLocalCommand(os.Args[2:])
       } else {
           handlePeerCommand(peerName, os.Args[2:])
       }
   }
   ```

4. **Positional identifiers**
   ```bash
   # Works with positional
   fazt app info myapp
   fazt app fork myapp --as myapp-v2

   # Explicit when needed
   fazt app info --id app_123
   ```

## Documentation Structure

### Single Source of Truth: knowledge-base/cli/

```
knowledge-base/cli/
├── _meta.yaml              # CLI metadata, no version field
├── fazt.md                 # Top-level overview
├── app/
│   ├── _index.md
│   ├── deploy.md
│   ├── list.md
│   └── ... (17 commands)
├── peer/                   # Renamed from remote
│   ├── _index.md
│   ├── add.md
│   ├── list.md
│   └── ... (7 commands)
├── server/
├── service/
└── topics/
    ├── peer-syntax.md      # @peer pattern guide
    └── local-first.md      # Philosophy
```

### Frontmatter Schema (Blog-Aligned)

```yaml
---
# Standard blog fields (works with any theme)
title: "App Deploy Command"
date: 2026-02-02
description: "Deploy apps to fazt instances"
tags: [cli, deployment]
category: commands

# CLI-specific extensions (optional)
command: "app deploy"
syntax: "fazt [@peer] app deploy <directory> [flags]"
peer:
  supported: true
  local: true
  remote: true
---

# App Deploy

Deploy a local directory to a fazt instance...
```

**Key decision:** No `version` field in frontmatter (use monorepo version.json)

### Markdown Guidelines

- Use GitHub Flavored Markdown (GFM)
- Keep lines <80 chars where reasonable
- Use code blocks for examples
- Use lists for arguments/flags
- Clean formatting (works as plain text)

## Implementation Strategy

### Phase 0: Preparation (1-2 days)
- Document current CLI behavior comprehensively
- Create test fixtures (mock DB, test peers, test apps)
- Identify all commands and their current flags

### Phase 1: Tests First (2-3 days)
- Write tests for @peer-primary behavior (TDD)
- Unit tests: @peer parsing, identifier resolution
- Integration tests: end-to-end command execution
- Regression tests: ensure old functionality works
- **Tests will fail initially** (expected)

### Phase 2: Documentation (2-3 days)
- Create all 34+ markdown files in knowledge-base/cli/
- Follow frontmatter schema (blog-aligned)
- Write clean markdown (<80 chars, GFM)
- Document @peer pattern clearly

### Phase 3: Help System (2-3 days)
- Create `internal/help/` package
- Embed docs at build time (go:embed)
- Parse frontmatter (yaml.v3)
- Render to terminal (glamour/charmbracelet)
- Wire `--help` to docs

### Phase 4: CLI Refactor (3-5 days)
- Rename `remote` → `peer` namespace
- Enhance `handlePeerCommand`
- Remove `--to`, `--from`, `--on` flags
- Make @peer the only remote pattern
- Update all commands
- **Tests should pass now**

### Phase 5: Polish (1-2 days)
- Error messages mention @peer pattern
- Helpful hints for common mistakes
- Update examples
- Final testing

**Total: 11-18 days**

## Files to Create/Modify

### New Files
```
knowledge-base/cli/**/*.md     # 34+ documentation files
internal/help/embed.go         # Embed docs
internal/help/parser.go        # Parse frontmatter
internal/help/renderer.go      # Terminal rendering
cmd/server/peer.go             # Renamed from remote.go
```

### Modified Files
```
cmd/server/main.go             # @peer handling, help system
cmd/server/app.go              # Remove peer flags
cmd/server/app_*.go            # Update all app commands
```

### Tests
```
cmd/server/main_test.go        # @peer parsing
cmd/server/app_test.go         # App commands with @peer
cmd/server/peer_test.go        # Peer management
internal/help/parser_test.go   # Doc parsing
```

## Success Criteria

- [ ] All commands use @peer pattern consistently
- [ ] `fazt peer` replaces `fazt remote`
- [ ] No `--to`/`--from`/`--on` flags remain
- [ ] Local operations work without peer flags
- [ ] `fazt --help` reads from knowledge-base/cli/
- [ ] All tests pass
- [ ] Documentation complete and accurate

## Related

- **Plan 25**: SQL Command (depends on this - needs @peer)
- **Plan 32**: Docs Rendering System (uses same markdown files)
- **THINKING_DIRECTIONS.md**: E1 - `fazt @peer` Pattern Audit

## Notes

- No backward compatibility needed (single user)
- Breaking changes acceptable
- Tests prevent regressions
- Documentation is single source of truth
