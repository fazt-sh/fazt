# DigitalOcean Setup Guide

**Recommended Droplet**:
- Image: Ubuntu 24.04 (LTS) or Debian 12
- Size: Basic (Running cleanly on 512MB RAM / 1 CPU)

## 1. Quick Install (Recommended)

SSH into your fresh droplet and run:

```bash
# 1. Download & Install
curl -sL https://raw.githubusercontent.com/fazt-sh/fazt/master/install.sh | bash

# 2. Setup Service (Auto-configures Systemd, HTTPS & Firewall)
sudo ./fazt service install --domain your-domain.com --email admin@example.com --https
```

That's it!

## 2. Troubleshooting

**"Port 80 already in use"**
If you have Nginx or Apache pre-installed:
```bash
systemctl stop nginx
systemctl disable nginx
# Then run the service install command again
```

**Logs**
```bash
journalctl -u fazt -f
```

**Status**
```bash
systemctl status fazt
```
