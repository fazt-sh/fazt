# Plan 30b: User Data Foundation

**Status**: Ready to implement
**Created**: 2026-01-30
**Target**: v0.15.0
**Depends on**: Plan 30a (Config Consolidation)
**Unlocks**: Plan 30c (Access Control)

## Overview

Redesign user data handling for automatic isolation, better IDs, and enhanced analytics.

## Goals

1. **Automatic user isolation** - `fazt.app.user.*` guarantees privacy
2. **Better ID format** - Stripe-style IDs with fazt prefix
3. **Enhanced analytics** - Track by app_id and user_id
4. **Admin visibility** - Query users, manage data
5. **GDPR compliance** - One-call user data deletion

## Non-Goals

- RBAC (Plan 30c)
- Domain gating (Plan 30c)
- Breaking existing `fazt.storage.*` immediately (migration path)

---

## API Namespace Redesign

### Current (Confusing)

```javascript
fazt.storage.ds.*   // "storage" doesn't convey app-level
fazt.user.ds.*      // Exists but not nested under app
fazt.private.*      // Works
fazt.auth.*         // Works
```

### New (Clear Hierarchy)

```javascript
// === APP DATA ===

// User's private data (auto user-scoped)
fazt.app.user.ds.insert('settings', { theme: 'dark' })
fazt.app.user.ds.find('drafts', {})
fazt.app.user.kv.set('last_seen', Date.now())
fazt.app.user.kv.get('preferences')
fazt.app.user.s3.put('images/photo.png', data)
fazt.app.user.s3.list('images/')

// App's shared data (developer controls access via API)
fazt.app.ds.insert('posts', { title: 'Hello', authorId: user.id })
fazt.app.ds.find('posts', { published: true })
fazt.app.kv.set('config', { maintenance: false })
fazt.app.s3.put('public/logo.png', data)

// Private files (bundled at deploy)
fazt.app.private.read('config.json')
fazt.app.private.readJSON('seed-data.json')
fazt.app.private.exists('feature-flags.json')
fazt.app.private.list()


// === FRAMEWORK SERVICES ===

// Authentication (existing + minor additions)
fazt.auth.getUser()
fazt.auth.isLoggedIn()
fazt.auth.requireLogin()
fazt.auth.requireAdmin()
fazt.auth.requireOwner()

// Admin (owner/admin only)
fazt.admin.users.list()
fazt.admin.users.get('fazt_usr_Nf4rFeUfNV2H')
fazt.admin.users.delete('fazt_usr_Nf4rFeUfNV2H')
fazt.admin.analytics.query({ appId: '...', userId: '...' })

// Analytics (tracking)
fazt.analytics.track('purchase', { amount: 99 })
```

### Hierarchy Visualization

```
fazt
├── app                      # This app's data
│   ├── user                 # Current user's private data
│   │   ├── ds.*             # Documents
│   │   ├── kv.*             # Key-value
│   │   └── s3.*             # Files/blobs
│   ├── ds.*                 # Shared documents
│   ├── kv.*                 # Shared key-value
│   ├── s3.*                 # Shared files
│   └── private.*            # Bundled private files
│
├── auth                     # Authentication service
│   ├── getUser()
│   ├── isLoggedIn()
│   ├── requireLogin()
│   ├── requireAdmin()
│   └── requireOwner()
│
├── admin                    # Admin operations (owner/admin only)
│   ├── users.*
│   └── analytics.*
│
└── analytics                # Event tracking
    └── track()
```

---

## User Isolation Model

### How It Works

| Namespace | Visibility | Enforcement |
|-----------|------------|-------------|
| `fazt.app.user.*` | Only that user | Database-level `user_id` filter |
| `fazt.app.*` | API decides | Developer implements access control |

### Storage Table Changes

```sql
-- Current
CREATE TABLE app_storage (
  app_id TEXT,
  collection TEXT,
  data JSON
);

-- Enhanced
CREATE TABLE app_storage (
  app_id TEXT,
  collection TEXT,
  user_id TEXT,        -- NULL for shared data, set for user data
  data JSON,
  created_at INTEGER,
  updated_at INTEGER
);

CREATE INDEX idx_storage_app_user ON app_storage(app_id, user_id);
CREATE INDEX idx_storage_app_collection ON app_storage(app_id, collection);
```

### Isolation Guarantee

```javascript
// User A is logged in
fazt.app.user.ds.find('notes', {})
// SQL: SELECT * FROM app_storage
//      WHERE app_id = ? AND collection = 'notes' AND user_id = 'fazt_usr_A...'

// Even if attacker tries:
fazt.app.user.ds.find('notes', { user_id: 'fazt_usr_B...' })
// The user_id in query is IGNORED - always uses current user
// Impossible to access another user's data
```

---

## ID Format Redesign

### Current (Inconsistent)

| Type | Format | Example | Issue |
|------|--------|---------|-------|
| App | `app_` + 8 chars | `app_7f3k9x2m` | No global prefix |
| User | UUID v4 | `550e8400-e29b-...` | Too long (36 chars) |

### New (Stripe-Style)

| Type | Format | Example | Length |
|------|--------|---------|--------|
| User | `fazt_usr_` + 12 | `fazt_usr_Nf4rFeUfNV2H` | 20 |
| App | `fazt_app_` + 12 | `fazt_app_qW8n4P1zXy3m` | 20 |
| Token | `fazt_tok_` + 12 | `fazt_tok_Ab3dEf6gHi9j` | 20 |
| Session | `fazt_ses_` + 12 | `fazt_ses_Kl0mNoPqRs2t` | 20 |
| Invite | `fazt_inv_` + 12 | `fazt_inv_Mn7oPqRs3tUv` | 20 |

### ID Generation

```go
const base62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func generateID(prefix string) string {
    b := make([]byte, 12)
    for i := range b {
        b[i] = base62[rand.Intn(62)]
    }
    return "fazt_" + prefix + "_" + string(b)
}

// Usage
generateID("usr")  // fazt_usr_Nf4rFeUfNV2H
generateID("app")  // fazt_app_qW8n4P1zXy3m
```

### Benefits

| Benefit | How |
|---------|-----|
| Globally identifiable | `fazt_` prefix recognizable anywhere |
| Type at a glance | `usr_`, `app_`, `tok_` identifies type |
| Greppable | `grep "fazt_usr_"` finds all users |
| Short | 20 chars vs 36 for UUID |
| URL-safe | No hyphens, no special chars |
| Collision-proof | 12 chars base62 = 3.2 x 10^21 combos |
| Copy-pasteable | Double-click selects whole ID |

---

## Analytics Enhancement

### Events Table Changes

```sql
-- Add columns
ALTER TABLE events ADD COLUMN app_id TEXT;
ALTER TABLE events ADD COLUMN user_id TEXT;

-- Add indexes
CREATE INDEX idx_events_app ON events(app_id, created_at);
CREATE INDEX idx_events_user ON events(user_id, created_at);
CREATE INDEX idx_events_app_user ON events(app_id, user_id, created_at);
```

### Automatic User Tracking

When a logged-in user triggers an event, `user_id` is automatically captured:

```javascript
// User is logged in, views a page
// Event automatically includes:
{
  app_id: 'fazt_app_qW8n4P1zXy3m',
  user_id: 'fazt_usr_Nf4rFeUfNV2H',  // Auto-captured
  event_type: 'pageview',
  path: '/dashboard',
  created_at: 1706620800
}

// Explicit tracking also includes user_id
fazt.analytics.track('purchase', { amount: 99 })
// Stored with user_id automatically
```

### Query API

```javascript
// All events for an app
fazt.admin.analytics.query({
  appId: 'fazt_app_qW8n4P1zXy3m'
})

// All events for a user (across ALL apps)
fazt.admin.analytics.query({
  userId: 'fazt_usr_Nf4rFeUfNV2H'
})

// User's activity in specific app
fazt.admin.analytics.query({
  appId: 'fazt_app_qW8n4P1zXy3m',
  userId: 'fazt_usr_Nf4rFeUfNV2H'
})

// Filtered by event type and time
fazt.admin.analytics.query({
  appId: 'fazt_app_qW8n4P1zXy3m',
  event: 'purchase',
  from: '2026-01-01',
  to: '2026-01-31'
})
```

---

## Admin APIs

### User Management

```javascript
// List all users who've used this app
fazt.admin.users.list()
// Returns: [{ id, email, name, role, lastSeen, createdAt }]

// Get specific user
fazt.admin.users.get('fazt_usr_Nf4rFeUfNV2H')
// Returns: { id, email, name, role, picture, provider, lastSeen, createdAt }

// Set base role (owner can promote to admin, admin cannot promote)
fazt.admin.users.setRole('fazt_usr_Nf4rFeUfNV2H', 'admin')

// GDPR: Delete user and ALL their data
fazt.admin.users.delete('fazt_usr_Nf4rFeUfNV2H')
```

### Base Role Model

| Role | `fazt.app.*` | `fazt.app.user.*` | `fazt.admin.*` |
|------|--------------|-------------------|----------------|
| `user` | Read shared | Full (own data) | No access |
| `admin` | Full | Full (own data) | Read-only |
| `owner` | Full | Full (own data) | Full |

- **One owner** per fazt instance (the person who deployed it)
- **Multiple admins** allowed
- **Unlimited users**

---

## GDPR Compliance

### One-Call Delete

```javascript
fazt.admin.users.delete('fazt_usr_Nf4rFeUfNV2H')
```

### What Gets Deleted

```sql
-- 1. User's storage data (all apps)
DELETE FROM app_storage WHERE user_id = ?;

-- 2. User's blobs (all apps)
DELETE FROM app_blobs WHERE user_id = ?;

-- 3. User's events
DELETE FROM events WHERE user_id = ?;

-- 4. User's sessions
DELETE FROM auth_sessions WHERE user_id = ?;

-- 5. User record
DELETE FROM auth_users WHERE id = ?;
```

### Optional: Anonymize Instead

```javascript
// Keep analytics, remove PII
fazt.admin.users.anonymize('fazt_usr_Nf4rFeUfNV2H')
// Sets email='[deleted]', name='[deleted]', picture=null
// Keeps events with user_id for aggregate analytics
```

---

## Migration Path

### Phase 1: Add New APIs (Non-Breaking)

1. Add `fazt.app.user.*` namespace
2. Add `fazt.app.*` as alias for `fazt.storage.*`
3. Add `fazt.admin.*` namespace
4. Add `user_id` column to storage tables
5. Update ID generation to new format

### Phase 2: Deprecation Warnings

1. Log warning when `fazt.storage.*` is used
2. Document migration guide

### Phase 3: Remove Old APIs (Breaking)

1. Remove `fazt.storage.*`
2. Remove old ID format generation

---

## Implementation Tasks

### Database

- [ ] Add `user_id` column to `app_storage` table
- [ ] Add `user_id` column to `app_blobs` table
- [ ] Add `app_id`, `user_id` columns to `events` table
- [ ] Add indexes for efficient queries
- [ ] Migration script for existing data

### ID Generation

- [ ] Create `generateFaztID(prefix string)` function
- [ ] Update `generateAppID()` to use new format
- [ ] Update user ID generation to use new format
- [ ] Update token generation
- [ ] Update session ID generation

### Runtime Bindings

- [ ] Implement `fazt.app.user.ds.*` bindings
- [ ] Implement `fazt.app.user.kv.*` bindings
- [ ] Implement `fazt.app.user.s3.*` bindings
- [ ] Implement `fazt.app.ds.*` (alias for storage)
- [ ] Implement `fazt.app.kv.*`
- [ ] Implement `fazt.app.s3.*`
- [ ] Move `fazt.private.*` to `fazt.app.private.*`
- [ ] Implement `fazt.admin.users.list()`
- [ ] Implement `fazt.admin.users.get()`
- [ ] Implement `fazt.admin.users.setRole()`
- [ ] Implement `fazt.admin.users.delete()`
- [ ] Implement `fazt.admin.users.anonymize()`
- [ ] Implement `fazt.admin.analytics.query()`
- [ ] Implement `fazt.analytics.track()`

### Analytics

- [ ] Capture `user_id` on page views (if logged in)
- [ ] Capture `app_id` on all events
- [ ] Implement analytics query API

### Testing

- [ ] User isolation tests (can't access other user's data)
- [ ] Admin API tests
- [ ] Analytics query tests
- [ ] GDPR delete tests
- [ ] ID format tests

### Documentation

- [ ] Update `serverless-api.md`
- [ ] Update `auth-integration.md`
- [ ] Add admin API docs
- [ ] Migration guide from old namespaces

---

## Example: Online Photo Editor

```javascript
// api/main.js

var user = fazt.auth.getUser()
var ds = fazt.app.user.ds
var s3 = fazt.app.user.s3

// Upload image (user's private storage)
if (resource === 'upload' && method === 'POST') {
  fazt.auth.requireLogin()

  var filename = request.body.filename
  var data = request.body.data

  s3.put('images/' + filename, data, 'image/png')
  fazt.analytics.track('upload', { filename: filename })

  respond({ ok: true })
}

// List user's images
if (resource === 'images' && method === 'GET') {
  fazt.auth.requireLogin()

  var images = s3.list('images/')
  respond({ images: images })
}

// User settings
if (resource === 'settings') {
  fazt.auth.requireLogin()

  if (method === 'GET') {
    respond(ds.findOne('settings', {}) || { theme: 'light' })
  }

  if (method === 'PUT') {
    ds.update('settings', {}, request.body)
    respond({ ok: true })
  }
}

// Public templates (admin creates, all can read)
if (resource === 'templates' && method === 'GET') {
  var templates = fazt.app.ds.find('templates', { published: true })
  respond({ templates: templates })
}

// Admin: create template
if (resource === 'templates' && method === 'POST') {
  fazt.auth.requireAdmin()
  fazt.app.ds.insert('templates', request.body)
  respond(201, { ok: true })
}
```

---

## Success Criteria

1. `fazt.app.user.*` automatically isolates user data
2. Analytics tracks `app_id` and `user_id` on all events
3. `fazt.admin.users.delete()` removes all user data (GDPR)
4. All new IDs use `fazt_<type>_<12 chars>` format
5. No breaking changes to existing apps (deprecation path)
6. Full test coverage for isolation guarantees
