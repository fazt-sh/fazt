# Tooling & Skills

## Managing Fazt: Skills vs MCP

Two ways to manage fazt instances:

### Claude Skills (`.claude/commands/fazt-*.md`)
- Human-readable prompts that guide Claude Code
- Work via HTTP/curl to fazt API
- Portable: can be copied to any project

### MCP Server (`internal/mcp/`)
- Machine protocol for tool integration
- Configured via `.mcp.json` (gitignored)
- Tighter integration, type-safe

**Current Approach**: Both exist. Skills are simpler and portable. MCP is more
powerful but requires server running.

## Management Skills

| Skill | Description |
|-------|-------------|
| `/open` | Begin work session |
| `/fazt-app` | Build and deploy apps with Claude |
| `/release` | Release new version (full workflow) |
| `/close` | End work session |

## Knowledge-Base

The `knowledge-base/` directory contains versioned documentation for Claude skills.
It's symlinked to `~/.claude/skills/fazt-app`.

```
knowledge-base/
├── version.json        # Tracks KB version vs fazt version
└── skills/
    └── app/            # /fazt-app skill
        ├── SKILL.md
        ├── fazt/       # CLI docs
        ├── patterns/   # UI/auth patterns
        ├── references/ # API docs
        └── templates/  # Code templates
```

**When to update:**
- New CLI flag/command → `knowledge-base/skills/app/fazt/cli-*.md`
- New serverless API → `knowledge-base/skills/app/references/serverless-api.md`
- New pattern/workflow → `knowledge-base/skills/app/patterns/`

**After updates**, bump `knowledge-base/version.json` to match fazt version.

## Releasing

Use `/release` skill, which calls `scripts/release.sh`:

```bash
source .env                    # loads GITHUB_PAT_FAZT
./scripts/release.sh vX.Y.Z    # build all platforms, upload to GitHub
```

Fast local release (~30s) vs GitHub Actions (~4min).

## Documentation Maintenance

| Doc | Purpose | Update When |
|-----|---------|-------------|
| `CLAUDE.md` | Stable reference | Capabilities change |
| `koder/STATE.md` | Current work | Each session |
| `koder/THINKING_DIRECTIONS.md` | Strategic ideas | New directions emerge |
| `koder/CAPACITY.md` | Performance data | After benchmarks |
| `CHANGELOG.md` | Version history | Each release |
| `knowledge-base/` | Skills & reference docs | CLI/API changes |

## Markdown Style

All markdown files must be readable in raw format (terminal, vim, cat):

- **80 character width** - Wrap prose at 80 chars
- **Blank lines** - Before/after headings, between paragraphs
- **Short lines** - One sentence per line when possible
- **Tables** - Narrow tables OK; wide tables → bullet lists
- **Minimal HTML** - Avoid inline HTML, use standard markdown
