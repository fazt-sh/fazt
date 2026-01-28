# Fazt Implementation State

**Last Updated**: 2026-01-28
**Current Version**: v0.11.5

## Status

State: CLEAN - Skills reorganized, moved to global ~/.claude/

---

## Last Session

**Skills & Commands Reorganization**

1. **Created /fazt-app skill** (comprehensive)
   - Full Vue + Pinia + Tailwind stack
   - Patterns: layout, modals, testing, google-oauth
   - Examples: cashflow reference app
   - Templates: complete app scaffold
   - Auth workflow with provider check

2. **Moved skills to global ~/.claude/**
   - `~/.claude/skills/fazt-app/` - app development
   - `~/.claude/skills/agent-browser/` - browser automation
   - `~/.claude/fazt-assets/` - branding assets

3. **Renamed repo commands** (shorter names)
   - `/fazt-start` → `/open`
   - `/fazt-stop` → `/close`
   - `/fazt-release` → `/release`
   - `/fazt-ideate` → `/ideate`
   - `/fazt-lite-extract` → `/lite-extract`

4. **Cleaned up repo's .claude/**
   - Removed skills/ (now global)
   - Removed fazt-assets/ (now global)
   - Only repo-specific commands remain

## Key Learnings

- Skills format: `~/.claude/skills/<name>/SKILL.md` with YAML frontmatter
- Commands format: `~/.claude/commands/<name>.md` (simpler)
- `fazt @<peer> auth providers` - check if OAuth configured
- Global skills work across all repos

---

## Next Explorations

### Guestbook Improvements (from previous session)
- [ ] Pagination - show 20 messages per page
- [ ] FAB + modal for adding entries
- [ ] Inner page version - guestbook as route on zyt.app

### Auth DX
- [ ] Runtime helpers for common auth patterns
- [ ] Headless auth components

---

## Quick Reference

```bash
# Session commands (this repo)
/open                    # Start session
/close                   # End session
/release                 # Release workflow

# Global skills (any repo)
/fazt-app               # Build fazt apps
/agent-browser          # Browser automation
/qwen-research          # Deep research

# Check OAuth status
fazt @zyt auth providers
```
