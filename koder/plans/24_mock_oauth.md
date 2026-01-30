# Plan 24: Mock OAuth Provider

**Status**: Complete (2026-01-31)
**Goal**: Enable full auth flow testing locally without code changes

## Problem

OAuth requires HTTPS + real domain. Local development can't test auth flows,
forcing developers to:
- Deploy to remote just to test auth (slow)
- Add mock code paths (diverges from production)
- Skip auth testing locally (risky)

## Solution

Add a `dev` OAuth provider that simulates the flow locally.

**Key property**: Zero code changes when deploying to remote.

```
Local:  "Sign in" → Dev form → Session → fazt.auth.getUser() ✓
Remote: "Sign in" → Google   → Session → fazt.auth.getUser() ✓
```

Same code. Same API. Different provider.

## User Experience

### Local Login Page

```
┌─────────────────────────────────────┐
│       Sign in to MyApp              │
│                                     │
│  ┌─────────────────────────────┐    │
│  │  🔵 Sign in with Google     │    │  ← Disabled/grayed (no HTTPS)
│  └─────────────────────────────┘    │
│                                     │
│  ──────── Development ────────      │
│                                     │
│  ┌─────────────────────────────┐    │
│  │  🔧 Dev Login               │    │  ← Available locally
│  └─────────────────────────────┘    │
│                                     │
│  ⚠️ Dev login simulates OAuth       │
│     for local testing only          │
└─────────────────────────────────────┘
```

### Dev Login Form

```
┌─────────────────────────────────────┐
│       Development Login             │
│                                     │
│  Email:  [dev@example.com      ]    │
│  Name:   [Dev User             ]    │
│  Role:   [user ▼]                   │
│          (user / admin / owner)     │
│                                     │
│  ┌─────────────────────────────┐    │
│  │       Sign In               │    │
│  └─────────────────────────────┘    │
│                                     │
│  This creates a real session.       │
│  Same as production OAuth.          │
└─────────────────────────────────────┘
```

### Production Login Page

```
┌─────────────────────────────────────┐
│       Sign in to MyApp              │
│                                     │
│  ┌─────────────────────────────┐    │
│  │  🔵 Sign in with Google     │    │  ← Real OAuth works
│  └─────────────────────────────┘    │
│                                     │
│  (Dev login not available on        │
│   production instances)             │
└─────────────────────────────────────┘
```

## Implementation

### Phase 1: Core Provider

**File**: `internal/auth/dev_provider.go`

```go
package auth

// DevProvider handles mock OAuth for local development
type DevProvider struct{}

// IsLocalMode detects if we're running without HTTPS
func IsLocalMode(r *http.Request) bool {
    // No TLS = local mode
    if r.TLS == nil {
        return true
    }
    // Also check for known local patterns
    host := r.Host
    return strings.HasPrefix(host, "localhost") ||
           strings.HasPrefix(host, "127.0.0.1") ||
           strings.Contains(host, ".nip.io") ||
           strings.Contains(host, ".local")
}

// DevLoginPageData for template
type DevLoginPageData struct {
    AppName     string
    RedirectTo  string
    DefaultEmail string
    DefaultName  string
}
```

### Phase 2: Routes

Add to auth handlers:

```go
// GET /auth/dev/login - Show dev login form
func (h *AuthHandlers) DevLoginForm(w http.ResponseWriter, r *http.Request) {
    if !IsLocalMode(r) {
        http.Error(w, "Dev login only available locally", http.StatusForbidden)
        return
    }

    redirectTo := r.URL.Query().Get("redirect")
    if redirectTo == "" {
        redirectTo = "/"
    }

    data := DevLoginPageData{
        AppName:      getAppName(r),
        RedirectTo:   redirectTo,
        DefaultEmail: "dev@example.com",
        DefaultName:  "Dev User",
    }

    renderDevLoginPage(w, data)
}

// POST /auth/dev/callback - Process dev login
func (h *AuthHandlers) DevLoginCallback(w http.ResponseWriter, r *http.Request) {
    if !IsLocalMode(r) {
        http.Error(w, "Dev login only available locally", http.StatusForbidden)
        return
    }

    email := r.FormValue("email")
    name := r.FormValue("name")
    role := r.FormValue("role")
    redirectTo := r.FormValue("redirect")

    if email == "" {
        email = "dev@example.com"
    }
    if name == "" {
        name = "Dev User"
    }
    if role == "" {
        role = "user"
    }

    // Create user (same as real OAuth)
    userInfo := &UserInfo{
        ID:      "dev_" + hashString(email),
        Email:   email,
        Name:    name,
        Picture: "", // Could generate gravatar URL
    }

    user, err := h.service.FindOrCreateUser(userInfo, "dev")
    if err != nil {
        http.Error(w, "Failed to create user", http.StatusInternalServerError)
        return
    }

    // Set role if specified
    if role != "user" {
        h.service.UpdateUserRole(user.ID, role)
    }

    // Create session (same as real OAuth)
    appID := getAppID(r)
    session, err := h.service.CreateUserSession(user.ID, appID)
    if err != nil {
        http.Error(w, "Failed to create session", http.StatusInternalServerError)
        return
    }

    // Set cookie (same as real OAuth)
    setAuthCookie(w, r, session.Token)

    // Redirect back (same as real OAuth)
    if redirectTo == "" {
        redirectTo = "/"
    }
    http.Redirect(w, r, redirectTo, http.StatusFound)
}
```

### Phase 3: Login Page Integration

Modify the login page template to show dev option:

```html
<!-- templates/auth/login.html -->
{{if .Providers}}
  {{range .Providers}}
    <a href="/auth/login/{{.Name}}?redirect={{$.Redirect}}"
       class="btn btn-provider {{if not $.IsHTTPS}}disabled{{end}}">
      Sign in with {{.DisplayName}}
    </a>
  {{end}}
{{end}}

{{if .IsLocalMode}}
  <div class="divider">Development</div>
  <a href="/auth/dev/login?redirect={{.Redirect}}" class="btn btn-dev">
    🔧 Dev Login
  </a>
  <p class="hint">Simulates OAuth for local testing</p>
{{end}}
```

### Phase 4: Dev Login Page

```html
<!-- templates/auth/dev-login.html -->
<!DOCTYPE html>
<html>
<head>
  <title>Dev Login - {{.AppName}}</title>
  <style>
    /* Simple, clean form styling */
    body { font-family: system-ui; max-width: 400px; margin: 50px auto; padding: 20px; }
    .form-group { margin: 16px 0; }
    label { display: block; margin-bottom: 4px; font-weight: 500; }
    input, select { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 6px; }
    button { width: 100%; padding: 12px; background: #3b82f6; color: white; border: none; border-radius: 6px; cursor: pointer; }
    button:hover { background: #2563eb; }
    .warning { background: #fef3c7; border: 1px solid #f59e0b; padding: 12px; border-radius: 6px; margin-bottom: 20px; }
  </style>
</head>
<body>
  <h1>🔧 Dev Login</h1>

  <div class="warning">
    ⚠️ Development mode only. Creates a real session for testing.
  </div>

  <form action="/auth/dev/callback" method="POST">
    <input type="hidden" name="redirect" value="{{.RedirectTo}}">

    <div class="form-group">
      <label>Email</label>
      <input type="email" name="email" value="{{.DefaultEmail}}" required>
    </div>

    <div class="form-group">
      <label>Name</label>
      <input type="text" name="name" value="{{.DefaultName}}" required>
    </div>

    <div class="form-group">
      <label>Role</label>
      <select name="role">
        <option value="user">User</option>
        <option value="admin">Admin</option>
        <option value="owner">Owner</option>
      </select>
    </div>

    <button type="submit">Sign In</button>
  </form>
</body>
</html>
```

## Route Summary

| Route | Method | Purpose |
|-------|--------|---------|
| `/auth/dev/login` | GET | Show dev login form |
| `/auth/dev/callback` | POST | Process dev login |

## Security

### Local-Only Enforcement

```go
func IsLocalMode(r *http.Request) bool {
    // Primary check: no TLS
    if r.TLS == nil {
        return true
    }

    // Secondary: known local domains
    host := strings.ToLower(r.Host)
    localPatterns := []string{
        "localhost",
        "127.0.0.1",
        ".nip.io",
        ".local",
        ".internal",
    }

    for _, pattern := range localPatterns {
        if strings.Contains(host, pattern) {
            return true
        }
    }

    return false
}
```

### Explicit Indicator

- Dev login page shows clear warning
- User created with `provider: "dev"` (visible in admin)
- Session marked as dev session (optional)

### No Secrets

- No client ID or secret needed
- No external OAuth flow
- Self-contained

## Testing

### Unit Tests

```go
func TestIsLocalMode(t *testing.T) {
    tests := []struct {
        host     string
        tls      bool
        expected bool
    }{
        {"localhost:8080", false, true},
        {"127.0.0.1:8080", false, true},
        {"app.192.168.1.1.nip.io", false, true},
        {"example.com", true, false},
        {"app.example.com", true, false},
    }
    // ...
}

func TestDevLoginCreatesSession(t *testing.T) {
    // POST to /auth/dev/callback
    // Verify session cookie set
    // Verify user created
    // Verify redirect works
}
```

### Integration Tests

```bash
# Start local server
fazt server start --port 8080 --db test.db

# Test dev login flow
curl -c cookies.txt -X POST \
  -d "email=test@example.com&name=Test&role=user&redirect=/" \
  http://localhost:8080/auth/dev/callback

# Verify session works
curl -b cookies.txt http://localhost:8080/api/me
# Should return user info
```

## Files to Create/Modify

| File | Action |
|------|--------|
| `internal/auth/dev_provider.go` | **Create** - Dev provider logic |
| `internal/auth/handlers.go` | **Modify** - Add dev routes |
| `internal/assets/templates/auth/dev-login.html` | **Create** - Login form |
| `internal/assets/templates/auth/login.html` | **Modify** - Add dev option |
| `internal/auth/handlers_test.go` | **Modify** - Add tests |

## Rollout

1. **Phase 1**: Core implementation (dev provider + routes)
2. **Phase 2**: UI integration (login page shows dev option)
3. **Phase 3**: Documentation (update skill docs)
4. **Phase 4**: CLI hint (`fazt server start` shows dev login available)

## Success Criteria

- [ ] Dev login form accessible at `/auth/dev/login` (local only)
- [ ] Submitting form creates real session
- [ ] `fazt.auth.getUser()` returns dev user
- [ ] Same code works on remote with real OAuth
- [ ] Dev provider blocked on HTTPS/production
- [ ] Clear visual indication of dev mode

## Open Questions

1. **Persist dev users?** Yes - same as real users, just `provider: "dev"`
2. **Allow role selection?** Yes - useful for testing admin/owner flows
3. **Show on main login page?** Yes - with clear "Development" divider
4. **Gravatar for dev users?** Optional - nice to have
