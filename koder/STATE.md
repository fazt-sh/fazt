# Fazt Implementation State

**Last Updated**: 2026-01-27
**Current Version**: v0.10.15

## Status

State: PLANNING - Auth primitive designed, ready for implementation

---

## Last Session

**Auth Primitive Design**

1. **Vision Evolution**
   - Fazt: single-owner compute node that can serve multiple users
   - Not multi-tenant, but owner can run apps with users (chat, docs, etc.)
   - Each fazt instance is sovereign, sets up their own OAuth providers

2. **OpenAuth Analysis**
   - Reviewed ~/Projects/openauth as reference
   - Decision: Extract patterns, implement in pure Go
   - No runtime dependency on Node.js

3. **Specs Created/Updated** (`koder/ideas/specs/v0.15-identity/`)
   - `README.md` - Updated with architecture diagram
   - `sso.md` - Clarified cookie-based approach
   - `users.md` - NEW: Multi-user support, roles, invites
   - `oauth.md` - NEW: Social login (Google, GitHub, Discord, Microsoft)
   - `issuer.md` - NEW: OIDC provider (SPEC ONLY, deferred)

4. **Implementation Plan**
   - Created `koder/plans/23_auth_primitive.md`
   - ~1,350 lines of Go for full implementation
   - 8 phases from core to admin operations

---

## Key Decisions

| Decision | Rationale |
|----------|-----------|
| Cookie-based SSO | Simpler than OIDC, covers subdomain apps |
| Social login only | No email verification needed |
| OIDC provider deferred | Wait until mobile/third-party needed |
| 4 providers | Google, GitHub, Discord, Microsoft (free, covers 95%) |
| Invite codes | For password users without email |

---

## Quick Reference

```bash
# Read the plan
cat koder/plans/23_auth_primitive.md

# Read the specs
ls koder/ideas/specs/v0.15-identity/
```

---

## Next Up

1. **Implement Plan 23** - Auth primitive
   - Start with Phase 1: Core infrastructure
   - Follow plan phases sequentially
   - Estimated ~1,350 LoC

2. **After auth is done**
   - Build a sample multi-user app to test
   - Consider v0.16 release with auth
