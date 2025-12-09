# Fazt.sh Assistant Guide

**Fazt** is a personal PaaS in a single Go binary - deploy static sites and serverless functions to your own VPS with zero dependencies.

## Quick Start
```
read and execute koder/start.md
```

## Core Philosophy
- **Cartridge Architecture**: One Binary (`fazt`) + One SQLite DB (`data.db`)
- **Zero Dependencies**: Pure Go + `modernc.org/sqlite`. NO CGO.
- **VFS**: Sites/assets live in DB, not filesystem
- **Safety**: `CGO_ENABLED=0` always

## Build & Test
```bash
go build -o fazt ./cmd/server          # Build
go test ./...                           # Test all
go run ./cmd/server server start --port 8080  # Run dev
```

## Key Directories
- `cmd/server/` - Entry point, CLI
- `internal/handlers/` - HTTP handlers
- `internal/api/` - Response helpers
- `internal/hosting/` - VFS, Deploy
- `admin/` - React SPA source
- `koder/plans/` - Implementation plans

## API Response Format
```go
// Success
api.Success(w, http.StatusOK, data)

// Errors
api.BadRequest(w, "message")
api.ValidationError(w, "message", "field", "constraint")
api.InternalError(w, err)
```

## Environment
- Running in podman container (no systemd)
- Git credentials available
- Deploy target: Ubuntu 24.04 x86

## Current Work
See `koder/NEXT_SESSION.md` for active task and context.
