---
command: "alias"
description: "Alias management commands"
syntax: "fazt alias <command> [options]"
version: "0.24.7"
updated: "2026-02-04"

examples:
  - title: "List all aliases"
    command: "fazt alias list"
    description: "List aliases on the local instance"
  - title: "List with pagination"
    command: "fazt alias list --limit 50"
    description: "List first 50 aliases"
  - title: "Show alias info"
    command: "fazt alias info --name myapp"
    description: "Show details for a specific alias"

related:
  - command: "app"
    description: "App management commands"
  - command: "user"
    description: "User management commands"
---

# fazt alias

Manage routing aliases - subdomains that map to apps.

## Commands

- `list` - List all aliases with pagination
- `info` - Show alias details (requires `--name`)

## Options

### Common Options
- `--name <subdomain>` - Alias subdomain name

### List Options
- `--offset <n>` - Skip first n results (default: 0)
- `--limit <n>` - Max results to return (default: 20)

## Alias Types

- **app** - Routes to an application (most common)
- **redirect** - HTTP redirect to a URL
- **reserved** - Reserved/blocked subdomain
- **split** - Traffic splitting between multiple apps

## Remote Execution

```bash
fazt @zyt alias list
fazt @zyt alias info --name myapp
```

## API Endpoints

- `GET /api/aliases` - List aliases (paginated)
- `GET /api/aliases/{subdomain}` - Alias details
- `POST /api/aliases` - Create alias
- `PUT /api/aliases/{subdomain}` - Update alias
- `DELETE /api/aliases/{subdomain}` - Delete alias

Endpoints require admin/owner role via AdminMiddleware.
