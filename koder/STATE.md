# Fazt Implementation State

**Last Updated**: 2026-01-29
**Current Version**: v0.11.5

## Status

State: CLEAN - skill docs fixed, SQL command planned

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

### Plan 25: SQL Command

Debug local and remote fazt instances with direct SQL queries.

```bash
fazt sql "SELECT * FROM apps"              # Local
fazt @zyt sql "SELECT * FROM auth_users"   # Remote
```

**Key features:**
- Read-only by default, `--write` flag for mutations
- Table, JSON, CSV output formats
- API endpoint `POST /api/sql` for remote
- Same syntax local and remote

See: `koder/plans/25_sql_command.md`

---

## Last Session

**fazt-app Skill Documentation Fixes**

1. **Fixed deployment docs** - `fazt app deploy` has built-in build
   - Point at project root, not `dist/`
   - Auto-detects package.json, runs build, deploys output
   - Updated SKILL.md, deployment.md, cli-app.md

2. **Fixed OAuth redirect bug in docs**
   - OAuth flows through root domain
   - Relative redirects (`/path`) lose subdomain after callback
   - Must use absolute URLs (`window.location.href`)
   - Updated auth-integration.md, SKILL.md, patterns/google-oauth.md

3. **Removed hardcoded zyt.app references**
   - Replaced with generic `<domain>`, `example.com` placeholders

4. **Created Plan 25: SQL Command**
   - `fazt sql` for local/remote database queries
   - Debugging without SSH or direct db access

## Issues Found (Lens Integration)

Real-world OAuth integration uncovered these fazt issues:

### 1. Deploy Doesn't Auto-Include api/ Folder

`fazt app deploy` only deploys the build output directory (`dist/`). Projects with `api/` at root must manually copy it into dist.

**Current workaround:** `"build": "vite build && cp -r api dist/"`

**Expected:** fazt should auto-detect `api/` at project root and merge with build output.

### 2. Serverless Code Has Quirks

- `return` statements cause "Illegal return statement" error
- Complex if/else chains can cause timeouts
- `fazt.auth.getUser()` intermittently times out (cold start?)

**Current workaround:** Keep API code minimal, single respond() call. Frontend needs retry logic.

**Expected:** More robust JS execution, better error messages.

### 3. Logout Requires POST (Not Documented)

`/auth/logout` returns 405 on GET. Must use POST. Not mentioned in skill docs.

**Fix:** Update auth docs to show POST requirement:
```javascript
fetch('/auth/logout', { method: 'POST' })
```

---

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
