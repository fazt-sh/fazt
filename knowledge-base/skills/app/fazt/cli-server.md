# fazt server - Server Management

Initialize and run fazt server instances.

## Commands

### fazt server init

Initialize a new fazt server (creates config and database).

```bash
fazt server init \
  --username <admin-user> \
  --password <admin-pass> \
  --domain <domain> \
  --db <path-to-db>
```

**Example:**
```bash
fazt server init \
  --username admin \
  --password secret123 \
  --domain example.com \
  --db ./data.db
```

### fazt server start

Start the server manually (for debugging/testing).

```bash
fazt server start --db <path-to-db>
fazt server start --db <path-to-db> --port 8080
```

**Options:**
- `--port` - HTTP port (default: 80 or 443)
- `--db` - Path to SQLite database
- `--domain` - Override domain

**Note**: For production, use systemd service instead of manual start.

### fazt server status

Show configuration and server status.

```bash
fazt server status --db <path-to-db>
```

### fazt server set-credentials

Update admin credentials (password reset).

```bash
fazt server set-credentials \
  --username <user> \
  --password <new-pass> \
  --db <path-to-db>
```

### fazt server set-config

Update server settings.

```bash
fazt server set-config \
  --domain <new-domain> \
  --db <path-to-db>
```

### fazt server create-key

Create an API key for deployments.

```bash
fazt server create-key --db <path-to-db>
fazt server create-key --name "my-laptop" --db <path-to-db>
```

**Output:**
```
API Key created successfully:
fzt_abc123def456...

Save this key - it won't be shown again.
```

### fazt server reset-admin

Reset admin dashboard to embedded version.

```bash
fazt server reset-admin --db <path-to-db>
```

## Systemd Service

For production, fazt runs as a systemd service.

**User service (local development):**
```bash
# Install
./install.sh  # Select option 2: Local Development

# Manage
systemctl --user start fazt-local
systemctl --user stop fazt-local
systemctl --user status fazt-local
journalctl --user -u fazt-local -f
```

**System service (production):**
```bash
# Install
sudo ./install.sh  # Select option 1: Production

# Manage
sudo systemctl start fazt
sudo systemctl status fazt
sudo journalctl -u fazt -f
```

## Typical Server Setup

```bash
# 1. Initialize
fazt server init \
  --username admin \
  --password supersecret \
  --domain mysite.com \
  --db /var/lib/fazt/data.db

# 2. Create API key for deployments
fazt server create-key --name "deploy-key" --db /var/lib/fazt/data.db

# 3. Install as service
sudo ./install.sh

# 4. Configure reverse proxy (nginx/caddy) if needed
```
