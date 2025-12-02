# Security Guide

## Overview

Fazt is secure by default.

## Authentication

Authentication is always enabled.

### Setup

```bash
fazt server set-credentials --username admin --password your-secure-password
```

This updates `~/.config/fazt/config.json` with a bcrypt hash.

## Session Management

- **HTTPOnly Cookies**: No JS access.
- **Secure Flag**: Enabled in production.
- **SameSite=Strict**: CSRF protection.
- **24h Expiry**: Sessions auto-refresh on activity.

## Rate Limiting

- **5 failed attempts** per IP per 15 minutes.
- **Lockout**: 15 minutes.

## Audit Logging

All login/logout events are logged to the `audit_logs` table in `data.db`.

## File Permissions

- **Config**: `0600`
- **Database**: `0600`
- **Backups**: `0700`

## Production Deployment Checklist

- [ ] Set `server.env` to `"production"`.
- [ ] Enable HTTPS (`fazt service install --https`).
- [ ] Use a strong password.
- [ ] Backup `data.db` regularly.

## Reporting Issues

Email: security@fazt.sh (or open a private advisory on GitHub).
