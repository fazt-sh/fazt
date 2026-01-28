# Fazt Implementation State

**Last Updated**: 2026-01-29
**Current Version**: v0.11.5

## Status

State: CLEAN - fazt-app skill restructured, mock OAuth planned

---

## Next Up

### Plan 24: Mock OAuth Provider

Enable full auth flow testing locally without code changes.

```
Local:  "Sign in" → Dev form → Session → fazt.auth.getUser() ✓
Remote: "Sign in" → Google   → Session → fazt.auth.getUser() ✓
```

Same code. Same API. Different provider.

**Key features:**
- Dev login form at `/auth/dev/login` (local only)
- Creates real session (same as production OAuth)
- Role selection for testing admin/owner flows
- Blocked on HTTPS (production safe)

See: `koder/plans/24_mock_oauth.md`

---

## Last Session

**fazt-app Skill Restructuring**

1. **Restructured into granular files** (was 938 lines, now 16 files)
   - `fazt/` - Platform docs (overview, cli-*, deployment)
   - `references/` - APIs, auth integration, design system
   - `patterns/` - Layout, modals, testing
   - `examples/` - Cashflow reference app

2. **Made skill generic** (removed zyt.app references)
   - Uses `<peer>`, `<remote-peer>`, `<domain>` placeholders
   - Anyone can use with their own fazt setup

3. **Added key documentation:**
   - "Build free, but buildable" paradigm explained
   - Auth evaluation workflow (when to ask user)
   - OAuth requires remote (HTTPS) - clearly documented
   - Always deploy `dist/` for production

4. **Updated philosophy** (from scratch/01_fork-to-multiuser.md)
   - "Single-owner compute node that can support multiple users"
   - Comparable to Supabase/Vercel in value, fully self-contained

## Ideas for Later

### Docs as Claude Skill (`fazt ai skill`)

Instead of building help search into CLI, ship docs as installable skill:
- Docs live with source (always synced)
- `fazt ai skill install --global` copies to ~/.claude/skills/fazt/
- LLM does search/comprehension (no CLI complexity)
- See: `koder/ideas/specs/v0.12-agentic/skill.md`

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
fazt @<peer> auth providers
```
