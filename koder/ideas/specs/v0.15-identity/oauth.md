# OAuth (Social Login)

## Summary

Fazt as an OAuth consumer—accepting Google, GitHub, and Discord logins. Users
authenticate with their existing accounts, no email verification needed.

## Supported Providers

| Provider | Scopes | What We Get |
|----------|--------|-------------|
| Google | `openid email profile` | email, name, picture |
| GitHub | `user:email` | email, name, avatar |
| Discord | `identify email` | email, username, avatar |

Three providers cover 95%+ of use cases. More can be added later.

## Configuration

Providers are configured in the database, set during `fazt init` or via admin:

```sql
CREATE TABLE auth_providers (
    name TEXT PRIMARY KEY,         -- 'google', 'github', 'discord'
    enabled INTEGER DEFAULT 0,
    client_id TEXT,
    client_secret TEXT,            -- Encrypted at rest
    created_at INTEGER
);
```

### Setup Flow

```bash
# During init or later
fazt auth provider google \
  --client-id "xxx.apps.googleusercontent.com" \
  --client-secret "GOCSPX-xxx"

# Enable/disable
fazt auth provider google --enable
fazt auth provider google --disable

# List configured providers
fazt auth providers
```

## OAuth Flow

### 1. Initiate Login

```
GET /auth/login/google
```

Fazt generates state token, stores in `auth_states`, redirects to Google:

```
https://accounts.google.com/o/oauth2/v2/auth?
  client_id=xxx&
  redirect_uri=https://abc.com/auth/callback/google&
  response_type=code&
  scope=openid%20email%20profile&
  state=random-state-token
```

### 2. Handle Callback

```
GET /auth/callback/google?code=xxx&state=yyy
```

Fazt:
1. Validates state token (CSRF protection)
2. Exchanges code for access token (server-to-server)
3. Fetches user info from provider
4. Creates/updates user in `auth_users`
5. Creates session, sets cookie
6. Redirects to original destination

### 3. Session Cookie

```go
cookie := &http.Cookie{
    Name:     "fazt_session",
    Value:    sessionToken,
    Domain:   ".abc.com",      // All subdomains
    Path:     "/",
    HttpOnly: true,
    Secure:   true,            // HTTPS only
    SameSite: http.SameSiteLaxMode,
    MaxAge:   30 * 24 * 3600,  // 30 days
}
```

## State Management

Temporary state tokens prevent CSRF attacks:

```sql
CREATE TABLE auth_states (
    state TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    redirect_to TEXT,          -- Where to go after login
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL
);
```

States expire after 10 minutes. Cleanup job removes expired states.

## Sessions

```sql
CREATE TABLE auth_sessions (
    token TEXT PRIMARY KEY,
    user_id TEXT REFERENCES auth_users(id),
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    last_seen INTEGER
);

CREATE INDEX idx_auth_sessions_user ON auth_sessions(user_id);
CREATE INDEX idx_auth_sessions_expires ON auth_sessions(expires_at);
```

Session tokens are:
- 32 bytes of random data, base64 encoded
- Stored hashed in database (bcrypt or SHA-256)
- Cookie contains raw token, DB contains hash

## Provider Details

### Google

```go
type GoogleProvider struct {
    AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth"
    TokenURL:    "https://oauth2.googleapis.com/token"
    UserInfoURL: "https://openidconnect.googleapis.com/v1/userinfo"
    Scopes:      []string{"openid", "email", "profile"}
}
```

Response:
```json
{
  "sub": "1234567890",
  "email": "alice@gmail.com",
  "name": "Alice Smith",
  "picture": "https://lh3.googleusercontent.com/..."
}
```

### GitHub

```go
type GitHubProvider struct {
    AuthURL:     "https://github.com/login/oauth/authorize"
    TokenURL:    "https://github.com/login/oauth/access_token"
    UserInfoURL: "https://api.github.com/user"
    EmailURL:    "https://api.github.com/user/emails"
    Scopes:      []string{"user:email"}
}
```

Response (two calls needed):
```json
// /user
{
  "id": 123456,
  "login": "alice",
  "name": "Alice Smith",
  "avatar_url": "https://avatars.githubusercontent.com/..."
}

// /user/emails (find primary)
[
  { "email": "alice@example.com", "primary": true, "verified": true }
]
```

### Discord

```go
type DiscordProvider struct {
    AuthURL:     "https://discord.com/api/oauth2/authorize"
    TokenURL:    "https://discord.com/api/oauth2/token"
    UserInfoURL: "https://discord.com/api/users/@me"
    Scopes:      []string{"identify", "email"}
}
```

Response:
```json
{
  "id": "123456789",
  "username": "alice",
  "email": "alice@example.com",
  "avatar": "abc123"
}
```

Avatar URL: `https://cdn.discordapp.com/avatars/{id}/{avatar}.png`

## HTTP Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/auth/login/:provider` | GET | Start OAuth flow |
| `/auth/callback/:provider` | GET | Handle OAuth callback |
| `/auth/logout` | POST | Clear session |
| `/auth/session` | GET | Get current session info |

## Login UI

Fazt provides a default login page at `/auth/login`:

```html
<!-- Rendered by fazt, customizable via templates -->
<div class="auth-providers">
  <a href="/auth/login/google" class="btn google">
    Continue with Google
  </a>
  <a href="/auth/login/github" class="btn github">
    Continue with GitHub
  </a>
  <a href="/auth/login/discord" class="btn discord">
    Continue with Discord
  </a>
</div>
```

Apps can use this default or build custom UI that links to the same endpoints.

## JS Runtime Integration

```javascript
// In an app's API handler
async function handler(req) {
  const user = await fazt.auth.getUser();

  if (!user) {
    // Not logged in - redirect to login
    return {
      status: 302,
      headers: { Location: '/auth/login?redirect=' + req.url }
    };
  }

  return {
    status: 200,
    body: JSON.stringify({ message: `Hello, ${user.name}!` })
  };
}
```

## Error Handling

| Error | Response |
|-------|----------|
| Invalid state | 400 Bad Request |
| Provider error | 502 Bad Gateway (log details) |
| User creation failed | 500 Internal Server Error |
| Provider not configured | 404 Not Found |

Errors redirect to `/auth/error?reason=xxx` with user-friendly messages.

## Security Considerations

1. **State tokens** - Prevent CSRF attacks
2. **HTTPS only** - Secure cookie flag
3. **HttpOnly cookies** - No JS access to session
4. **Secret encryption** - Client secrets encrypted in DB
5. **Token hashing** - Session tokens stored as hashes
6. **Short-lived codes** - OAuth codes expire quickly (provider-enforced)

## Implementation Notes

Use Go's `golang.org/x/oauth2` package:

```go
import "golang.org/x/oauth2"
import "golang.org/x/oauth2/google"
import "golang.org/x/oauth2/github"

// Config for each provider
googleConfig := &oauth2.Config{
    ClientID:     cfg.ClientID,
    ClientSecret: cfg.ClientSecret,
    RedirectURL:  "https://abc.com/auth/callback/google",
    Scopes:       []string{"openid", "email", "profile"},
    Endpoint:     google.Endpoint,
}
```

Estimated implementation: ~500 lines of Go.
