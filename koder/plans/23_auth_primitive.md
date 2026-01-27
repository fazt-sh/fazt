# Plan 23: Auth Primitive

Multi-user authentication for fazt: social logins (OAuth consumer) + cookie-based
subdomain SSO.

## Summary

Add `internal/auth/` package providing:
- **Social Login**: Google, GitHub, Discord, Microsoft OAuth 2.0
- **User Management**: SQLite-backed users with roles
- **Subdomain SSO**: Domain-wide session cookies
- **JS API**: `fazt.auth.getUser()`, `fazt.auth.requireLogin()`

**Scope**: OAuth consumer + SSO. OIDC provider deferred (see `issuer.md` spec).

## Specs

- `koder/ideas/specs/v0.15-identity/README.md` - Overview
- `koder/ideas/specs/v0.15-identity/users.md` - User management
- `koder/ideas/specs/v0.15-identity/oauth.md` - Social login details
- `koder/ideas/specs/v0.15-identity/sso.md` - Cookie-based SSO

## Files to Create

```
internal/auth/
  auth.go          # AuthService - main orchestrator
  providers.go     # OAuth provider configs (Google, GitHub, etc.)
  oauth.go         # OAuth flow handlers (login, callback)
  session.go       # Session management, cookie handling
  users.go         # User CRUD operations
  invites.go       # Invite code generation and redemption
  middleware.go    # Auth middleware for request context
  bindings.go      # JS API (fazt.auth.*)
  handlers.go      # HTTP handlers for /auth/* routes
```

## Files to Modify

| File | Change |
|------|--------|
| `internal/database/migrations/015_auth.sql` | New tables |
| `internal/handlers/routes.go` | Mount /auth/* routes |
| `internal/runtime/handler.go` | Add auth injector |
| `cmd/server/main.go` | Init AuthService |
| `cmd/server/init.go` | Add auth provider setup to init flow |

## Database Schema

```sql
-- Migration 015_auth.sql

-- OAuth providers configuration
CREATE TABLE auth_providers (
    name TEXT PRIMARY KEY,           -- 'google', 'github', 'discord', 'microsoft'
    enabled INTEGER DEFAULT 0,
    client_id TEXT,
    client_secret TEXT,              -- Encrypted
    created_at INTEGER NOT NULL
);

-- Users
CREATE TABLE auth_users (
    id TEXT PRIMARY KEY,             -- UUID
    email TEXT UNIQUE NOT NULL,
    name TEXT,
    picture TEXT,
    provider TEXT NOT NULL,          -- 'google', 'github', 'discord', 'microsoft', 'password'
    provider_id TEXT,                -- External ID (null for password)
    password_hash TEXT,              -- Only if provider='password'
    role TEXT DEFAULT 'user',        -- 'admin', 'user'
    invited_by TEXT,                 -- User ID or 'owner'
    created_at INTEGER NOT NULL,
    last_login INTEGER
);

CREATE INDEX idx_auth_users_email ON auth_users(email);
CREATE INDEX idx_auth_users_provider ON auth_users(provider, provider_id);

-- Sessions
CREATE TABLE auth_sessions (
    token_hash TEXT PRIMARY KEY,     -- SHA-256 of session token
    user_id TEXT NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    last_seen INTEGER
);

CREATE INDEX idx_auth_sessions_user ON auth_sessions(user_id);
CREATE INDEX idx_auth_sessions_expires ON auth_sessions(expires_at);

-- OAuth state tokens (temporary, CSRF protection)
CREATE TABLE auth_states (
    state TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    redirect_to TEXT,
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL
);

-- Invite codes
CREATE TABLE auth_invites (
    code TEXT PRIMARY KEY,
    role TEXT DEFAULT 'user',
    created_by TEXT NOT NULL,        -- User ID or 'owner'
    created_at INTEGER NOT NULL,
    expires_at INTEGER,
    used_by TEXT,
    used_at INTEGER
);
```

## Implementation Phases

### Phase 1: Core Infrastructure
**Goal**: AuthService struct, database, basic user operations

1. Create migration `015_auth.sql`

2. Create `internal/auth/auth.go`:
   - `AuthService` struct with db reference
   - `NewAuthService(db)` constructor
   - Configuration loading from `auth_providers` table

3. Create `internal/auth/users.go`:
   - `CreateUser(email, name, picture, provider, providerID)`
   - `GetUserByID(id)`
   - `GetUserByEmail(email)`
   - `GetUserByProvider(provider, providerID)`
   - `UpdateLastLogin(id)`
   - `ListUsers()` (admin only)
   - `UpdateUserRole(id, role)`
   - `DeleteUser(id)`

4. Wire into `cmd/server/main.go`:
   - Init after database setup
   - Store in server context

### Phase 2: OAuth Providers
**Goal**: Configure and use Google, GitHub, Discord, Microsoft

1. Create `internal/auth/providers.go`:
   ```go
   type OAuthProvider struct {
       Name        string
       AuthURL     string
       TokenURL    string
       UserInfoURL string
       Scopes      []string
       ParseUser   func(data []byte) (*UserInfo, error)
   }

   var Providers = map[string]OAuthProvider{
       "google":    {...},
       "github":    {...},
       "discord":   {...},
       "microsoft": {...},
   }
   ```

2. Provider details:
   | Provider | Auth URL | Token URL | UserInfo URL |
   |----------|----------|-----------|--------------|
   | Google | accounts.google.com/o/oauth2/v2/auth | oauth2.googleapis.com/token | openidconnect.googleapis.com/v1/userinfo |
   | GitHub | github.com/login/oauth/authorize | github.com/login/oauth/access_token | api.github.com/user + /user/emails |
   | Discord | discord.com/api/oauth2/authorize | discord.com/api/oauth2/token | discord.com/api/users/@me |
   | Microsoft | login.microsoftonline.com/.../authorize | login.microsoftonline.com/.../token | graph.microsoft.com/oidc/userinfo |

3. Add CLI commands in `cmd/server/auth.go`:
   ```bash
   fazt auth provider google --client-id XXX --client-secret YYY
   fazt auth provider google --enable
   fazt auth provider google --disable
   fazt auth providers  # List configured
   ```

### Phase 3: OAuth Flow
**Goal**: Login and callback handlers

1. Create `internal/auth/oauth.go`:
   - `StartOAuthFlow(provider, redirectTo)` - generates state, returns auth URL
   - `HandleCallback(provider, code, state)` - exchanges code, gets user info
   - `ExchangeCode(provider, code)` - server-to-server token exchange
   - `FetchUserInfo(provider, accessToken)` - get user profile

2. Create `internal/auth/handlers.go`:
   ```go
   // GET /auth/login/:provider - Start OAuth flow
   // GET /auth/callback/:provider - Handle OAuth callback
   // GET /auth/login - Render login page with provider buttons
   // POST /auth/logout - Clear session
   // GET /auth/session - Get current session info (API)
   ```

3. Create `internal/auth/states.go`:
   - State token generation (32 bytes random)
   - State validation and cleanup
   - 10-minute expiry

### Phase 4: Sessions & SSO
**Goal**: Cookie-based sessions that work across subdomains

1. Create `internal/auth/session.go`:
   - `CreateSession(userID)` - generates token, stores hash, returns cookie
   - `ValidateSession(tokenHash)` - checks expiry, updates last_seen
   - `DeleteSession(tokenHash)` - logout
   - `CleanExpiredSessions()` - background cleanup

2. Cookie configuration:
   ```go
   func (a *AuthService) SessionCookie(token string) *http.Cookie {
       return &http.Cookie{
           Name:     "fazt_session",
           Value:    token,
           Domain:   "." + a.domain,  // .abc.com for all subdomains
           Path:     "/",
           HttpOnly: true,
           Secure:   a.secure,        // true if HTTPS
           SameSite: http.SameSiteLaxMode,
           MaxAge:   30 * 24 * 3600,  // 30 days
       }
   }
   ```

3. Create `internal/auth/middleware.go`:
   - Read session cookie from request
   - Validate session, get user
   - Inject user into request context
   - `GetUserFromContext(ctx)` helper

### Phase 5: JS Runtime API
**Goal**: Apps can check auth status and require login

1. Create `internal/auth/bindings.go`:
   ```go
   func InjectAuthNamespace(vm *goja.Runtime, ctx RequestContext) {
       auth := vm.NewObject()
       auth.Set("getUser", func() goja.Value {...})
       auth.Set("isLoggedIn", func() bool {...})
       auth.Set("isOwner", func() bool {...})
       auth.Set("hasRole", func(role string) bool {...})
       auth.Set("requireLogin", func() {...})  // throws redirect
       auth.Set("logout", func() {...})

       fazt := vm.Get("fazt").ToObject(vm)
       fazt.Set("auth", auth)
   }
   ```

2. Update `internal/runtime/handler.go`:
   - Add auth injector to `ExecuteWithInjectors()`

3. JS API:
   ```javascript
   // Get current user (null if not logged in)
   const user = await fazt.auth.getUser();
   // { id, email, name, picture, role, provider }

   // Check login status
   if (await fazt.auth.isLoggedIn()) { ... }

   // Check if owner (Persona)
   if (await fazt.auth.isOwner()) { ... }

   // Check role
   if (await fazt.auth.hasRole('admin')) { ... }

   // Require login - redirects if not authenticated
   await fazt.auth.requireLogin();

   // Logout
   await fazt.auth.logout();
   ```

### Phase 6: Invite System
**Goal**: Owner/admin can invite users (no email needed)

1. Create `internal/auth/invites.go`:
   - `CreateInvite(role, createdBy)` - generates code
   - `GetInvite(code)` - validates, checks expiry
   - `RedeemInvite(code, userID)` - marks as used
   - `ListInvites()` - admin view

2. Add handlers:
   ```go
   // POST /auth/invite - Create invite (admin)
   // GET /auth/invite/:code - Render signup form
   // POST /auth/invite/:code - Complete signup with password
   ```

3. Add JS API:
   ```javascript
   // Admin only
   const invite = await fazt.auth.createInvite({ role: 'user' });
   // { code: 'abc123', url: 'https://abc.com/auth/invite/abc123' }

   const invites = await fazt.auth.listInvites();
   ```

### Phase 7: Login UI
**Goal**: Default login page

1. Create embedded HTML template for `/auth/login`:
   ```html
   <div class="auth-container">
     <h1>Sign in to {{.Domain}}</h1>
     <div class="providers">
       {{range .Providers}}
       <a href="/auth/login/{{.Name}}" class="btn {{.Name}}">
         Continue with {{.DisplayName}}
       </a>
       {{end}}
     </div>
   </div>
   ```

2. Minimal CSS (embedded):
   - Clean, centered layout
   - Provider-colored buttons
   - Mobile responsive

3. Error page at `/auth/error?reason=xxx`

### Phase 8: Admin Operations
**Goal**: CLI and API for user management

1. Add CLI commands:
   ```bash
   fazt auth users                    # List users
   fazt auth user <id>                # Show user details
   fazt auth user <id> --role admin   # Change role
   fazt auth user <id> --delete       # Remove user
   fazt auth invite --role user       # Create invite
   fazt auth invites                  # List invites
   ```

2. Add admin API endpoints:
   ```go
   // GET /auth/users - List users (admin)
   // GET /auth/users/:id - Get user (admin)
   // PATCH /auth/users/:id - Update user (admin)
   // DELETE /auth/users/:id - Delete user (admin)
   ```

## HTTP Routes Summary

| Route | Method | Description | Auth |
|-------|--------|-------------|------|
| `/auth/login` | GET | Login page with providers | Public |
| `/auth/login/:provider` | GET | Start OAuth flow | Public |
| `/auth/callback/:provider` | GET | OAuth callback | Public |
| `/auth/logout` | POST | Clear session | User |
| `/auth/session` | GET | Current session info | Public |
| `/auth/invite` | POST | Create invite | Admin |
| `/auth/invite/:code` | GET | Invite signup form | Public |
| `/auth/invite/:code` | POST | Complete invite signup | Public |
| `/auth/users` | GET | List users | Admin |
| `/auth/users/:id` | GET | Get user | Admin |
| `/auth/users/:id` | PATCH | Update user | Admin |
| `/auth/users/:id` | DELETE | Delete user | Admin |

## Key Patterns to Reuse

| Pattern | Source | Reuse For |
|---------|--------|-----------|
| HTTP handlers | `internal/handlers/apps.go` | Auth handlers |
| JSON responses | `internal/api/response.go` | API responses |
| JS bindings | `internal/storage/bindings.go` | Auth bindings |
| Request context | `internal/runtime/handler.go` | User injection |
| Middleware chain | `internal/handlers/middleware.go` | Auth middleware |
| Config encryption | `internal/config/config.go` | Client secrets |

## Verification

### Unit Tests

```bash
go test ./internal/auth/...
```

Test files:
- `users_test.go` - CRUD operations
- `session_test.go` - Create, validate, expire
- `oauth_test.go` - State tokens, code exchange (mocked)
- `invites_test.go` - Create, redeem, expire

### Manual Tests

1. **Provider setup**:
   ```bash
   fazt auth provider google \
     --client-id "xxx.apps.googleusercontent.com" \
     --client-secret "GOCSPX-xxx"
   fazt auth provider google --enable
   fazt auth providers
   ```

2. **OAuth flow**:
   - Visit `https://abc.com/auth/login`
   - Click "Continue with Google"
   - Approve on Google
   - Verify redirected back, session cookie set
   - Check `fazt auth users` shows new user

3. **SSO across subdomains**:
   - Login at `abc.com`
   - Visit `app1.abc.com`
   - Verify still logged in (same cookie)

4. **JS API**:
   ```javascript
   // api/whoami.js
   async function handler(req) {
       const user = await fazt.auth.getUser();
       return {
           status: 200,
           body: JSON.stringify(user || { message: 'not logged in' })
       };
   }
   ```

5. **Require login**:
   ```javascript
   // api/protected.js
   async function handler(req) {
       await fazt.auth.requireLogin();
       const user = await fazt.auth.getUser();
       return {
           status: 200,
           body: JSON.stringify({ welcome: user.name })
       };
   }
   ```

6. **Invite flow**:
   ```bash
   fazt auth invite --role user
   # Output: https://abc.com/auth/invite/abc123
   ```
   - Visit invite URL
   - Set password
   - Verify login works

## Dependencies

- `golang.org/x/oauth2` - OAuth 2.0 client
- `golang.org/x/crypto/bcrypt` - Password hashing (already in use)
- No new external dependencies needed

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Secret exposure in logs | Never log client secrets, mask in CLI output |
| Session fixation | Generate new token on login, not before |
| CSRF on OAuth | State token validation, 10-min expiry |
| Cookie theft | HttpOnly, Secure, SameSite=Lax |
| Timing attacks on session | Use constant-time comparison |
| SQLite contention | Use write queue for session updates |

## Out of Scope

- OIDC provider (issuing identities) - see `issuer.md`, implement later
- Email verification - not needed with social login
- Password reset via email - manual reset by owner
- Rate limiting on login - personal scale assumption
- MFA/2FA - future enhancement
- Apple Sign-In - requires $99/yr developer account
- Twitter/X - requires $100/mo API access

## Estimated Effort

| Phase | Lines (Go) | Effort |
|-------|------------|--------|
| Phase 1: Core | ~200 | Small |
| Phase 2: Providers | ~150 | Small |
| Phase 3: OAuth flow | ~250 | Medium |
| Phase 4: Sessions | ~200 | Medium |
| Phase 5: JS API | ~150 | Small |
| Phase 6: Invites | ~150 | Small |
| Phase 7: Login UI | ~100 | Small |
| Phase 8: Admin | ~150 | Small |
| **Total** | **~1,350** | |

Plus tests: ~400 lines
