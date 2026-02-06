# Tooling & Skills

## Unified Skills Architecture

**Single Source of Truth**: All skills live in `knowledge-base/skills/` (git-tracked).

**Multi-Agent Support**: Skills work with both Claude Code and OpenAI Codex via symlinks:
- `.claude/skills/` → `knowledge-base/skills/` (Claude discovery)
- `.agents/skills/` → `knowledge-base/skills/` (Codex/OpenCode discovery)

**Global vs Local**:
- **Global**: `~/.claude/skills/fazt-app` → `~/Projects/fazt/knowledge-base/skills/app`
- **Local**: Project skills discovered via `.claude/skills/` and `.agents/skills/`

## Managing Fazt: Skills vs MCP

Two ways to manage fazt instances:

### Skills (`knowledge-base/skills/`)
- Human-readable prompts that guide AI agents
- Work via HTTP/curl to fazt API
- Portable: can be copied to any project
- **New**: Multi-agent compatible (Claude + Codex)

### MCP Server (`internal/mcp/`)
- Machine protocol for tool integration
- Configured via `.mcp.json` (gitignored)
- Tighter integration, type-safe

**Current Approach**: Both exist. Skills are simpler and portable. MCP is more
powerful but requires server running.

## Available Skills

| Skill | Description | Scope |
|-------|-------------|-------|
| `/fazt-open` | Begin work session | Local |
| `/fazt-close` | End work session | Local |
| `/fazt-release` | Release new version (full workflow) | Local |
| `/fazt-ideate` | Brainstorm ideas (uses reasoning model) | Local |
| `/fazt-lite-extract` | Evaluate library extraction feasibility | Local |
| `/fazt-app` | Build and deploy apps | Global |

## Knowledge-Base Structure

The `knowledge-base/` directory contains versioned documentation and all skills:

```
knowledge-base/
├── AGENTS.md           # Instructions (symlinked to root as CLAUDE.md & AGENTS.md)
├── version.json        # Tracks KB version vs fazt version
└── skills/             # All skills (single source)
    ├── app/            # /fazt-app skill (global)
    │   ├── SKILL.md
    │   ├── fazt/       # CLI docs
    │   ├── patterns/   # UI/auth patterns
    │   ├── references/ # API docs
    │   └── templates/  # Code templates
    ├── release/        # /fazt-release skill (local)
    ├── open/           # /fazt-open skill (local)
    ├── close/          # /fazt-close skill (local)
    ├── ideate/         # /fazt-ideate skill (local)
    └── lite-extract/   # /fazt-lite-extract skill (local)
```

**When to update content:**
- New CLI flag/command → `knowledge-base/skills/app/fazt/cli-*.md`
- New serverless API → `knowledge-base/skills/app/references/serverless-api.md`
- New pattern/workflow → `knowledge-base/skills/app/patterns/`

**Version syncing:**
- **Always** bump `knowledge-base/version.json` to match fazt version on release
- KB version is a sync marker, not just a content change indicator
- Ensures KB is verified as compatible with the released binary version

## Releasing

Use `/fazt-release` skill, which calls `scripts/release.sh`:

```bash
source .env                    # loads GITHUB_PAT_FAZT
./scripts/release.sh vX.Y.Z    # build all platforms, upload to GitHub
```

Fast local release (~30s) vs GitHub Actions (~4min).

## Documentation Maintenance

| Doc | Purpose | Update When |
|-----|---------|-------------|
| `knowledge-base/AGENTS.md` | Instructions (single source) | Capabilities change |
| `CLAUDE.md` / `AGENTS.md` | Symlinks to knowledge-base/ | Auto-synced |
| `koder/STATE.md` | Current work | Each session |
| `koder/THINKING_DIRECTIONS.md` | Strategic ideas | New directions emerge |
| `koder/CAPACITY.md` | Performance data | After benchmarks |
| `CHANGELOG.md` | Version history | Each release |
| `knowledge-base/skills/` | All skills & reference docs | CLI/API changes |

## Markdown Style

All markdown files must be readable in raw format (terminal, vim, cat):

- **80 character width** - Wrap prose at 80 chars
- **Blank lines** - Before/after headings, between paragraphs
- **Short lines** - One sentence per line when possible
- **Tables** - Narrow tables OK; wide tables → bullet lists
- **Minimal HTML** - Avoid inline HTML, use standard markdown
