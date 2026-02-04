# Plan 35: SDK Real Data Integration

**Created**: 2026-02-03
**Status**: In Progress
**Goal**: Web-based interface to manage a Fazt node with real data

## Background

The admin UI was built with example/mock data to demonstrate UI patterns. Now we need to replace those examples with real data from the Fazt backend, ensuring all interfaces evolve together:

```
CLI (primary) → API (powers remotes) → SDK (JavaScript client) → Admin (visual interface)
```

**Principles**:
- CLI is the source of truth
- SDK first, test, then admin
- Conservative: 90% of capabilities needed to manage a node
- Useful now > feature-complete later
- Logs & analytics extend after core is solid

## Key Clarifications

1. **SDK First**: Build and verify in fazt-sdk before touching admin UI
2. **Admin Alias**: Will run on reserved `admin` alias (admin.zyt.app)
3. **Can Replace**: Old admin implementation can be deleted and rebuilt
4. **Auth**: Uses standard session auth (no special native auth)
5. **Responsive**: Significant effort in mobile/desktop foundation - preserve it
6. **Defer**: Logs & analytics UI can extend after core system works

## Current State

### API Endpoints (Backend)

| Category | Endpoint | Description |
|----------|----------|-------------|
| **Apps** | `GET /api/apps` | List all apps |
| | `GET /api/apps/{id}` | App details |
| | `GET /api/apps/{id}/files` | File listing |
| | `GET /api/apps/{id}/source` | Source tracking |
| | `POST /api/apps/install` | Install from Git |
| | `DELETE /api/apps/{id}` | Delete app |
| **Aliases** | `GET /api/aliases` | List aliases |
| | `GET /api/aliases/{subdomain}` | Alias details |
| | `POST /api/aliases` | Create alias |
| | `DELETE /api/aliases/{subdomain}` | Delete alias |
| **System** | `GET /api/system/health` | Server status |
| | `GET /api/system/config` | Configuration |
| | `GET /api/system/db` | Database stats |
| | `GET /api/system/capacity` | Resource usage |
| **Events** | `GET /api/events` | Analytics events |
| | `GET /api/stats` | Event statistics |
| **Logs** | `GET /api/logs` | Runtime logs |
| **Templates** | `GET /api/templates` | Available templates |

### SDK Coverage (fazt-sdk)

```javascript
✓ apps.list(), apps.get(), apps.files(), apps.delete(), apps.install()
✓ aliases.list(), aliases.get(), aliases.create(), aliases.update(), aliases.delete()
✓ system.health(), system.config(), system.db(), system.capacity()
✓ events.list()
✓ logs.list()
✓ templates.list()
✓ auth.session(), auth.signOut()
```

### Data Stores (admin/src/stores/data.js)

```javascript
✓ apps - loadApps(client)
✓ aliases - loadAliases(client)
✓ health - loadHealth(client)
✗ events - loadEvents() exists but not used in dashboard
✗ system db/capacity - not loaded
```

## Implementation Plan

### Part A: SDK Verification (First)

Test and verify SDK methods work correctly with real backend before touching admin UI.

#### A1. Test Existing SDK Methods

```javascript
// Test in browser console at http://admin.192.168.64.3.nip.io:8080
const { client } = window.__fazt_admin

// Apps - core management
await client.apps.list()         // Should return 16 apps
await client.apps.get('admin-ui')

// Aliases - routing config
await client.aliases.list()      // Should return subdomain mappings

// System - node health
await client.system.health()     // status, version, uptime
await client.system.config()     // server configuration
await client.system.db()         // database stats
await client.system.capacity()   // resource usage

// Auth - session state
await client.auth.session()      // authenticated, user

// Events - activity
await client.events.list({ limit: 10 })
```

#### A2. Verify Response Structures

Check actual API responses and ensure SDK/mock match:
- Apps: id, name, size_bytes, updated_at, file_count
- Aliases: subdomain, type, targets
- Health: status, version, uptime_seconds
- Events: event_type, domain, created_at

#### A3. Update Mock Fixtures

Ensure `admin/packages/fazt-sdk/fixtures/` match real API responses exactly.

---

### Part B: Admin Dashboard (After SDK Verified)

Rebuild dashboard with real data, preserving responsive design.

#### B1. System Status Section

**Source**: `system.health()` combined with `system.db()`, `system.capacity()`

**Shows**: Health status, version, uptime, memory, DB size

**Work**:
1. Add `systemInfo` store in `data.js`
2. Add `loadSystemInfo(client)` function
3. Update dashboard status panel

#### B2. Apps Section

**Source**: `apps.list()` (already loaded)

**Shows**: App list with name, size, last updated

**Work**:
1. Verify `loadApps()` populates store
2. Update dashboard apps panel to use real data
3. Remove mock/example data

#### B3. Aliases Section

**Source**: `aliases.list()` (already loaded)

**Shows**: Subdomain → target routing

**Work**:
1. Add aliases panel to dashboard
2. Show real routing configuration

#### B4. Cleanup

- Remove fake "Notifications" section
- Remove or wire "Quick Actions"
- Remove hardcoded stats/examples

---

### Part C: Deferred (After Core Works)

1. **Logs UI** - `logs.list()` API exists, viewer later
2. **Analytics UI** - `stats.analytics()` exists, viewer later
3. **Events page** - Already exists, refine later

---

## Execution Steps

```
1. [ ] A1: Test SDK methods in browser console
2. [ ] A2: Document actual API response structures
3. [ ] A3: Update mock fixtures if needed
4. [ ] B1: Add systemInfo store and loader
5. [ ] B2: Verify apps panel works with real data
6. [ ] B3: Add aliases section to dashboard
7. [ ] B4: Remove example/mock data from dashboard
8. [ ] Test: Verify in real mode (no ?mock=true)
9. [ ] Test: Verify mock mode still works
10. [ ] Deploy to admin-ui alias
```

## Success Criteria

- [ ] SDK methods return expected data structures
- [ ] Dashboard shows real apps (16 from local)
- [ ] Dashboard shows real system health/version
- [ ] Dashboard shows real aliases
- [ ] No fake metrics or hardcoded examples
- [ ] Works in both real and mock modes
- [ ] Responsive design preserved

## Out of Scope (Future)

- SQL query interface (needs design thought)
- Peers management (needs `/api/peers` endpoint)
- Logs/Analytics UI (extend after core works)

## Related

- `knowledge-base/workflows/admin-ui/architecture.md` - Data flow
- `knowledge-base/workflows/admin-ui/checklist.md` - Validation checklist
- `admin/packages/fazt-sdk/` - SDK implementation
