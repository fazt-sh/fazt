---
command: "logs"
description: "Activity log management - query, analyze, and clean up activity logs"
syntax: "fazt logs <command> [options]"
version: "0.24.13"
updated: "2026-02-04"

examples:
  - title: "List recent activity"
    command: "fazt logs list"
    description: "Show the 20 most recent activity log entries"
  - title: "Filter by subdomain using URL"
    command: "fazt logs list --url https://tetris.zyt.app/game"
    description: "Show pageviews for tetris subdomain (accepts any URL format)"
  - title: "Filter by alias"
    command: "fazt logs list --alias fun-game"
    description: "Show pageviews for fun-game subdomain"
  - title: "Filter by app with time range"
    command: "fazt logs list --app my-app --since 24h"
    description: "Show activity for my-app in the last 24 hours"
  - title: "High-priority events only"
    command: "fazt logs list --min-weight 5"
    description: "Show only important events (data mutations, deploys, auth, security)"
  - title: "Get activity statistics"
    command: "fazt logs stats"
    description: "Show statistics about activity logs (count by weight, size, time range)"
  - title: "Preview cleanup"
    command: "fazt logs cleanup --max-weight 2 --until 7d"
    description: "Preview what would be deleted (old analytics/pageviews)"
  - title: "Execute cleanup"
    command: "fazt logs cleanup --max-weight 2 --until 7d --force"
    description: "Actually delete old low-priority entries"
  - title: "Export to CSV"
    command: "fazt logs export --app my-app -f csv -o logs.csv"
    description: "Export app activity to CSV file"
  - title: "Remote peer logs"
    command: "fazt @zyt logs list --link tetris.zyt.app"
    description: "Query activity logs on remote peer"

related:
  - command: "app"
    description: "App management commands"
  - command: "user"
    description: "User management commands"
---

# fazt logs

Unified activity logging system for querying, analyzing, and managing activity logs. Tracks security events, deployments, authentication, data mutations, pageviews, and more with weight-based prioritization.

## Commands

- `list` - List activity log entries with filtering
- `stats` - Show activity log statistics
- `cleanup` - Delete entries matching filters (preview by default, use `--force` to delete)
- `export` - Export entries to JSON or CSV

## Filter Options

All filter options work across all commands (list, stats, cleanup, export):

### Weight Filters
- `--min-weight N` - Minimum weight (0-9)
- `--max-weight N` - Maximum weight (0-9)

### Resource Filters
- `--app ID` - Filter by app ID (matches app resources, KV, docs, pages)
- `--alias URL` - Filter by subdomain/alias (accepts any URL format)
- `--url URL` - Alias for `--alias` (more intuitive for URLs)
- `--link URL` - Alias for `--alias` (quick shorthand)
- `--type T` - Filter by resource type (app/user/session/kv/doc/page/config)
- `--resource ID` - Filter by resource ID

### Actor Filters
- `--user ID` - Filter by user ID
- `--actor-type T` - Filter by actor type (user/system/api_key/anonymous)

### Action Filters
- `--action A` - Filter by action (e.g., pageview, deploy, login, create, delete)
- `--result R` - Filter by result (success/failure)

### Time Filters
- `--since TIME` - Show entries since (e.g., '24h', '7d', '2024-01-15')
- `--until TIME` - Show entries until (same format)

### Pagination
- `--limit N` - Number of entries (default: 20, max: 1000 for export)
- `--offset N` - Skip first n results

## Permissive URL Parsing

The `--alias`, `--url`, and `--link` flags accept URLs in any format:

```bash
# All of these are equivalent:
fazt logs list --alias tetris
fazt logs list --alias tetris.zyt.app
fazt logs list --url https://tetris.zyt.app/
fazt logs list --link https://tetris.zyt.app/game?level=5
```

The system automatically extracts the subdomain/alias from the URL. Just copy-paste from your browser!

## Weight Scale (0-9)

Activity logs use a weight-based priority system:

| Weight | Category | Examples |
|--------|----------|----------|
| 9 | Security | API key changes, role changes, security events |
| 8 | Auth | Login, logout, session creation/destruction |
| 7 | Config | Alias changes, redirect updates, domain config |
| 6 | Deployment | App deploy, app delete, version updates |
| 5 | Data Mutation | KV writes, doc updates, blob uploads |
| 4 | User Action | Form submissions, user-triggered operations |
| 3 | Navigation | Page navigation within apps |
| 2 | Analytics | Pageviews, clicks, tracking events |
| 1 | System | Health checks, server start/stop |
| 0 | Debug | Timing info, cache hits, development logs |

Use weight filters to focus on what matters:
- `--min-weight 5` - Only important events (data, deploy, auth, security)
- `--max-weight 2` - Only low-priority events (analytics, system)

## Command-Specific Options

### cleanup
- `--force` - Actually delete entries (default: preview only)

**Safety**: Cleanup always runs in dry-run mode unless `--force` is specified. Review the count before forcing deletion.

### export
- `-f FORMAT` - Output format: `json` or `csv` (default: json)
- `-o FILE` - Output file (default: stdout)

**Note**: Export defaults to limit=1000. Use `--limit` to override.

## Remote Execution

All logs commands support remote peer execution:

```bash
fazt @zyt logs list --url https://tetris.zyt.app/
fazt @zyt logs stats --app my-app
fazt @zyt logs cleanup --max-weight 2 --until 7d --force
fazt @zyt logs export --app my-app --limit 5000 -f csv -o remote-logs.csv
```

## Common Workflows

### Monitor recent important activity
```bash
fazt logs list --min-weight 5 --since 1h
```

### Track app pageviews
```bash
fazt logs list --app my-app --action pageview --since 7d
```

### Clean up old analytics
```bash
# Preview first
fazt logs cleanup --max-weight 2 --until 30d

# Then execute
fazt logs cleanup --max-weight 2 --until 30d --force
```

### Export user activity for compliance
```bash
fazt logs export --user user_abc123 --limit 10000 -f csv -o user-activity.csv
```

### Check subdomain traffic
```bash
# Just paste the URL from your browser
fazt logs list --url https://my-game.zyt.app/play
fazt logs stats --link my-game.zyt.app
```

## API Endpoints

- `GET /api/system/logs` - Query logs with filters
- `GET /api/system/logs/stats` - Get statistics
- `POST /api/system/logs/cleanup` - Delete entries matching filters

All endpoints require admin/owner role (session auth) or API key auth.

## Architecture

Activity logs use:
- **Buffered writes** - 10-second flush interval for performance
- **Weight-based indexing** - Fast filtering by priority
- **Full-text search** - Efficient resource/action/result filtering
- **SQLite storage** - Transactional integrity, easy backup
- **Automatic injection** - Analytics tracking automatically injected into HTML pages

Logs are stored in the `activity_log` table in the main SQLite database.
