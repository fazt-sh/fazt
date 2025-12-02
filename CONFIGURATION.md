# Configuration Guide

## Overview

Fazt uses a flexible JSON-based configuration system with support for CLI flags and environment variables.

## Priority

1. **CLI Flags** (Highest)
2. **JSON Config File**
3. **Environment Variables**
4. **Defaults** (Lowest)

## Default Locations

- **Config File**: `~/.config/fazt/config.json`
- **Database**: `~/.config/fazt/data.db`

## Configuration File Format

```json
{
  "server": {
    "port": "4698",
    "domain": "https://fazt.sh",
    "env": "production"
  },
  "database": {
    "path": "~/.config/fazt/data.db"
  },
  "auth": {
    "username": "admin",
    "password_hash": "$2a$12$..."
  },
  "ntfy": {
    "topic": "your-topic",
    "url": "https://ntfy.sh"
  },
  "api_key": {
    "token": "...",
    "name": "deployment-token"
  }
}
```

## CLI Commands

```bash
fazt <command> [flags] [arguments]
```

### Server Management

| Command | Description |
|---------|-------------|
| `fazt server init` | Initialize server config & auth |
| `fazt server start` | Start the server |
| `fazt server status` | Check status |
| `fazt server set-credentials` | Update password |
| `fazt server set-config` | Update settings |

### Client / Deploy

| Command | Description |
|---------|-------------|
| `fazt deploy` | Deploy current directory |
| `fazt client set-auth-token` | Set local deployment token |

## Environment Variables

| Variable | Config Equivalent |
|----------|-------------------|
| `PORT` | `server.port` |
| `DB_PATH` | `database.path` |
| `ENV` | `server.env` |
| `FAZT_DOMAIN` | `server.domain` |

## Security

The server automatically enforces secure permissions:
- Config file: `0600` (User RW)
- Config dir: `0700` (User RWX)
