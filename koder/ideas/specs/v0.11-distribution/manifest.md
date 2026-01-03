# App Manifest (app.json)

## Summary

`app.json` is the manifest file that defines an app's metadata, permissions,
configuration, and capabilities. It enables declarative app configuration and
permission-based security.

## Location

```
my-app/
├── app.json            # Manifest (required for marketplace)
├── index.html
└── api/
    └── main.js
```

## Full Schema

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "description": "A sample application",
  "author": "Alice <alice@example.com>",
  "license": "MIT",
  "homepage": "https://github.com/alice/my-app",

  "fazt": {
    "minVersion": "0.10.0"
  },

  "permissions": [
    "storage:kv",
    "storage:ds",
    "net:fetch"
  ],

  "env": [
    {
      "name": "API_KEY",
      "description": "External API key",
      "required": true
    },
    {
      "name": "DEBUG",
      "description": "Enable debug mode",
      "required": false,
      "default": "false"
    }
  ],

  "cron": [
    {
      "schedule": "0 * * * *",
      "handler": "api/hourly.js"
    }
  ],

  "visibility": "public"
}
```

## Fields

### Metadata

| Field         | Required | Description                            |
| ------------- | -------- | -------------------------------------- |
| `name`        | Yes      | App identifier (alphanumeric + hyphen) |
| `version`     | Yes      | Semver version string                  |
| `description` | No       | Short description (max 200 chars)      |
| `author`      | No       | Author name and email                  |
| `license`     | No       | SPDX license identifier                |
| `homepage`    | No       | Project URL                            |

### Fazt Requirements

```json
{
  "fazt": {
    "minVersion": "0.10.0",
    "maxVersion": "0.12.0"
  }
}
```

Installation fails if Fazt version is outside range.

### Permissions

Permissions follow a `category:action` pattern:

| Permission      | Description                    |
| --------------- | ------------------------------ |
| `storage:kv`    | Key-value storage              |
| `storage:ds`    | Document storage               |
| `storage:s3`    | Blob storage                   |
| `net:fetch`     | HTTP requests to external URLs |
| `net:vpn`       | Access VPN status (v0.13+)     |
| `kernel:deploy` | Deploy other apps (v0.12+)     |
| `kernel:status` | Read system status             |
| `security:sign` | Use kernel signing (v0.14+)    |
| `ai:complete`   | Use AI completions (v0.12+)    |

### Permission Enforcement

```javascript
// In serverless handler
const data = await fazt.storage.ds.find('users', {});

// If app lacks 'storage:ds' permission:
// Error: Permission denied: storage:ds
```

### Environment Variables

```json
{
  "env": [
    {
      "name": "OPENAI_API_KEY",
      "description": "API key for OpenAI",
      "required": true
    }
  ]
}
```

During installation:

```
Installing my-app v1.0.0...

Required configuration:
  OPENAI_API_KEY: API key for OpenAI
  > Enter value: sk-...

App installed successfully.
```

### Cron Jobs

```json
{
  "cron": [
    {
      "schedule": "*/5 * * * *",
      "handler": "api/sync.js",
      "timeout": 60000,
      "skipIfRunning": true
    }
  ]
}
```

See `v0.10-runtime/cron.md` for details.

### Visibility

| Value      | Description                        |
| ---------- | ---------------------------------- |
| `public`   | Accessible from internet (default) |
| `vpn-only` | Only via VPN tunnel (v0.13+)       |
| `internal` | Only via kernel IPC                |

## Validation

### On Deploy

```bash
fazt deploy ./my-app

# Validates:
# - app.json exists and is valid JSON
# - name matches slug pattern
# - version is valid semver
# - permissions are recognized
# - env vars have valid structure
```

### On Install

```bash
fazt app install blog

# Validates all above, plus:
# - fazt.minVersion satisfied
# - fazt.maxVersion satisfied
# - Required env vars prompted
```

## Default Permissions

If no `app.json` exists, app gets minimal permissions:

```json
{
  "permissions": [
    "storage:kv"
  ]
}
```

No external network, no document store, no system access.

## Upgrades

When updating an app, permission changes require approval:

```
Updating blog v1.2.0 → v1.3.0

New permissions requested:
  + ai:complete    Use AI completions

Proceed? [y/N]
```

## Examples

### Simple Static Site

```json
{
  "name": "portfolio",
  "version": "1.0.0",
  "description": "My personal portfolio"
}
```

### Backend App

```json
{
  "name": "crm",
  "version": "2.0.0",
  "description": "Customer relationship manager",
  "permissions": [
    "storage:ds",
    "net:fetch"
  ],
  "env": [
    { "name": "SMTP_HOST", "required": true }
  ],
  "cron": [
    { "schedule": "0 9 * * *", "handler": "api/daily-report.js" }
  ]
}
```

### AI-Powered App

```json
{
  "name": "summarizer",
  "version": "1.0.0",
  "permissions": [
    "storage:kv",
    "ai:complete"
  ],
  "env": [
    { "name": "OPENAI_API_KEY", "required": true }
  ]
}
```
