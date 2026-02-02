# API Reference

**Updated**: 2026-02-02

## Admin API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/deploy` | POST | Deploy ZIP archive |
| `/api/apps` | GET | List apps |
| `/api/apps/{id}` | GET/DELETE | App details/delete |
| `/api/apps/{id}/source` | GET | App source tracking |
| `/api/apps/{id}/files` | GET | List app files |
| `/api/apps/{id}/files/{path}` | GET | Get file content |
| `/api/upgrade` | POST | Upgrade server |
| `/health` | GET | Health check |

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
fazt peer status zyt                # Health, version, uptime
fazt peer upgrade zyt               # Upgrade to latest version
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

### Server Management

```bash
fazt server init --username <u> --password <p> --domain <d> --db <path>
fazt server start --port <p> --domain <d> --db <path>
fazt server create-key --db <path>
```
