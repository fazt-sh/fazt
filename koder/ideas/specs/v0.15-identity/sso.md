# Subdomain SSO

## Summary

Cookie-based single sign-on across all fazt apps. Login once at the primary
domain, and every subdomain app instantly shares the session. No OAuth dance
between fazt apps—just a domain-wide cookie.

## How It Works

### Traditional OAuth Flow (NOT fazt)

```
1. Visit notion.abc.com
2. Redirect to auth.abc.com/authorize
3. Enter credentials
4. Redirect back with code
5. Exchange code for token
6. Finally authenticated
```

### Fazt SSO Flow

```
1. Visit notion.abc.com
2. Kernel reads session cookie (.abc.com)
3. User is authenticated (session inherited)
```

The key: **One cookie serves all subdomains.**

## Implementation

### Session Cookie

```go
cookie := &http.Cookie{
    Name:     "fazt_session",
    Value:    sessionToken,
    Domain:   ".abc.com",          // Note the dot: all subdomains
    Path:     "/",
    HttpOnly: true,
    Secure:   true,                // HTTPS only
    SameSite: http.SameSiteLaxMode,
    MaxAge:   30 * 24 * 3600,      // 30 days
}
```

### Auth Middleware

The kernel injects user context into every request:

```go
func (k *Kernel) AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w, r *http.Request) {
        // Read session cookie (works on any subdomain)
        session := k.Sessions.Get(r)

        if session != nil && session.Valid() {
            // Inject user into request context
            ctx := context.WithValue(r.Context(), "user", session.User)
            next.ServeHTTP(w, r.WithContext(ctx))
            return
        }

        // No session - proceed without user context
        next.ServeHTTP(w, r)
    })
}
```

### Session Storage

```sql
CREATE TABLE auth_sessions (
    token TEXT PRIMARY KEY,        -- Hashed session token
    user_id TEXT NOT NULL,         -- References auth_users(id)
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    last_seen INTEGER
);
```

## Login Flow

1. User visits `notion.abc.com/dashboard`
2. App calls `fazt.auth.requireLogin()`
3. User is not logged in → redirect to `/auth/login`
4. User clicks "Login with Google"
5. OAuth flow completes at `abc.com/auth/callback/google`
6. Session cookie set for `.abc.com`
7. Redirect back to `notion.abc.com/dashboard`
8. Cookie is now readable → user is authenticated

## JS Runtime API

```javascript
// Get current user (works on ANY subdomain)
const user = await fazt.auth.getUser();
// { id, email, name, picture, role, provider }

// Require login (redirects if not authenticated)
await fazt.auth.requireLogin();
// After this line, user is guaranteed to be logged in

// Check if owner (Persona, not a regular user)
if (await fazt.auth.isOwner()) {
    // Show admin features
}

// Logout (clears cookie for all subdomains)
await fazt.auth.logout();
```

## Example: Protected App

```javascript
// api/dashboard.js
async function handler(req) {
    // Require login - redirects if not authenticated
    await fazt.auth.requireLogin();

    // Get user (guaranteed to exist after requireLogin)
    const user = await fazt.auth.getUser();

    return {
        status: 200,
        body: JSON.stringify({
            message: `Welcome, ${user.name}!`,
            role: user.role
        })
    };
}
```

## Benefits

| Benefit | Description |
|---------|-------------|
| **Zero Config** | Apps don't configure auth—kernel handles it |
| **Seamless UX** | Login once, access everything |
| **No Secrets** | Apps never see credentials |
| **Unified** | Same user identity everywhere |
| **Simple** | Just cookies, no token exchange |

## Limitations

This approach works for:
- ✓ Subdomain apps (notion.abc.com, chat.abc.com)
- ✓ Same fazt instance

Does NOT work for:
- ✗ External domains (different-site.com)
- ✗ Mobile apps (can't share browser cookies)
- ✗ Third-party integrations

For those cases, see `issuer.md` (OIDC provider, future spec).
