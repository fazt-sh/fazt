# Plan: Auto-Provisioning (Server Install)

**Status:** Completed ✅
**Target Version:** v0.6.0
**Goal:** Make `fazt` capable of installing itself as a system service on a fresh Linux server.

## Overview
The `fazt service install` command automates the setup of the application on a production server (specifically targeting Ubuntu/Debian/Systemd based systems). It reduced the deployment manual from ~10 steps to 1 step.

## Implemented Features

1.  **`internal/provision` Package**:
    *   **User Management**: Checks/Creates `fazt` system user.
    *   **Systemd**: Generates `/etc/systemd/system/fazt.service` from a template.
    *   **Permissions**: Uses `setcap` to allow port 80/443 binding without root.
    *   **Config Gen**: Creates `config.json` with domain/HTTPS settings.

2.  **CLI Command**:
    *   `fazt service install --domain ... --https`
    *   Generates secure admin credentials if not provided.

3.  **Embeddable Architecture**:
    *   **Migrations**: SQL files embedded via `embed.FS`.
    *   **Templates**: HTML/CSS embedded via `embed.FS`.
    *   **Driver**: Switched to `modernc.org/sqlite` (Pure Go) for static binaries.

## Usage

```bash
# On a fresh VPS
./fazt service install \
  --domain https://example.com \
  --email admin@example.com \
  --https
```

## Security Considerations
-   **Least Privilege**: Service runs as `fazt`, not `root`.
-   **File Permissions**: Config/DB locked to `fazt` user (0600).
-   **Capabilities**: `CAP_NET_BIND_SERVICE` allows low port binding.

## Retrospective
-   **Challenge**: `setcap` attributes are lost when replacing binary via `scp`.
-   **Solution**: Future `fazt upgrade` command will handle this. For now, manual `setcap` is required on update.
-   **Success**: Successfully deployed to DigitalOcean $6 droplet with automatic HTTPS.
