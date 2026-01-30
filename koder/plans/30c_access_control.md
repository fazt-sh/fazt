# Plan 30c: Access Control

**Status**: Ready to implement
**Created**: 2026-01-30
**Target**: v0.16.0
**Depends on**: Plan 30a (Config), Plan 30b (User Data)

## Overview

Add lightweight RBAC and domain gating for fine-grained access control.

## Goals

1. **Lightweight RBAC** - Hierarchical roles with expiry
2. **Domain gating** - Restrict app access by email domain

## Non-Goals

- Permission-based RBAC (`can_edit_posts`) - apps can implement this
- Role inheritance/groups - hierarchy handles most cases

---

## RBAC (Lightweight)

### Philosophy

- Invisible to those who don't use it
- Powerful for those who do
- No config until you need it

### Role Storage

Roles stored as array on user object:

```javascript
{
  id: 'fazt_usr_Nf4rFeUfNV2H',
  email: 'user@example.com',
  roles: [
    { id: 'teacher/chemistry/lab', createdAt, updatedAt, createdBy, expiry: 1735689600 },
    { id: 'guardian', createdAt, updatedAt, createdBy, disabled: true },
    { id: 'clubAdmin/secretary', createdAt, updatedAt, createdBy }
  ]
}
```

### What This Enables

| Feature | How |
|---------|-----|
| Hierarchical roles | `teacher/chemistry/*` matches any chemistry teacher |
| Expiry | Automatic expiration for temporary access |
| Multiple roles | User can have many roles |
| Disable without delete | Suspend access, restore later |
| Audit trail | Who assigned, when |

### API

**App-level (checking roles):**

```javascript
// Check if user has role (supports glob patterns)
if (fazt.auth.hasRole('teacher/*')) {
  // User is any kind of teacher
}

if (fazt.auth.hasRole('teacher/chemistry/*')) {
  // User is a chemistry teacher
}

// Gate endpoint (throws 403 if not)
fazt.auth.requireRole('editor')
```

**Admin-level (managing roles):**

```javascript
// Add role with optional expiry
fazt.admin.users.role.add('fazt_usr_...', 'teacher/chemistry/lab', {
  expiry: Date.now() + 7 * 24 * 60 * 60 * 1000  // 7 days
})

// Remove role
fazt.admin.users.role.remove('fazt_usr_...', 'teacher/chemistry/lab')

// List roles for a user
fazt.admin.users.role.list('fazt_usr_...')
// -> ['teacher/chemistry/lab', 'guardian']

// Find users with matching role
fazt.admin.users.role.find('teacher/*')
// -> ['fazt_usr_abc...', 'fazt_usr_xyz...']
```

### Role Matching

Pattern matching uses glob-style:

| Pattern | Matches |
|---------|---------|
| `teacher` | Exact match only |
| `teacher/*` | `teacher/chemistry`, `teacher/physics` |
| `teacher/chemistry/*` | `teacher/chemistry/lab`, `teacher/chemistry/theory` |
| `*/admin` | `club/admin`, `dept/admin` |

### Expiry Enforcement

`hasRole()` and `requireRole()` automatically check expiry:

```javascript
// Role with expiry in the past -> treated as not having role
fazt.auth.hasRole('guest')  // -> false if expired
```

### Database Changes

```sql
-- Add roles column to users table
ALTER TABLE auth_users ADD COLUMN roles JSON DEFAULT '[]';

-- Index for role queries (SQLite JSON functions)
CREATE INDEX idx_users_roles ON auth_users(roles);
```

### Base Roles vs Custom Roles

Base roles (from Plan 30b) and custom roles serve different purposes:

| Type | Purpose | Examples |
|------|---------|----------|
| Base roles | Fazt admin access | `owner`, `admin`, `user` |
| Custom roles | App-level permissions | `teacher/chemistry`, `editor`, `dept/exec` |

```javascript
// Base role check (built-in)
fazt.auth.requireAdmin()

// Custom role check (RBAC)
fazt.auth.requireRole('teacher/*')
```

Custom roles are additive to base roles.

---

## Domain Gating

Restrict app access by email domain at OAuth level.

### Use Cases

| Scenario | Config |
|----------|--------|
| Employees only | `allowedDomains: ['storybrain.com']` |
| Block disposable emails | `blockedDomains: ['tempmail.com', 'fakeinbox.net']` |
| Partner access | `allowedDomains: ['company.com', 'partner.com']` |

### API

```javascript
// Allow only specific domains (whitelist)
fazt.admin.config.auth.domains.allow(['storybrain.com', 'partner.com'])

// Block specific domains (blacklist)
fazt.admin.config.auth.domains.block(['tempmail.com', 'fakeinbox.net'])

// Get current config
fazt.admin.config.auth.domains.list()
// -> { allowed: ['storybrain.com'], blocked: ['tempmail.com'] }

// Granular updates
fazt.admin.config.auth.domains.allowed.add('newpartner.com')
fazt.admin.config.auth.domains.allowed.remove('oldpartner.com')
fazt.admin.config.auth.domains.blocked.add('spam.net')
```

### Enforcement Logic

Checked at OAuth callback, before session creation:

```
1. User authenticates with Google/GitHub
2. Extract email domain
3. If allowedDomains is set:
   - Domain must be in list -> continue
   - Otherwise -> reject with "Access restricted to @domain.com"
4. If blockedDomains is set:
   - Domain in list -> reject with "Email domain not allowed"
5. Create session
```

If both allowed and blocked are set, allowed is checked first.

### Storage

Uses config table from Plan 30a:

```sql
-- App-level domain config
INSERT INTO config (key, value) VALUES
  ('app.storybrain.auth.domains.allowed', '["storybrain.com"]'),
  ('app.storybrain.auth.domains.blocked', '["tempmail.com"]');
```

### Future Extensions

The `fazt.admin.config.*` namespace can accommodate other app settings:

```javascript
fazt.admin.config.auth.domains.*           // Domain gating
fazt.admin.config.auth.requireLogin        // Force login for all routes
fazt.admin.config.limits.storage           // Per-user storage limits
fazt.admin.config.features.*               // Feature flags
```

---

## Updated API Hierarchy

Adding to Plan 30b's hierarchy:

```
fazt
├── auth
│   ├── ... (existing from 30b)
│   ├── hasRole(pattern)      # NEW: Check role
│   └── requireRole(pattern)  # NEW: Gate by role
│
└── admin
    ├── users
    │   ├── ... (existing from 30b)
    │   └── role                # NEW: Role management
    │       ├── add(userId, role, opts)
    │       ├── remove(userId, role)
    │       ├── list(userId)
    │       └── find(pattern)
    │
    └── config                  # NEW: App config
        └── auth
            └── domains
                ├── allow([...])
                ├── block([...])
                ├── list()
                ├── allowed.add(domain)
                ├── allowed.remove(domain)
                ├── blocked.add(domain)
                └── blocked.remove(domain)
```

---

## Implementation Tasks

### RBAC

- [ ] Add `roles` JSON column to `auth_users` table
- [ ] Implement `fazt.auth.hasRole(pattern)` with glob matching
- [ ] Implement `fazt.auth.requireRole(pattern)` (throws 403)
- [ ] Implement `fazt.admin.users.role.add(userId, role, opts)`
- [ ] Implement `fazt.admin.users.role.remove(userId, role)`
- [ ] Implement `fazt.admin.users.role.list(userId)`
- [ ] Implement `fazt.admin.users.role.find(pattern)`
- [ ] Automatic expiry checking in hasRole/requireRole
- [ ] Include `createdBy` (current admin) when adding roles

### Domain Gating

- [ ] Implement `fazt.admin.config.auth.domains.allow([...])`
- [ ] Implement `fazt.admin.config.auth.domains.block([...])`
- [ ] Implement `fazt.admin.config.auth.domains.list()`
- [ ] Implement granular `.allowed.add/remove`
- [ ] Implement granular `.blocked.add/remove`
- [ ] Enforce domain restrictions at OAuth callback
- [ ] Error messages for rejected domains ("Access restricted to @domain.com")

### Testing

- [ ] RBAC: hasRole with exact match
- [ ] RBAC: hasRole with glob patterns
- [ ] RBAC: requireRole throws 403 when missing
- [ ] RBAC: expiry checking
- [ ] RBAC: disabled roles ignored
- [ ] RBAC: role.find queries
- [ ] Domain gating: allowed domains only
- [ ] Domain gating: blocked domains rejected
- [ ] Domain gating: both allowed and blocked

### Documentation

- [ ] Add RBAC docs with examples
- [ ] Add domain gating docs
- [ ] Update auth-integration.md

---

## Example: B2B Internal Tool (Domain-Gated)

A startup's internal dashboard, accessible only to employees.

**Setup (one-time, via admin API or future admin UI):**

```javascript
// Only allow company employees
fazt.admin.config.auth.domains.allow(['storybrain.com'])

// Set up department roles
fazt.admin.users.role.add('fazt_usr_ceo...', 'dept/exec')
fazt.admin.users.role.add('fazt_usr_eng1...', 'dept/engineering')
fazt.admin.users.role.add('fazt_usr_eng2...', 'dept/engineering', {
  expiry: Date.now() + 90 * 24 * 60 * 60 * 1000  // Contractor, 90 days
})
```

**App code:**

```javascript
// api/main.js

// All routes require login (domain-gated at OAuth level)
fazt.auth.requireLogin()

// Dashboard - all employees
if (resource === 'dashboard') {
  var metrics = fazt.app.ds.findOne('metrics', { current: true })
  respond({ metrics: metrics })
}

// Engineering-only: deployment controls
if (resource === 'deploy' && method === 'POST') {
  fazt.auth.requireRole('dept/engineering')
  // trigger deployment...
  fazt.analytics.track('deploy', { env: request.body.env })
  respond({ ok: true })
}

// Exec-only: sensitive financials
if (resource === 'financials') {
  fazt.auth.requireRole('dept/exec')
  var data = fazt.app.private.readJSON('financials.json')
  respond(data)
}

// Admin: manage employee roles
if (resource === 'admin/roles' && method === 'POST') {
  fazt.auth.requireAdmin()
  var { userId, role, expiry } = request.body
  fazt.admin.users.role.add(userId, role, expiry ? { expiry } : {})
  respond({ ok: true })
}
```

**What happens:**
- Non-@storybrain.com emails -> rejected at OAuth ("Access restricted to @storybrain.com")
- Employees can access dashboard
- Only engineering dept can deploy
- Only execs see financials
- Contractor's access auto-expires after 90 days

---

## Example: Photo Editor with Roles

Extending the photo editor from Plan 30b:

```javascript
// Role-gated: only editors can feature images
if (resource === 'feature' && method === 'POST') {
  fazt.auth.requireRole('editor')
  var imageId = request.body.imageId
  fazt.app.ds.update('featured', { id: imageId }, { featured: true })
  respond({ ok: true })
}

// Moderators can delete any user's public images
if (resource === 'moderate/delete' && method === 'POST') {
  fazt.auth.requireRole('moderator')
  var imageId = request.body.imageId
  fazt.app.ds.delete('public_images', { id: imageId })
  fazt.analytics.track('moderation', { action: 'delete', imageId })
  respond({ ok: true })
}
```

---

## Success Criteria

1. `fazt.auth.hasRole()` supports hierarchical patterns
2. `fazt.auth.requireRole()` throws 403 when role missing
3. Expired roles automatically rejected
4. `fazt.admin.users.role.*` API works for role management
5. `fazt.admin.config.auth.domains.*` gates access at OAuth level
6. Domain gating configurable via API (no code changes needed)
7. Both allowed and blocked domain lists work correctly
