# API Reference

**Updated**: 2026-02-04

## Global CLI Flags

These flags work with all commands and can be placed anywhere:

| Flag | Description |
|------|-------------|
| `--verbose` | Show detailed output (database migrations, debug info) |
| `--format <fmt>` | Output format: `markdown` (default) or `json` |

**Examples**:
```bash
fazt --verbose @local app list      # Show migrations
fazt peer list --format json        # JSON output
fazt @zyt sql "..." --verbose       # Verbose + remote execution
```

---

## Admin API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/deploy` | POST | Deploy ZIP archive |
| `/api/apps` | GET | List apps |
| `/api/apps/{id}` | GET/DELETE | App details/delete |
| `/api/apps/{id}/source` | GET | App source tracking |
| `/api/apps/{id}/files` | GET | List app files |
| `/api/apps/{id}/files/{path}` | GET | Get file content |
| `/api/system/health` | GET | Health check |
| `/api/system/logs` | GET | Activity logs (with query params) |
| `/api/system/logs/stats` | GET | Activity log statistics |
| `/api/system/logs/cleanup` | POST | Delete logs (with filters) |
| `/api/upgrade` | POST | Upgrade server |

## API Response Format

```go
// Success
api.Success(w, http.StatusOK, data)

// Errors
api.BadRequest(w, "message")
api.ValidationError(w, "message", "field", "constraint")
api.InternalError(w, err)
```

## CLI Commands

### Peer Management

```bash
fazt peer list                      # List all peers
fazt peer list --format json        # List peers in JSON format
fazt @zyt status                    # Health, version, uptime
fazt @zyt upgrade                   # Upgrade to latest version
fazt peer add <name> --url <url> --token <token>
```

### App Management

```bash
fazt app list                       # List local apps
fazt @zyt app list                  # List apps on remote peer
fazt app deploy <dir>               # Deploy to local
fazt @zyt app deploy <dir>          # Deploy to remote peer
fazt @zyt app deploy <dir> --include-private  # Include private/
fazt @zyt app install <url>         # Install from GitHub
fazt @zyt app upgrade <app>         # Upgrade git-sourced app
fazt @zyt app pull <app> --to ./local    # Download app files
fazt @zyt app info <app>            # Show app details
fazt @zyt app remove <app>          # Remove an app
fazt @zyt app logs <app> -f         # Tail logs
```

### SQL Queries

```bash
fazt sql "SELECT * FROM apps"       # Query local database
fazt @zyt sql "SELECT * FROM apps"  # Query remote database
fazt sql "..." --format json        # JSON output
```

### Activity Logs

```bash
fazt logs list                      # Recent activity (default: 20)
fazt logs list --limit 50           # More entries
fazt logs stats                     # Overview statistics
fazt logs cleanup --max-weight 2    # Preview cleanup (dry-run)
fazt logs cleanup --max-weight 2 --force  # Actually delete
fazt logs export -f csv -o out.csv  # Export to file

# Remote peers
fazt @zyt logs list
fazt @local logs stats
```

**Filters** (work with list, stats, cleanup, export):

| Filter | Description | Example |
|--------|-------------|---------|
| `--action` | Filter by action | `--action pageview` |
| `--app` | Filter by app ID | `--app my-app` |
| `--user` | Filter by user ID | `--user abc123` |
| `--min-weight` | Minimum weight (0-9) | `--min-weight 5` |
| `--max-weight` | Maximum weight (0-9) | `--max-weight 2` |
| `--type` | Resource type | `--type page` |
| `--result` | success/failure | `--result failure` |
| `--since` | Time range start | `--since 24h` |
| `--until` | Time range end | `--until 1h` |
| `--limit` | Max results | `--limit 100` |
| `--offset` | Pagination offset | `--offset 20` |

**Weight Scale** (0-9, higher = more important):

| Wt | Category | Actions |
|----|----------|---------|
| 9 | Security | API key create/delete |
| 8 | Auth | `login`, `logout` |
| 7 | Config | Alias/redirect changes |
| 6 | Deployment | App deploy/delete |
| 5 | Data | KV/doc mutations |
| 4 | User Action | Form submissions |
| 3 | Navigation | Internal page views |
| 2 | Analytics | `pageview`, `click` |
| 1 | System | `start`, `stop`, `upgraded` |
| 0 | Debug | Timing, cache hits |

**Resource Types**: `server`, `session`, `page`, `app`, `kv`, `doc`, `config`

**Actor Types**: `user`, `system`, `api_key`, `anonymous`

### Server Management

```bash
fazt server init --username <u> --password <p> --domain <d> --db <path>
fazt server start --port <p> --domain <d> --db <path>
fazt server create-key --db <path>
```
