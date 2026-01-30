# Mock OAuth Provider - Feature Spec

**Status**: Proposed
**Purpose**: Enable full auth flow testing locally without code changes

## Problem

OAuth requires HTTPS + real domain for callbacks. Local development can't use
real OAuth providers, forcing developers to either:
- Test auth only on remote (slow iteration)
- Add mock code paths (diverges from production)
- Skip auth testing locally (risky)

## Solution

Add a built-in `dev` OAuth provider that simulates the OAuth flow locally.

## How It Works

### User Flow
```
1. User clicks "Sign in with Google" (or any provider)
2. Fazt detects non-HTTPS → offers "Dev Login" option
3. User sees simple form: email, name (pre-filled with dev@example.com)
4. User submits → Fazt creates real session with auth cookie
5. Redirect back to app with valid session
6. fazt.auth.getUser() returns the mock user
```

### Key Properties

- **Same code path**: No conditional logic in app code
- **Real session**: Same cookie mechanism as production OAuth
- **Same API response**: `fazt.auth.getUser()` returns identical structure
- **Automatic**: Only appears on non-HTTPS instances
- **No configuration**: Works out of the box locally

## Implementation Notes

### Detection
```go
// In auth handlers
func isLocalMode(r *http.Request) bool {
    return r.TLS == nil // No HTTPS = local mode
}
```

### Provider Entry
```go
// Add to Providers map
"dev": {
    Name:        "dev",
    DisplayName: "Dev Login",
    // No external URLs - handled internally
}
```

### Login Form (served by fazt)
```html
<form action="/auth/callback/dev" method="POST">
  <input name="email" value="dev@example.com">
  <input name="name" value="Dev User">
  <button type="submit">Sign In</button>
</form>
```

### Session Creation
```go
// In dev callback handler
func handleDevCallback(w http.ResponseWriter, r *http.Request) {
    email := r.FormValue("email")
    name := r.FormValue("name")

    // Create user if not exists (same as real OAuth)
    user, _ := s.FindOrCreateUser(&UserInfo{
        ID:    "dev_" + hashEmail(email),
        Email: email,
        Name:  name,
    }, "dev")

    // Create session (same as real OAuth)
    session := s.CreateUserSession(user.ID, appID)
    setSessionCookie(w, session)

    // Redirect back (same as real OAuth)
    http.Redirect(w, r, redirectTo, http.StatusFound)
}
```

## Security

- **Local only**: Dev provider only available when `r.TLS == nil`
- **No secrets**: No client ID/secret needed
- **Obvious**: Login page clearly shows "Development Mode"
- **Not in production**: Automatically disabled on HTTPS

## User Experience

### Login Page (Local)
```
┌─────────────────────────────────┐
│     Sign in to MyApp            │
│                                 │
│  [🔵 Sign in with Google]       │  ← Real provider (won't work locally)
│                                 │
│  ─────── Development ───────    │
│                                 │
│  [🔧 Dev Login]                 │  ← Mock provider (local only)
│                                 │
│  ⚠️ Dev login only works        │
│     on local instances          │
└─────────────────────────────────┘
```

### Login Page (Production)
```
┌─────────────────────────────────┐
│     Sign in to MyApp            │
│                                 │
│  [🔵 Sign in with Google]       │  ← Real provider works
│                                 │
│  (Dev login not available       │
│   on production)                │
└─────────────────────────────────┘
```

## Benefits

1. **Zero code changes**: Same app code works locally and remotely
2. **Full flow testing**: Test protected routes, user data, sessions
3. **Fast iteration**: No need to deploy to test auth
4. **No workarounds**: No mock headers, no conditional code
5. **Obvious**: Clear visual indication of dev mode
