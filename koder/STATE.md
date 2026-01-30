# Fazt Implementation State

**Last Updated**: 2026-01-30
**Current Version**: v0.13.0

## Status

State: **CLEAN** - Plan 30 designed, CLAUDE.md refactored

---

## Active Directions

See `koder/THINKING_DIRECTIONS.md` for full list.

**Next up**:
- Plan 30: User Isolation & Analytics (ready to implement)
- P2: Nexus App (stress test all capabilities)
- E4: Plan 24 - Mock OAuth (local auth testing)

---

## Pending Plans

| Plan | Status | Purpose |
|------|--------|---------|
| 30: User Isolation & Analytics | Designed | User data isolation, analytics, GDPR |
| 29: Private Directory | ✅ Released v0.13.0 | `private/` with dual access |
| 24: Mock OAuth | Not started | Local auth testing |
| 25: SQL Command | Not started | Remote DB debugging |

---

## RBAC Notes (For Future Discussion)

Full RBAC was discussed but deferred. Key points:

**Pros:**
- Fine-grained permissions (`posts:create`, `posts:delete:own`)
- Custom roles per app (`editor`, `moderator`, `viewer`)
- Framework-level enforcement

**Cons:**
- Significant API complexity
- Most apps need only owner/admin/user
- Can be implemented at app level if needed
- Delays other features

**Decision:** Keep simple roles (owner/admin/user). Revisit when needed.

---

## Current Session (2026-01-30)

**Plan 30 Design + CLAUDE.md Refactor**

### Plan 30: User Isolation & Analytics

Comprehensive design for v0.14.0:

1. **API Namespace (Option C2)**
   - `fazt.app.user.*` - User's private data (auto-scoped)
   - `fazt.app.*` - App's shared data
   - `fazt.app.private.*` - Bundled private files
   - `fazt.auth.*` - Authentication
   - `fazt.admin.*` - Admin operations
   - `fazt.analytics.*` - Event tracking

2. **ID Format (Stripe-style)**
   - `fazt_usr_<12 chars>` - User
   - `fazt_app_<12 chars>` - App
   - `fazt_tok_<12 chars>` - Token
   - `fazt_ses_<12 chars>` - Session

3. **Analytics Enhancement**
   - Add `app_id`, `user_id` to events table
   - Query by app, user, or both
   - Track user journey across apps

4. **GDPR Compliance**
   - `fazt.admin.users.delete(userId)` removes all data

5. **Role Model**
   - One owner per instance
   - Multiple admins allowed
   - Simple roles: owner/admin/user

### CLAUDE.md Refactor

Split CLAUDE.md for ~80% token reduction:

| Before | After |
|--------|-------|
| 456 lines (~3.9k tokens) | 85 lines (~800 tokens) |

**New structure:**
```
knowledge-base/
├── agent-context/         # Fazt development context
│   ├── setup.md           # Local server, SSH
│   ├── architecture.md    # How fazt works
│   ├── api.md             # API endpoints, CLI
│   └── tooling.md         # Skills, releasing
└── skills/app/            # App development
```

### Files Created/Changed

- `koder/plans/30_user_isolation_analytics.md` - Full implementation plan
- `CLAUDE.md` - Lean core (~85 lines)
- `knowledge-base/agent-context/*.md` - Detailed context files

---

## Quick Reference

```bash
# Deploy with private files
fazt app deploy ./my-app --to zyt --include-private

# Access private in serverless
var config = fazt.app.private.readJSON('config.json')  # Future API

# Plan 30 new API examples
fazt.app.user.ds.insert('settings', { theme: 'dark' })
fazt.admin.users.delete('fazt_usr_Nf4rFeUfNV2H')
```
