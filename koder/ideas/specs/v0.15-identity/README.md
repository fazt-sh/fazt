# v0.15 - Identity

**Theme**: Sovereign identity, multi-user support, and authentication.

## Summary

v0.15 evolves fazt from a personal server to a single-owner compute node that
can serve multiple users. The owner retains full control via the hardware-bound
Persona, while users authenticate via social logins and share sessions across
all subdomain apps.

## Documents

| Document | Status | Description |
|----------|--------|-------------|
| `persona.md` | Implement | Hardware-bound owner identity |
| `sso.md` | Implement | Cookie-based subdomain SSO |
| `users.md` | Implement | Multi-user support |
| `oauth.md` | Implement | Social login (Google, GitHub, etc.) |
| `issuer.md` | Spec Only | OIDC provider (future) |

## Key Capabilities

### Owner (Persona)

- Cryptographic identity tied to kernel
- Not a password—a keypair
- Signs assertions for verification
- Full admin access to all apps

### Users

- Managed by owner, stored in SQLite
- Authenticate via social providers (Google, GitHub, Discord)
- Or via password + invite code (no email required)
- Roles: `owner`, `admin`, `user`

### Social Login (OAuth Consumer)

- Accept Google, GitHub, Discord logins
- No email verification needed (providers handle it)
- Frictionless signup for app users

### Subdomain SSO

- Login at `abc.com` propagates to all subdomains
- Cookie-based (`.abc.com` domain cookie)
- No OAuth dance between fazt apps
- Apps call `fazt.auth.getUser()` to get current user

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Apps: notion.abc.com, chat.abc.com, docs.abc.com          │
│  JS Runtime: fazt.auth.getUser(), fazt.auth.requireLogin() │
└──────────────────────────┬──────────────────────────────────┘
                           │ cookie-based SSO
┌──────────────────────────▼──────────────────────────────────┐
│  Auth Layer (abc.com)                                       │
│  - Social login handlers (Google, GitHub, Discord)          │
│  - Session management (domain-wide cookies)                 │
│  - User storage (SQLite)                                    │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│  Kernel                                                     │
│  - Persona (owner identity)                                 │
│  - Session validation middleware                            │
└─────────────────────────────────────────────────────────────┘
```

## JS Runtime API

```javascript
// Get current user (works on any subdomain)
const user = await fazt.auth.getUser();
// { id, email, name, picture, role, provider }

// Check if logged in
if (await fazt.auth.isLoggedIn()) { ... }

// Require login (redirects if not authenticated)
await fazt.auth.requireLogin();

// Check role
if (await fazt.auth.hasRole('admin')) { ... }

// Check if current user is owner
if (await fazt.auth.isOwner()) { ... }

// Logout
await fazt.auth.logout();
```

## Future: OIDC Provider (v0.18+)

Spec'd in `issuer.md` but NOT implementing in v0.15.

When needed:
- Mobile apps that can't share cookies
- Third-party integrations
- "Sign in with ABC Corp" for external sites

## Dependencies

- v0.14 (Security): Persona, session infrastructure
