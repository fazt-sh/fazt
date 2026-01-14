---
description: Deploy a site/app to a fazt server
model: sonnet
allowed-tools: Read, Glob, Bash, WebFetch
---

# Fazt Deploy

Deploy a local directory as a site/app to a fazt instance.

## Arguments

`$ARGUMENTS` format: `<directory> [site_name] [--server=<name>]`

Examples:
- `/fazt-deploy sites/my-site` - Deploy directory, infer site name
- `/fazt-deploy . myapp` - Deploy current dir as "myapp"
- `/fazt-deploy sites/blog blog --server=zyt` - Deploy to specific server

## Server Selection

1. List available servers:
   ```bash
   ls servers/
   ```

2. Parse `--server=<name>` from arguments if provided
3. If only one server exists, use it
4. If multiple servers and no `--server`, ask user

## Read Server Config

```
servers/<name>/config.json
```

Expected format:
```json
{
  "name": "server-name",
  "url": "https://admin.example.com",
  "domain": "example.com",
  "token": "your-api-key"
}
```

**Important**: If `token` is missing, check for environment variable:
```bash
echo $FAZT_TOKEN_<NAME>
```
Where `<NAME>` is uppercase server name (e.g., `FAZT_TOKEN_ZYT`).

If no token found, error: "No API token configured for server '<name>'"

## Prepare Deployment

1. **Validate directory exists**:
   ```bash
   ls <directory>
   ```

2. **Check for manifest.json** (optional):
   ```bash
   cat <directory>/manifest.json 2>/dev/null
   ```

3. **Determine site name** (in order of precedence):
   - Explicit argument
   - `name` field from manifest.json
   - Directory basename

4. **Create ZIP archive**:
   ```bash
   cd <directory> && zip -r /tmp/deploy-<site_name>.zip .
   ```

## Execute Deploy

```bash
curl -X POST "<url>/api/deploy" \
  -H "Authorization: Bearer <token>" \
  -F "site_name=<site_name>" \
  -F "file=@/tmp/deploy-<site_name>.zip"
```

## Output Format

On success:
```
Deployed: <site_name>
URL: https://<site_name>.<domain>
Files: <count>
Size: <bytes>
```

On failure, report the error message from the API.

## Cleanup

Remove temporary ZIP file:
```bash
rm /tmp/deploy-<site_name>.zip
```

## Error Handling

- Directory not found: "Directory '<path>' not found"
- No token: "No API token for server '<name>'. Set FAZT_TOKEN_<NAME> or add token to config.json"
- API error: Report error message from response
- Network error: Report curl error
