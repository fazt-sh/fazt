# Fazt Implementation State

**Last Updated**: 2026-02-06
**Current Version**: v0.25.4

## Status

State: CLEAN
Admin Vue port complete. Dashboard rendering errors resolved. All components granular.

---

## Last Session (2026-02-06) - Fix Vue Rendering Errors

### What Was Done

#### 1. Fixed Vue `insertBefore` null crashes (issue #04)
**Root cause**: `lucide.createIcons()` replaces `<i>` elements with `<svg>` in the DOM — external DOM mutation that breaks Vue's VDOM patcher. When Pinia store updates triggered re-renders, Vue referenced nodes that no longer existed.

**Fix**: Rewrote `admin/src/lib/icons.js` to inject SVGs **inside** `<i>` elements instead of replacing them. Vue still owns the outer element, patches work correctly.

#### 2. Fixed router timing race condition
Changed `main.js` to `router.isReady().then(() => app.mount('#app'))` — prevents router's `install()` from triggering navigation before DOM exists.

#### 3. Extracted App.js into 7 granular components
Broke 775-line monolith into: Sidebar, HeaderBar, CommandPalette, SettingsPanel, NewAppModal, CreateAliasModal, EditAliasModal. Single root element per component.

#### 4. Consolidated dashboard to stat cards
Replaced 3-panel layout (System + Apps table + Aliases table) with single panel of 7 stat cards. Clickable overview cards navigate to detail pages.

#### 5. Updated fazt-app skill documentation
Added "External DOM Mutation (CRITICAL)" section to `frontend-patterns.md` — documents the rule, common offenders, canonical fix pattern, and why it crashes.

---

## Open Items

### fazt.http - External HTTP Calls in Serverless
**Priority**: HIGH - Architectural decision needed
See `koder/STATE.md` history for full context. Key decisions: security model (SSRF), async model (blocking vs callbacks), timeout budget.

---

## Previous Session (2026-02-05) - v0.25.4 Release

### Released v0.25.4 (Bug Fix)
- Fixed admin UI health endpoint: `requireAPIKeyAuth` → `requireAdminAuth`
- Released and deployed to all peers

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
