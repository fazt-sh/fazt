# API Reference

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
fazt remote list                    # List all peers
fazt remote status zyt              # Health, version, uptime
fazt remote upgrade zyt             # Upgrade to latest version
fazt remote add <name> --url <url> --token <token>
```

### App Management

```bash
fazt app list <peer>                # List deployed apps
fazt app deploy <dir> --to <peer>   # Deploy from local directory
fazt app deploy <dir> --to <peer> --include-private  # Include private/
fazt app install <url> --to <peer>  # Install from GitHub
fazt app upgrade <app> --on <peer>  # Upgrade git-sourced app
fazt app pull <app> --to ./local    # Download app files
fazt app info <app> --on <peer>     # Show app details
fazt app remove <app> --from <peer> # Remove an app
fazt app logs <app> --on <peer> -f  # Tail logs
```

### Server Management

```bash
fazt server init --username <u> --password <p> --domain <d> --db <path>
fazt server start --port <p> --domain <d> --db <path>
fazt server create-key --db <path>
```
