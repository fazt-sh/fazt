# Claude Skill Management

**Status**: Proposed
**Namespace**: `fazt ai skill`

## Context

Fazt documentation should live with source code (always synced) but be usable
by LLMs for comprehension and assistance. Instead of building search into fazt,
ship docs as a Claude skill that LLMs can read.

## The `fazt ai` Namespace

AI-related features are separated from core fazt functionality:

```bash
fazt ai skill     # Claude skill management
fazt ai mcp       # MCP server (future)
fazt ai agents    # Agent capabilities (future)
fazt ai api       # AI API access (future)
fazt ai providers # AI provider config (future)
```

This namespace is fluid - features will evolve. But segregating AI features
from core commands (`app`, `remote`, `auth`, `server`) keeps things clean.

## Skill Management

### Commands

```bash
fazt ai skill install              # Install to .claude/skills/fazt/
fazt ai skill install --global     # Install to ~/.claude/skills/fazt/
fazt ai skill update               # Update to match current fazt version
fazt ai skill remove               # Remove installed skill
fazt ai skill path                 # Show where skill would be installed
```

### Independence

Skill management is **completely independent** from fazt installation:

```bash
# These are separate operations
./install.sh                 # Install fazt binary
fazt ai skill install        # Install Claude skill (optional, separate)
```

The skill is a local-only feature for developers using Claude Code. It's not
required for fazt to function.

## How It Works

```
fazt repo/
├── docs/                    # Lives with source code
│   └── skill/               # Structured as Claude skill
│       ├── SKILL.md
│       ├── platform/
│       │   ├── overview.md
│       │   ├── cli-app.md
│       │   └── deployment.md
│       └── references/
│           └── serverless-api.md
└── cmd/server/
    └── ai.go                # fazt ai commands
```

### Installation Flow

```bash
$ fazt ai skill install --global

Installing fazt skill to ~/.claude/skills/fazt/
  Copying platform/overview.md
  Copying platform/cli-app.md
  ...
Done. Skill version: 0.11.5

Restart Claude Code to load the skill.
```

### Usage

```
User: "How do I deploy an app with auth?"

Claude: [Has fazt skill loaded]
        [Reads docs/skill/platform/deployment.md]
        [Reads docs/skill/references/auth-integration.md]
        [Understands context, executes commands]
```

## Benefits

### 1. Always Synced

Docs live in same repo as code:
```bash
git log --oneline docs/ cmd/
# Changes to CLI and docs in same commits
```

### 2. LLM Does Heavy Lifting

- Semantic understanding (not keyword match)
- Context-aware answers
- Can combine multiple docs
- No CLI complexity for search

### 3. Version Matched

```bash
fazt version              # 0.11.5
fazt ai skill install     # Installs skill for 0.11.5
```

### 4. Works Offline

Skill is local files. No API calls.

## Skill Structure

```
docs/skill/
├── SKILL.md                    # Entry point
├── platform/                   # Core fazt docs
│   ├── overview.md
│   ├── deployment.md
│   ├── cli-app.md
│   ├── cli-remote.md
│   ├── cli-auth.md
│   └── cli-server.md
├── references/
│   ├── serverless-api.md
│   ├── storage-api.md
│   └── auth-integration.md
├── patterns/
│   ├── layout.md
│   └── modals.md
└── examples/
    └── cashflow.md
```

## Skill Metadata

```yaml
# SKILL.md frontmatter
---
name: fazt
description: Documentation for fazt - single-owner compute platform
version: 0.11.5
context: fork
---
```

## Implementation

```go
// cmd/server/ai.go

func aiSkillInstall(global bool) error {
    var target string
    if global {
        home, _ := os.UserHomeDir()
        target = filepath.Join(home, ".claude", "skills", "fazt")
    } else {
        target = filepath.Join(".claude", "skills", "fazt")
    }

    // Remove existing
    os.RemoveAll(target)

    // Copy embedded docs
    return fs.WalkDir(docsFS, "docs/skill", func(path string, d fs.DirEntry, err error) error {
        // Copy files to target...
    })
}
```

## Relation to fazt-app Skill

Currently `~/.claude/skills/fazt-app/` exists for building apps. Options:

1. **Keep separate**: fazt (platform) + fazt-app (building)
2. **Merge**: Single comprehensive fazt skill

Recommendation: Merge into one skill installed via `fazt ai skill install`.

## Open Questions

1. **Auto-update prompt?** Warn if skill version != fazt version?
2. **MCP integration?** `fazt ai mcp` could start MCP server
3. **Skill registry?** List available skills from fazt ecosystem?

## Success Criteria

- [ ] `fazt ai skill install` works
- [ ] Skill version matches fazt version
- [ ] Claude can use skill for fazt assistance
- [ ] Docs synced with code (same repo)
- [ ] Independent from fazt binary installation
