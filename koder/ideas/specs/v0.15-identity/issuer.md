# OIDC Issuer (Identity Provider)

```
┌─────────────────────────────────────────────────────────┐
│  STATUS: SPEC ONLY - NOT IMPLEMENTING IN v0.15         │
│                                                         │
│  Trigger: When mobile apps or third-party integrations │
│           need "Sign in with ABC Corp"                 │
└─────────────────────────────────────────────────────────┘
```

## Summary

Fazt as an OIDC provider—issuing identities that external apps can verify.
This enables "Sign in with ABC Corp" for apps that can't share cookies
with the fazt instance.

## When This Is Needed

| Scenario | Cookie SSO | OIDC Issuer |
|----------|------------|-------------|
| Subdomain apps (notion.abc.com) | ✓ Works | Overkill |
| iOS/Android mobile app | ✗ Can't share | ✓ Needed |
| Third-party SaaS integration | ✗ Different domain | ✓ Needed |
| Partner company access | ✗ | ✓ Needed |

**Don't implement until one of these scenarios is actually needed.**

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  External App (mobile, third-party)                     │
│  "Sign in with ABC Corp"                                │
└────────────────────────────┬────────────────────────────┘
                             │ OIDC flow
┌────────────────────────────▼────────────────────────────┐
│  Fazt OIDC Issuer (abc.com)                             │
│  - Discovery: /.well-known/openid-configuration         │
│  - Authorization: /oauth/authorize                      │
│  - Token: /oauth/token                                  │
│  - UserInfo: /oauth/userinfo                            │
│  - JWKS: /.well-known/jwks.json                         │
└────────────────────────────┬────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────┐
│  User Database (same auth_users table)                  │
└─────────────────────────────────────────────────────────┘
```

## OIDC Discovery

```
GET /.well-known/openid-configuration
```

Response:
```json
{
  "issuer": "https://abc.com",
  "authorization_endpoint": "https://abc.com/oauth/authorize",
  "token_endpoint": "https://abc.com/oauth/token",
  "userinfo_endpoint": "https://abc.com/oauth/userinfo",
  "jwks_uri": "https://abc.com/.well-known/jwks.json",
  "response_types_supported": ["code"],
  "subject_types_supported": ["public"],
  "id_token_signing_alg_values_supported": ["RS256"],
  "scopes_supported": ["openid", "email", "profile"],
  "token_endpoint_auth_methods_supported": ["client_secret_basic", "client_secret_post"],
  "claims_supported": ["sub", "email", "name", "picture"]
}
```

## Client Registration

External apps need to be registered:

```sql
CREATE TABLE oauth_clients (
    id TEXT PRIMARY KEY,           -- Client ID (UUID)
    secret_hash TEXT NOT NULL,     -- Hashed client secret
    name TEXT NOT NULL,            -- Display name
    redirect_uris TEXT NOT NULL,   -- JSON array of allowed URIs
    created_by TEXT NOT NULL,      -- User ID or 'owner'
    created_at INTEGER NOT NULL
);
```

### CLI Registration

```bash
# Register a new client
fazt auth client create \
  --name "Mobile App" \
  --redirect-uri "myapp://callback" \
  --redirect-uri "https://myapp.com/callback"

# Output:
# Client ID: abc123
# Client Secret: secret456 (save this, shown only once)

# List clients
fazt auth clients

# Revoke client
fazt auth client revoke abc123
```

## Authorization Flow

### 1. Authorization Request

External app redirects user to:

```
GET /oauth/authorize?
  client_id=abc123&
  redirect_uri=myapp://callback&
  response_type=code&
  scope=openid%20email%20profile&
  state=random-state&
  code_challenge=xxx&           # PKCE
  code_challenge_method=S256
```

### 2. User Consent

If user is logged in, show consent screen:
```
"Mobile App" wants to access your account:
- Email address
- Profile information

[Allow] [Deny]
```

If not logged in, redirect to login first.

### 3. Authorization Response

Redirect back with code:
```
myapp://callback?code=authcode123&state=random-state
```

### 4. Token Exchange

```
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&
code=authcode123&
redirect_uri=myapp://callback&
client_id=abc123&
client_secret=secret456&
code_verifier=xxx              # PKCE
```

Response:
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refresh123",
  "id_token": "eyJhbGciOiJSUzI1NiIs..."
}
```

### 5. UserInfo

```
GET /oauth/userinfo
Authorization: Bearer eyJhbGciOiJSUzI1NiIs...
```

Response:
```json
{
  "sub": "user-uuid",
  "email": "alice@example.com",
  "name": "Alice Smith",
  "picture": "https://..."
}
```

## JWT Signing

Fazt needs an RSA keypair for signing JWTs:

```sql
-- Stored in identity table (encrypted)
INSERT INTO identity (key, value) VALUES
  ('oidc_private_key', encrypted_pem),
  ('oidc_public_key', public_pem);
```

### JWKS Endpoint

```
GET /.well-known/jwks.json
```

Response:
```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "alg": "RS256",
      "kid": "key-1",
      "n": "base64-modulus",
      "e": "AQAB"
    }
  ]
}
```

## ID Token Claims

```json
{
  "iss": "https://abc.com",
  "sub": "user-uuid",
  "aud": "client-id",
  "exp": 1234567890,
  "iat": 1234567800,
  "email": "alice@example.com",
  "name": "Alice Smith",
  "picture": "https://..."
}
```

## Storage

```sql
-- Authorization codes (short-lived)
CREATE TABLE oauth_codes (
    code TEXT PRIMARY KEY,
    client_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    redirect_uri TEXT NOT NULL,
    scope TEXT NOT NULL,
    code_challenge TEXT,
    code_challenge_method TEXT,
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL
);

-- Refresh tokens
CREATE TABLE oauth_refresh_tokens (
    token_hash TEXT PRIMARY KEY,
    client_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    scope TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    expires_at INTEGER
);
```

## Implementation Estimate

| Component | Lines (Go) |
|-----------|------------|
| Discovery endpoint | ~50 |
| Authorization endpoint | ~150 |
| Token endpoint | ~200 |
| UserInfo endpoint | ~50 |
| JWKS endpoint | ~50 |
| Client management | ~100 |
| **Total** | **~600** |

## Why Wait?

1. **Cookie SSO covers subdomain apps** - The primary use case
2. **Adds complexity** - RSA keys, token management, consent UI
3. **No immediate need** - Wait until mobile/third-party integration requested
4. **Can be added later** - Same user table, just new endpoints

## Trigger Conditions

Implement when:
- User requests mobile app authentication
- Third-party integration requires OIDC
- Partner access needed for external domain

Until then, cookie-based SSO is sufficient.
