# Local Development Setup

**Updated**: 2026-02-02

Detailed setup instructions for fazt local development.

## Remote Server Access

Production instances run behind Cloudflare. For SSH access, use the actual
server IP (not the domain). IPs are stored in `.env`:

```bash
source .env
ssh root@$ZYT_IP   # SSH into production server
```

The VM has SSH access to remote servers. Use this for emergency recovery
or manual deployments when the `fazt peer upgrade` command can't reach
the server.

## Local Server Setup

The local server runs as a **user systemd service** (`fazt-local`).
It auto-starts on boot and persists across sessions.

### Quick Commands

```bash
systemctl --user status fazt-local    # Check status
systemctl --user restart fazt-local   # Restart after rebuild
journalctl --user -u fazt-local -f    # View logs
```

### First-Time Setup

**1. Build fazt with embedded admin:**
```bash
npm run build --prefix admin
cp -r admin/dist internal/assets/system/admin
go build -o fazt ./cmd/server
```

**2. Install local server:**

Use the unified install script:
```bash
./install.sh  # Select option 2: Local Development
```

Or manually:
```bash
mkdir -p servers/local
fazt server init \
  --username dev \
  --password dev \
  --domain 192.168.64.3 \
  --db servers/local/data.db
```

**3. Create API key and add as peer:**
```bash
fazt server create-key --db servers/local/data.db
# Save the token output

fazt peer add local \
  --url http://192.168.64.3:8080 \
  --token <API_KEY>
```

**4. Start local server:**
```bash
systemctl --user start fazt-local
```

Or manually (for one-off testing):
```bash
fazt server start \
  --port 8080 \
  --domain 192.168.64.3 \
  --db servers/local/data.db
```

### Quick Deploy (Static Only)

For apps without serverless (`/api`) endpoints:

```bash
python3 -m http.server 7780 --directory servers/zyt/my-app
# Access at http://192.168.64.3:7780
```

### Deploy and Test

```bash
fazt @local app deploy servers/zyt/my-app
# Access at http://my-app.192.168.64.3.nip.io:8080

# Or with curl:
curl -H "Host: my-app.192.168.64.3" http://192.168.64.3:8080/api/hello
```

## Troubleshooting

### Verbose Output

Use `--verbose` flag to see detailed output including database migrations and debug info:

```bash
fazt --verbose @local app list
fazt --verbose peer status
fazt sql "..." --verbose
```

This is useful when:
- Debugging database issues
- Understanding what migrations are running
- Troubleshooting slow commands
- Reporting bugs (include verbose output)

### Common Issues

**Service not starting:**
```bash
journalctl --user -u fazt-local -n 50    # Check logs
systemctl --user restart fazt-local       # Restart service
```

**Database locked errors:**
```bash
# Check if multiple processes are accessing the DB
lsof ~/.fazt/data.db
```
