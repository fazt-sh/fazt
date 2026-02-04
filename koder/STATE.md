# Fazt Implementation State

**Last Updated**: 2026-02-04
**Current Version**: v0.24.13

## Status

State: CLEAN
Activity logging with permissive URL parsing for alias filter.

---

## Current Session (2026-02-04) - Permissive URL Parsing

### What Was Done

#### Permissive Alias URL Parsing
Enhanced `--alias` filter to accept URLs in any format:
- Added `normalizeAlias()` function that extracts subdomain from various inputs:
  - Full URLs: `https://fun-game.zyt.app/path` → `fun-game`
  - Domain only: `fun-game.zyt.app` → `fun-game`
  - With path/query: `fun-game.zyt.app/game?level=5` → `fun-game`
  - Just alias: `fun-game` → `fun-game`
- Works seamlessly for both local and remote peer queries
- Updated help text and examples to show flexible format
- Users can now copy-paste full URLs directly from browser

#### Flag Aliases: --url and --link
Added intuitive aliases for the `--alias` flag:
- `--url` - matches user mental model of "I have a URL"
- `--link` - natural shorthand for "I have a link"
- All three flags (`--alias`, `--url`, `--link`) are complete synonyms
- Same permissive URL parsing applies to all variants

**Example usage:**
```bash
# All of these work identically:
fazt logs list --alias tetris
fazt logs list --url https://tetris.zyt.app/
fazt logs list --link tetris.zyt.app/game
```

### Commits
```
ab0faf4 feat: add --url and --link aliases for --alias flag
1a822ba feat: permissive URL parsing for --alias filter
```

---

## Last Session (2026-02-04) - Activity Logging System

### What Was Done

#### 1. Unified Activity Logging (v0.24.8 - v0.24.12)
Created comprehensive activity logging to replace separate audit/analytics systems:
- **`internal/activity/`** package with buffered writes (10s flush interval)
- **Weight-based prioritization** (0-9 scale): Security(9) → Debug(0)
- **Full query/filter support**: app, user, action, result, time range, alias
- **CLI commands**: `fazt logs list|stats|cleanup|export`
- **Remote peer support**: `fazt @zyt logs list`
- **Analytics SDK injection**: Auto-injects tracking script into HTML pages

#### 2. Alias Filter (v0.24.13)
- Added `--alias` filter to logs command
- Filters pageviews by subdomain: `fazt logs list --alias fun-game`
- Works across list, stats, cleanup, export commands
- Remote peer support included

#### 3. Analytics Tracking Fix (v0.24.12)
- Fixed tracking URL to use centralized `admin.<domain>/track`
- All app subdomains now send analytics to admin subdomain
- Script extracts domain from hostname for proper routing

### Key Files Created/Modified
- `internal/activity/logger.go` - Buffer, logging helpers, weight constants
- `internal/activity/query.go` - Query, cleanup, stats with full filter support
- `internal/hosting/analytics_inject.go` - SDK injection before `</body>`
- `cmd/server/logs.go` - CLI commands (local + remote)
- `internal/handlers/system.go` - API endpoints for remote logs
- `internal/database/migrations/018_activity_log.sql` - Schema

### Commits
```
01ca974 release: v0.24.13 - add --alias filter to logs command
3fac9cf release: v0.24.12 - fix analytics tracking URL
... (multiple commits from v0.24.8 through v0.24.13)
```

---

## Next Up

### High Priority
1. **Root domain tracking** - Currently no pageviews for `zyt.app` (only subdomains tracked)
2. **Consider `--domain` filter** - Match full domain pattern vs alias prefix
3. **Admin UI integration** - Display activity logs in web dashboard

### Future Work
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
cat version.json | jq -r '.version'                                    # 0.24.13
grep "var Version" internal/config/config.go | grep -oE '[0-9..]+'    # 0.24.13
fazt --version                                                         # 0.24.13

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
