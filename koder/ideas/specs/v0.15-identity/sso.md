# Sovereign SSO

## Summary

Zero-handshake single sign-on across all apps. Authenticate once at
`os.<domain>`, and every subdomain app instantly knows who you are.

## How It Works

### Traditional OAuth Flow

```
1. Visit blog.example.com
2. Redirect to auth.example.com/authorize
3. Enter credentials
4. Redirect back with code
5. Exchange code for token
6. Finally authenticated
```

### Fazt SSO Flow

```
1. Visit blog.example.com
2. Kernel checks session
3. Authenticated (session inherited)
```

## Implementation

### Session Inheritance

The kernel manages sessions at the domain level:

```go
func (k *Kernel) AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w, r *http.Request) {
        // Check session cookie (domain-wide)
        session := k.Sessions.Get(r)

        if session.Valid() {
            // Inject persona into request context
            ctx := context.WithValue(r.Context(), "persona", session.Persona)
            next.ServeHTTP(w, r.WithContext(ctx))
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

### Cookie Configuration

```go
cookie := &http.Cookie{
    Name:     "fazt_session",
    Domain:   ".example.com",  // Note the dot: all subdomains
    Path:     "/",
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteLaxMode,
}
```

## JS Runtime

```javascript
// Force login if not authenticated
await fazt.security.requireAuth();
// Redirects to os.<domain>/login if needed
// Returns to current page after login

// Get current user (works on any subdomain)
const persona = await fazt.security.getPersona();
```

## Benefits

1. **Zero Config**: Apps don't need auth code
2. **Seamless UX**: Login once, access everything
3. **No Secrets**: Apps don't store credentials
4. **Unified**: Same identity everywhere
