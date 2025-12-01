# fazt.sh

A unified analytics, monitoring, tracking platform, and **Personal Cloud** with static hosting and serverless JavaScript functions.

**A completely self-contained "Cartridge" Application.**

## Features

### Personal Cloud (PaaS)
- **Single Binary & Single DB** - The entire platform runs from `fazt` executable and `data.db`.
- **Zero Dependencies** - No Nginx required. Native automatic HTTPS via Let's Encrypt (CertMagic).
- **Virtual Filesystem (VFS)** - Sites and assets are stored in the SQLite database.
- **Static Site Hosting** - Deploy static websites via CLI.
- **Serverless JavaScript** - Run JavaScript functions with `main.js`.
- **WebSocket Support** - Real-time communication.

### Analytics & Tracking
- **Universal Tracking Endpoint** - Auto-detects domains and tracks pageviews/events.
- **Real-time Dashboard** - Interactive charts and live updates.

## Quick Start (Production)

Deploying `fazt` to a Linux server (Ubuntu/Debian) is a single command.

### 1. Build
```bash
# Build a static binary (works on any Linux distro)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o fazt ./cmd/server
```

### 2. Install
Upload the binary to your server and run the installer. This will:
*   Create a system user.
*   Setup a systemd service.
*   Provision automatic HTTPS.

```bash
# On your server (as root or sudo)
./fazt service install \
  --domain https://your-domain.com \
  --email admin@your-example.com \
  --https
```

### 3. Deploy
From your local machine:

```bash
# 1. Login to your new dashboard to get an API Token.

# 2. Configure local client
./fazt client set-auth-token --token <YOUR_TOKEN>

# 3. Deploy a site
./fazt deploy --path ./my-website --domain blog --server https://your-domain.com
```

Your site is now live at `https://blog.your-domain.com`!

## CLI Reference

### Service Management (Production)
Commands for managing the background daemon.
*   `fazt service install`: Install systemd service & user.
*   `fazt service start`: Start the daemon.
*   `fazt service stop`: Stop the daemon.
*   `fazt service status`: Check daemon health.
*   `fazt service logs`: Tail system logs.

### Server Management (Manual/Dev)
Commands for running the process directly.
*   `fazt server start`: Run in foreground.
*   `fazt server init`: Generate config file.
*   `fazt server status`: Check app internal state.

### Client
*   `fazt deploy`: Deploy a directory.
*   `fazt client set-auth-token`: Save API credentials.

## "Cartridge" Architecture

**fazt** follows a "Cartridge" architecture:
- **State**: All state (Users, Analytics, Sites, Files, SSL Certs) lives in a single SQLite file (`data.db`).
- **Stateless Binary**: The `fazt` binary contains all logic, migrations, and UI templates. Updating is as simple as replacing the binary.
- **Backup/Restore**: Just copy `data.db`.

## License
MIT License
