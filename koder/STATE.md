# Fazt Implementation State

**Last Updated**: 2026-02-05
**Current Version**: v0.25.4

## Status

State: IN PROGRESS
Admin UI Vue port underway - rebuilding with fazt-app patterns and proper state management.

---

## Current Session (2026-02-05) - Admin Vue Port

### In Progress

#### Admin UI Vue Port
Rebuilding admin UI from React to Vue using fazt-app patterns:
- **Why**: Ecosystem consistency (all fazt apps use Vue), better reactivity model
- **Approach**: Port components 1:1 (preserve design), rebuild data layer with fazt-sdk
- **Key improvements**:
  - Proper caching via @tanstack/vue-query
  - fazt-sdk as single source of API communication
  - No more count=0 flicker issues
  - Consistent patterns with other fazt apps

### Plan Location
`koder/plans/38_admin_vue_port.md`

---

## Immediate Attention Required

### fazt.http - External HTTP Calls in Serverless

**Priority**: HIGH - Architectural decision needed before implementation

**What**: Add `fazt.http.*` API for serverless functions to make external HTTP calls.

**Current state**: Serverless can only access `fazt.storage.*` and `fazt.auth.*`. No external network calls.

**Why it's possible**: Goja can call Go functions. Storage APIs already do this (sync Go calls from JS).

**Key decisions needed** (I need to learn more about these):

1. **Security Model**
   - SSRF prevention (block private IPs: 10.x, 192.168.x, 127.x, 169.254.x)
   - Domain allowlists per app?
   - Rate limiting (calls per request, calls per minute)
   - Only HTTPS or allow HTTP?

2. **Async Model**
   - Current runtime is synchronous (Goja ES5)
   - Options:
     a. **Blocking calls** (like storage) - simpler, may block VM pool
     b. **Callback pattern** - awkward in ES5
     c. **Promise polyfill** - complex, but modern feel
   - Need to understand tradeoffs

3. **Timeout Budget**
   - Current: 5 seconds total per request
   - With HTTP calls: need longer? (15s?)
   - Share budget between JS execution and HTTP calls
   - `timeout.Budget` system already exists

**Proposed API** (tentative):
```javascript
var resp = fazt.http.get('https://api.example.com/data')
var resp = fazt.http.post('https://api.example.com/data', {
  body: { key: 'value' },
  headers: { 'Authorization': 'Bearer ...' },
  timeout: 3000  // ms
})
// resp = { status: 200, body: '...', headers: {...} }
```

**Next steps**:
1. Research SSRF prevention best practices
2. Understand async patterns in embedded JS runtimes
3. Draft security spec
4. Implement with proper guardrails

---

## Previous Session (2026-02-05) - v0.25.4 Release

### What Was Done

#### Released v0.25.4 (Bug Fix)
Fixed admin UI health endpoint and released:
- **Problem**: User reported `GET https://admin.zyt.app/api/system/health 401 (Unauthorized)` when accessing admin/logs
- **Root cause**: Health endpoint used `requireAPIKeyAuth` (only CLI/remote), not `requireAdminAuth` (CLI + admin UI sessions)
- **Fix**: Changed `/api/system/health` to use `requireAdminAuth` instead of `requireAPIKeyAuth`
- **Released**: Built all platforms, created GitHub release, pushed v0.25.4
- **Deployed**: Upgraded local binary and all peers (local, zyt) via canonical `fazt upgrade` path

### Commits
```
49fe2a5 release: v0.25.4
7873ef3 fix: allow admin session auth for health endpoint
```

---

## Future Work

1. **Cleanup automation** - Scheduled cleanup of low-weight old entries
2. **Real-time streaming** - WebSocket feed for live activity
3. **Export formats** - Add more export options (Parquet, etc.)

---

## Quick Reference

```bash
# Activity logs
fazt logs list                        # Recent activity
fazt logs list --alias tetris         # Filter by subdomain
fazt logs list --app my-app           # Filter by app
fazt logs list --action pageview      # Filter by action
fazt logs list --min-weight 5         # Important events only
fazt logs stats                       # Statistics
fazt logs cleanup --max-weight 2 --until 7d --force  # Delete old analytics
fazt @zyt logs list --alias fun-game  # Remote peer with filters

# Weight scale
# 9=Security, 8=Auth, 7=Config, 6=Deploy, 5=Data, 4=UserAction, 3=Nav, 2=Analytics, 1=System, 0=Debug

# Version verification
cat version.json | jq -r '.version'                                    # 0.25.4
grep "var Version" internal/config/config.go | grep -oE '[0-9..]+'    # 0.25.4
fazt --version                                                         # 0.25.4

# Remote peers
fazt @zyt logs stats                  # Check activity on production
fazt @local logs list                 # Check local activity
```

---

## Architecture Notes

### Activity Logging System

```
┌─────────────────────────────────────────────────────────────┐
│                    Activity Sources                          │
├──────────┬──────────┬──────────┬──────────┬────────────────┤
│ Auth     │ Deploy   │ Storage  │ Config   │ Pageviews      │
│ (wt 8-9) │ (wt 6)   │ (wt 5)   │ (wt 7)   │ (wt 2)         │
└────┬─────┴────┬─────┴────┬─────┴────┬─────┴───────┬────────┘
     │          │          │          │             │
     └──────────┴──────────┼──────────┴─────────────┘
                           ▼
              ┌────────────────────────┐
              │   activity.Log(Entry)  │
              │   (buffered writes)    │
              └───────────┬────────────┘
                          ▼
              ┌────────────────────────┐
              │   activity_log table   │
              │   (SQLite)             │
              └────────────────────────┘
```

### Analytics SDK Injection

Injected before `</body>` in HTML responses:
```javascript
var h=location.hostname, d=h.split('.').slice(1).join('.');
var u=location.protocol+'//admin.'+d+'/track';
navigator.sendBeacon(u, JSON.stringify({h:h, p:location.pathname, e:'pageview'}))
```

All apps send to `admin.<domain>/track` (centralized tracking endpoint).

### Admin Auth Pattern (v0.25.3+)

System endpoints now use `requireAdminAuth` instead of `requireAPIKeyAuth`:
- **Supports both**: API key auth (CLI/remote) + session auth with admin role (admin UI)
- **Endpoints migrated**: `/api/system/health`, `/api/system/logs/*`
- **Pattern**: Use `requireAdminAuth` for endpoints that admin UI needs to access
