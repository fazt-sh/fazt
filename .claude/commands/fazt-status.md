---
description: Check fazt server status and health
model: haiku
allowed-tools: Read, Glob, Bash, WebFetch
---

# Fazt Server Status

Check the health and status of a fazt instance.

## Server Selection

1. List available servers:
   ```bash
   ls servers/
   ```

2. If argument `$ARGUMENTS` is provided, use that as server name
3. If only one server exists, use it automatically
4. If multiple servers and no argument, ask user which to use

## Read Server Config

Once server is selected, read its config:
```
servers/<name>/config.json
```

Expected format:
```json
{
  "name": "server-name",
  "url": "https://admin.example.com",
  "domain": "example.com",
  "description": "Description"
}
```

## Health Checks

Perform these checks and report status:

### 1. Admin API Health
```bash
curl -s -o /dev/null -w "%{http_code}" <url>/health
```
Expected: 200

### 2. Public Domain Check
```bash
curl -s -o /dev/null -w "%{http_code}" https://<domain>/
```
Expected: 200 or redirect

### 3. API Version (if available)
```bash
curl -s <url>/api/version
```

## Output Format

Report as a concise status summary:

```
Server: <name> (<domain>)
Admin:  <url>

Health:
  Admin API:    [OK/FAIL] (HTTP <code>)
  Public Site:  [OK/FAIL] (HTTP <code>)
  Version:      <version or unknown>
```

## Error Handling

- If `servers/` doesn't exist: "No servers configured. Create servers/<name>/config.json"
- If server not found: "Server '<name>' not found. Available: <list>"
- If URL unreachable: Report connection error, continue with other checks
