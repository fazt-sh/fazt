# Fazt Implementation State

**Last Updated**: 2026-02-05
**Current Version**: v0.25.4

## Status

State: RELEASED
v0.25.4 shipped - Fixed admin UI health endpoint authentication (401 error resolved).

---

## Current Session (2026-02-05) - v0.25.4 Release

### What Was Done

#### Released v0.25.4 (Bug Fix)
Fixed admin UI health endpoint and released:
- **Problem**: User reported `GET https://admin.zyt.app/api/system/health 401 (Unauthorized)` when accessing admin/logs
- **Root cause**: Health endpoint used `requireAPIKeyAuth` (only CLI/remote), not `requireAdminAuth` (CLI + admin UI sessions)
- **Fix**: Changed `/api/system/health` to use `requireAdminAuth` instead of `requireAPIKeyAuth`
  - Allows both API key auth (CLI/remote) and session auth with admin role (admin UI)
  - Matches the same pattern applied to logs endpoints in v0.25.3
- **Released**: Built all platforms, created GitHub release, pushed v0.25.4
- **Deployed**: Upgraded local binary and all peers (local, zyt) via canonical `fazt upgrade` path
- **Verified**: Both peers at v0.25.4, healthy status ✓

### Commits
```
49fe2a5 release: v0.25.4
7873ef3 fix: allow admin session auth for health endpoint
```

---

## Last Session (2026-02-04) - v0.25.2 Release

### What Was Done

#### Released v0.25.2 (Bug Fix)
Fixed root domain pageview tracking and released:
- **Root domain tracking**: Fixed analytics script domain extraction logic
  - Before: `zyt.app` → `admin.app/track` ❌ (invalid URL)
  - After: `zyt.app` → `admin.zyt.app/track` ✓
  - Logic: Check if sliced result contains `.` (domain.tld), else use full hostname
- **Released**: Built all platforms, created GitHub release, pushed v0.25.2
- **Deployed**: Upgraded local binary and all peers (local, zyt) via canonical `fazt upgrade` path
- **Verified**: Analytics script deployed correctly on https://zyt.app/ ✓

### Commits
```
35dfea6 release: v0.25.2
ff5f754 fix: root domain pageview tracking
```

---

## Last Session (2026-02-04) - v0.25.1 Release

### What Was Done

#### Released v0.25.1 (Bug Fix)
Fixed remote SQL command and released:
- **Panic on nil fields**: Added safe type assertions for `count` and `time_ms` with ok-pattern checks
- **Authentication failure**: Moved `/api/sql` to API key auth bypass list
- **Error handling**: Added HTTP status code checking before JSON decode
- **Released & Deployed**: All peers upgraded to v0.25.1

### Commits
```
815a405 release: v0.25.1
6af532f docs: update STATE.md - remote SQL fix complete
3e91f95 fix: remote SQL command authentication and panic
```

---

## Next Up

### High Priority
1. **Admin UI testing** - Verify logs page works correctly with health endpoint fix
2. **Admin UI integration** - Continue building out admin pages (apps, users, etc.)

### Future Work
1. **Root domain tracking** - Verify pageviews working for `zyt.app` (fixed in v0.25.2)
2. **Cleanup automation** - Scheduled cleanup of low-weight old entries
3. **Real-time streaming** - WebSocket feed for live activity
4. **Export formats** - Add more export options (Parquet, etc.)

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
