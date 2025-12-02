# Command Center v0.2.0 Release

## Release Date: November 12, 2024

## Overview

Command Center v0.2.0 is a major upgrade adding authentication, security features, and a flexible JSON-based configuration system.

## What's New

### 🔐 Authentication & Security
- Username/password authentication
- Secure session management
- Rate limiting (brute-force protection)
- Audit logging
- Security headers (CSP, HSTS, etc.)

### ⚙️ Configuration System
- JSON configuration files
- Simple credential setup: `./cc-server --username admin --password pass`
- Environment-specific configs
- CLI flags for all options
- Backward compatible with v0.1.0

### 🛠️ Enhanced CLI
- `--version`, `--help`, `--verbose`, `--quiet` flags
- Beautiful startup banner
- Better error messages

### 📦 Database Improvements
- Migration system with automatic backups
- Audit logs table
- New default location: `~/.config/cc/`

### 📚 Documentation
- SECURITY.md - Security guide
- CONFIGURATION.md - Config reference
- UPGRADE.md - Migration guide

## Quick Start

### New Installation

```bash
# Download and extract
wget https://github.com/fazt-sh/fazt/releases/download/v0.2.0/fazt-v0.2.0-linux-amd64.tar.gz
tar -xzf fazt-v0.2.0-linux-amd64.tar.gz
# Binary is now named 'fazt-linux-amd64' inside the tar, usually renamed to 'fazt'
mv fazt-linux-amd64 fazt
chmod +x fazt

# Setup authentication
./fazt server init --username admin --password your-secure-password --domain https://example.com

# Start server
./fazt server start
```

### Upgrading from v0.1.0

```bash
# Backup your database
cp cc.db cc.db.backup

# Download v0.2.0
wget https://github.com/fazt-sh/fazt/releases/download/v0.2.0/fazt-v0.2.0-linux-amd64.tar.gz
tar -xzf fazt-v0.2.0-linux-amd64.tar.gz

# Start server
./fazt server start
```

See [UPGRADE.md](UPGRADE.md) for detailed migration guide.

## Downloads

| Platform | Architecture | Download |
|----------|--------------|----------|
| Linux | x64 | [fazt-v0.2.0-linux-amd64.tar.gz](https://github.com/fazt-sh/fazt/releases/download/v0.2.0/fazt-v0.2.0-linux-amd64.tar.gz) |

### Checksums

```
# SHA256 checksums will be added to GitHub release
```

## Breaking Changes

⚠️ **Configuration Method**
- New default config location: `~/.config/fazt/config.json`
- New default database location: `~/.config/fazt/data.db`
- Environment variables still work but are deprecated

⚠️ **Dashboard Access**
- Dashboard now protected by authentication (when enabled)
- Tracking endpoints remain public

## Upgrade Path

v0.1.0 users can upgrade smoothly:
1. Backup database
2. Download v0.2.0
3. Set up authentication (optional)
4. Start server

See [UPGRADE.md](UPGRADE.md) for details.

## Documentation

- [README.md](README.md) - Overview and quick start
- [SECURITY.md](SECURITY.md) - Security features and best practices
- [CONFIGURATION.md](CONFIGURATION.md) - Complete configuration reference
- [UPGRADE.md](UPGRADE.md) - Migration guide from v0.1.0
- [CHANGELOG.md](CHANGELOG.md) - Detailed changelog

## Support

- Issues: https://github.com/fazt-sh/fazt/issues
- Documentation: https://github.com/fazt-sh/fazt

## Contributors

Built by [fazt-sh](https://github.com/fazt-sh) with autonomous AI assistance.

## License

See [LICENSE](LICENSE) for details.

---

**Full Changelog**: https://github.com/fazt-sh/fazt/compare/v0.1.0...v0.2.0
