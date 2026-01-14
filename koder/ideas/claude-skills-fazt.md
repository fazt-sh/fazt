# Claude Skills for Fazt Management

## Context (recorded 2026-01-14)

### Core Need
Build Claude skills in this repo for managing fazt instances, designed to:
1. Co-evolve fazt development alongside zyt.app management
2. Keep instance-specific config (like `zyt.app`) out of committed code
3. Be modular enough to extract globally for any Claude instance later

### Key Requirements

**Server Configuration**
- Don't hardcode server URLs (e.g., `zyt.app`) in repo
- Support multiple servers (not just single .env)
- Current structure: `servers/zyt/` - gitignored, instance-specific
- Potential: `servers/<name>/config.json` for each instance?

**Skills to Build**
- Deploy to fazt instance
- Check server status/health
- Upgrade remote fazt binary
- List apps on instance
- View logs/analytics
- Manage env vars

**Integration Points**
- MCP server already exists (`internal/mcp/`)
- `.mcp.json` (gitignored) holds instance config
- CLAUDE.md should be fazt-aware
- Skills should know current fazt version, capabilities

### Architecture Thoughts

```
servers/                    # gitignored
├── zyt/
│   ├── config.json        # { "url": "https://admin.zyt.app", "token": "..." }
│   └── xray/              # sites for this instance
└── local/
    └── config.json        # local dev instance

.claude/
└── skills/
    └── fazt/
        ├── deploy.md      # skill: deploy to fazt
        ├── status.md      # skill: check status
        └── upgrade.md     # skill: upgrade instance
```

### Open Questions
- How do skills reference which server to use?
- Should skills read from `servers/<name>/config.json`?
- Or use environment variables?
- How to make skills portable (extractable to global)?

### Implementation Status (2026-01-14)

**Completed**:
- [x] Server config structure: `servers/<name>/config.json`
- [x] Created skills in `.claude/commands/`:
  - `fazt-status.md` - Health checks
  - `fazt-deploy.md` - Deploy sites/apps
  - `fazt-apps.md` - List/manage apps
  - `fazt-upgrade.md` - Check/perform upgrades
- [x] Updated CLAUDE.md with fazt awareness
- [x] Token via config.json or `FAZT_TOKEN_<NAME>` env var

**Next Steps**:
1. Add token to `servers/zyt/config.json`
2. Test skills with zyt.app
3. Consider extracting skills globally later
