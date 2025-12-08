# Upgrade to v0.7.2 Guide

**Version**: v0.7.2
**Release Date**: December 8, 2025
**Compatibility**: Upgrades from v0.7.0, v0.7.1

---

## 🔍 Pre-Upgrade Checklist

Before upgrading, verify:

1. **Current Version**: Run `fazt --version` on your droplet
2. **Service Status**: Check `systemctl status fazt` - should be "active (running)"
3. **Database Backup**: Recommended (optional, migrations are non-destructive)
   ```bash
   sudo cp ~fazt/.config/fazt/data.db ~fazt/.config/fazt/data.db.backup
   ```

---

## ✅ Upgrade Methods

### Method 1: Built-in Upgrade Command (RECOMMENDED)

The `fazt upgrade` command handles everything automatically:

```bash
# SSH into your droplet
ssh root@your-droplet-ip

# Run upgrade
sudo fazt upgrade

# It will:
# 1. Check for new version
# 2. Download v0.7.2 binary
# 3. Replace binary
# 4. Restart systemd service automatically
```

**What happens:**
- Downloads new binary from GitHub releases
- Backs up current binary to `/usr/local/bin/fazt.old`
- Replaces binary atomically
- Applies `setcap` for port binding (< 1024)
- Detects active systemd service
- Runs `systemctl restart fazt` automatically
- **Zero downtime** (< 2 seconds during restart)

---

### Method 2: Re-run install.sh (Alternative)

⚠️ **IMPORTANT**: This method requires stopping the service first.

```bash
# SSH into your droplet
ssh root@your-droplet-ip

# 1. Stop the service FIRST
sudo systemctl stop fazt

# 2. Run install script
curl -fsSL https://github.com/fazt-sh/fazt/raw/master/install.sh | bash

# 3. Select Option 1 (Headless Server)
# 4. Enter your domain when prompted

# Service will restart automatically
```

**Why stop first?**
- `install.sh` checks if ports 80/443 are available
- If service is running, port check fails
- Stopping first ensures clean upgrade

---

### Method 3: Manual Binary Replacement (Advanced)

For custom deployments or troubleshooting:

```bash
# 1. Download new binary
wget https://github.com/fazt-sh/fazt/releases/download/v0.7.2/fazt-v0.7.2-linux-amd64.tar.gz

# 2. Extract
tar -xzf fazt-v0.7.2-linux-amd64.tar.gz

# 3. Stop service
sudo systemctl stop fazt

# 4. Replace binary
sudo mv fazt /usr/local/bin/fazt
sudo chmod +x /usr/local/bin/fazt

# 5. Apply capabilities (required for port binding)
sudo setcap CAP_NET_BIND_SERVICE=+eip /usr/local/bin/fazt

# 6. Start service
sudo systemctl start fazt

# 7. Check status
sudo systemctl status fazt
sudo journalctl -u fazt -f
```

---

## 🔬 What Happens During Upgrade

### Database Migrations
**v0.7.2 has NO new migrations**. Your database schema is unchanged.

Existing migrations will be checked and skipped:
```
Migration 1 (initial_schema) already applied, skipping
Migration 2 (paas_tables) already applied, skipping
Migration 3 (env_vars) already applied, skipping
Migration 4 (vfs_schema) already applied, skipping
Migration 5 (site_logs) already applied, skipping
Migration 6 (config_table) already applied, skipping
Migrations completed successfully
```

### New Initialization
The server will initialize new components:
```
Analytics: Write buffer initialized
```

This is the new event buffering system that prevents database write storms.

### Startup Sequence
```
Starting fazt.sh...
Configuration loaded: Environment=production, Port=443, DB=~/.config/fazt/data.db
Migrations completed successfully
Database initialized successfully
Configuration overlaid from Database
Analytics: Write buffer initialized
Hosting initialized (VFS Mode)
Server starting on :443
Dashboard: https://your-domain.com
```

---

## ✅ Post-Upgrade Verification

### 1. Check Service Status
```bash
sudo systemctl status fazt
```

Expected output:
```
● fazt.service - Fazt PaaS
     Loaded: loaded (/etc/systemd/system/fazt.service; enabled; ...)
     Active: active (running) since ...
```

### 2. Check Logs
```bash
sudo journalctl -u fazt --since "5 minutes ago"
```

Look for:
- ✅ "Analytics: Write buffer initialized"
- ✅ "Server starting on :443" (or :80)
- ✅ No error or panic messages

### 3. Test Dashboard
```bash
curl -I https://your-domain.com
```

Should return `HTTP/2 303` (redirect to login).

### 4. Test API
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" https://your-domain.com/api/system/health
```

Should return JSON with:
```json
{
  "status": "healthy",
  "version": "0.7.2",
  "uptime_seconds": ...
}
```

---

## 🆕 New Features in v0.7.2

After upgrading, you'll have access to:

### System Observability APIs
```bash
# System health
curl https://your-domain.com/api/system/health

# Resource limits
curl https://your-domain.com/api/system/limits

# VFS cache stats
curl https://your-domain.com/api/system/cache

# Database stats
curl https://your-domain.com/api/system/db

# Server config
curl https://your-domain.com/api/system/config
```

### Site Detail APIs
```bash
# Get single site
curl https://your-domain.com/api/sites/mysite

# List files
curl https://your-domain.com/api/sites/mysite/files

# Download file
curl https://your-domain.com/api/sites/mysite/files/index.html
```

### Traffic Configuration
```bash
# Delete redirect
curl -X DELETE https://your-domain.com/api/redirects/123

# Delete webhook
curl -X DELETE https://your-domain.com/api/webhooks/456

# Update webhook
curl -X PUT https://your-domain.com/api/webhooks/456 \
  -d '{"is_active": false}'
```

---

## 🐛 Troubleshooting

### Issue: Service won't start after upgrade

**Check logs:**
```bash
sudo journalctl -u fazt -n 50
```

**Common causes:**
1. **Port already in use**: Another process is using port 80/443
   ```bash
   sudo netstat -tulpn | grep -E ':(80|443)'
   ```
2. **Capabilities not set**: Binary can't bind to privileged ports
   ```bash
   sudo setcap CAP_NET_BIND_SERVICE=+eip /usr/local/bin/fazt
   ```
3. **Database permissions**: Wrong ownership on data.db
   ```bash
   sudo chown fazt:fazt ~fazt/.config/fazt/data.db*
   ```

### Issue: Analytics buffer warnings

If you see:
```
Warning: Analytics buffer not initialized, dropping event
```

**This is normal** during startup if events arrive before buffer initializes. No action needed.

### Issue: Old version still showing

```bash
fazt --version  # Shows old version
```

**Solution**: Clear shell cache
```bash
hash -r
fazt --version  # Should show v0.7.2
```

---

## 🔙 Rollback Procedure

If you encounter issues:

### Using backup binary
```bash
# Stop service
sudo systemctl stop fazt

# Restore backup
sudo mv /usr/local/bin/fazt.old /usr/local/bin/fazt

# Start service
sudo systemctl start fazt
```

### From database backup
```bash
sudo systemctl stop fazt
sudo cp ~fazt/.config/fazt/data.db.backup ~fazt/.config/fazt/data.db
sudo chown fazt:fazt ~fazt/.config/fazt/data.db
sudo systemctl start fazt
```

---

## 📊 Performance Impact

### Expected Behavior
- **Downtime**: < 2 seconds during `systemctl restart`
- **Memory**: +5-10MB for analytics buffer
- **CPU**: No significant change
- **Database**: Fewer write operations (events are buffered)

### Resource Usage
The new analytics buffer uses:
- **RAM**: ~1MB per 1000 queued events (max 1000 events default)
- **Flush Interval**: 30 seconds (automatic)
- **Batch Size**: 1000 events per batch

System limits are auto-detected:
- **VFS Cache**: 25% of total RAM
- **Max Upload**: 10% of RAM (capped at 100MB, min 10MB)

---

## 🔐 Security Notes

### New Behavior
- **Bearer Token Auth**: Now works correctly (was broken in v0.7.1)
- **API Keys**: Existing keys continue to work
- **Session Cookies**: Unchanged

### Recommendations
After upgrading:
1. Test CLI deployments: `fazt client deploy`
2. Verify API access with Bearer tokens
3. Check session-based dashboard login still works

---

## 📞 Support

**Upgrade succeeded?**
- Check new API endpoints
- Monitor `journalctl -u fazt -f` for first 5 minutes
- Test deployments work correctly

**Upgrade failed?**
- Review troubleshooting section
- Check GitHub issues: https://github.com/fazt-sh/fazt/issues
- Include `journalctl -u fazt -n 100` output when reporting

---

## ✅ Upgrade Summary

**Tested Paths:**
- ✅ v0.7.1 → v0.7.2
- ✅ v0.7.0 → v0.7.2
- ✅ Fresh install

**Compatibility:**
- ✅ No breaking changes
- ✅ No new migrations
- ✅ Existing data unchanged
- ✅ API backward compatible
- ✅ CLI unchanged

**Recommendation**: **SAFE TO UPGRADE** using Method 1 (Built-in upgrade command)
