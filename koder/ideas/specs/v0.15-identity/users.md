# Users

## Summary

Multi-user support for fazt. The owner (Persona) manages users who can
authenticate via social login or password. Users are stored locally in SQLite
and share sessions across all subdomain apps.

## Mental Model

```
Owner (Persona)          Users
─────────────────        ─────────────────
- Hardware-bound         - Stored in SQLite
- Keypair identity       - Password or social login
- Full admin             - Role-based access
- One per instance       - Many per instance
```

The owner is NOT a user row. Persona is kernel-level. Users are app-level.

## Storage

```sql
CREATE TABLE auth_users (
    id TEXT PRIMARY KEY,           -- UUID
    email TEXT UNIQUE NOT NULL,
    name TEXT,
    picture TEXT,                  -- Avatar URL from provider
    provider TEXT NOT NULL,        -- 'google', 'github', 'discord', 'password'
    provider_id TEXT,              -- External ID (null for password)
    password_hash TEXT,            -- Only if provider='password'
    role TEXT DEFAULT 'user',      -- 'admin', 'user'
    invited_by TEXT,               -- User ID who invited, or 'owner'
    created_at INTEGER NOT NULL,
    last_login INTEGER
);

CREATE INDEX idx_auth_users_email ON auth_users(email);
CREATE INDEX idx_auth_users_provider ON auth_users(provider, provider_id);
```

## Roles

| Role | Capabilities |
|------|--------------|
| `owner` | Everything (Persona, not in users table) |
| `admin` | Manage users, access admin features |
| `user` | Use apps, no admin access |

Apps can define custom roles via their own tables. The auth layer only
manages these three base roles.

## User Lifecycle

### Creation via Social Login

```
1. User clicks "Login with Google"
2. OAuth flow completes
3. Fazt receives: { email, name, picture, provider_id }
4. Check: Does user with this email exist?
   - Yes: Update last_login, return session
   - No: Create user row, return session
```

### Creation via Password (Invite Flow)

```
1. Owner/admin generates invite code
2. New user visits invite link
3. User sets email + password
4. User row created with provider='password'
```

No email verification needed—invite code proves authorization.

### Invite Codes

```sql
CREATE TABLE auth_invites (
    code TEXT PRIMARY KEY,         -- Random string
    role TEXT DEFAULT 'user',      -- Role to assign
    created_by TEXT NOT NULL,      -- User ID or 'owner'
    created_at INTEGER NOT NULL,
    expires_at INTEGER,            -- Optional expiry
    used_by TEXT,                  -- User ID who used it
    used_at INTEGER
);
```

## JS Runtime API

```javascript
// Get current user
const user = await fazt.auth.getUser();
// Returns null if not logged in
// Returns: { id, email, name, picture, role, provider }

// Check login status
if (await fazt.auth.isLoggedIn()) { ... }

// Require login (redirects to login page if not authenticated)
await fazt.auth.requireLogin();
// Throws redirect, code after this only runs if logged in

// Check roles
if (await fazt.auth.hasRole('admin')) { ... }

// Check if owner (Persona, not a user)
if (await fazt.auth.isOwner()) { ... }

// Get all users (admin only)
const users = await fazt.auth.listUsers();

// Logout current user
await fazt.auth.logout();
```

## Admin Operations

Owner and admins can manage users:

```javascript
// Create invite code
const invite = await fazt.auth.createInvite({ role: 'user' });
// { code: 'abc123', url: 'https://abc.com/auth/invite/abc123' }

// List users
const users = await fazt.auth.listUsers();

// Update user role
await fazt.auth.updateUser(userId, { role: 'admin' });

// Remove user
await fazt.auth.removeUser(userId);
```

## HTTP Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/auth/invite` | POST | Create invite code (admin) |
| `/auth/invite/:code` | GET | Render invite signup form |
| `/auth/invite/:code` | POST | Complete invite signup |
| `/auth/users` | GET | List users (admin) |
| `/auth/users/:id` | PATCH | Update user (admin) |
| `/auth/users/:id` | DELETE | Remove user (admin) |

## Privacy

- User data never leaves the fazt instance
- No analytics or tracking
- Owner has full access to all user data
- Users can request their data (GDPR-style)

## Edge Cases

### Email Collision

User signs up with Google (alice@example.com), later tries GitHub with same
email:

- **Policy**: Link accounts—same user, multiple providers
- Update `provider` to most recent, keep `provider_id` from first

### Owner Email Match

If a user signs up with owner's email:

- **Policy**: Reject—owner email is reserved
- Owner uses Persona, not user table

### Invite Expiry

Expired invite codes return 410 Gone. Admins can create new codes.
