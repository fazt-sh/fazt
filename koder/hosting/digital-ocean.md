# Hosting Fazt on DigitalOcean 🚀

This guide details setting up a **Fazt** instance on DigitalOcean.
It targets the **Basic Droplet** tier, which offers ample performance
for personal PaaS usage.

## 1. Droplet Configuration 💧

*   **Image**: **Ubuntu 24.04 (LTS) x64** (Recommended for long-term stability).
*   **Plan**: **Basic (Shared CPU)**.
*   **Size**: **$6/mo** (1GB RAM / 1 CPU / 25GB SSD).
    *   *Note*: The $4/mo (512MB) plan is possible but risks OOM kills during
        heavy compilation or traffic spikes. 1GB is the safe baseline.
*   **Region**: Select a datacenter geographically close to you or your users.
*   **Networking**: Enable **IPv6** (Free & Future-proof).
*   **Authentication**: Add your SSH public key.

## 2. DNS Configuration 🌐

Configure your domain registrar (e.g., Namecheap, Cloudflare) to point to your
new Droplet IP.

*   **A Record**: `@` → `YOUR_DROPLET_IP`
*   **A Record**: `*` → `YOUR_DROPLET_IP` (Required for wildcard subdomains).
*   *(Optional)* **AAAA Record**: `@` → `YOUR_DROPLET_IPV6`

*Allow propagation time before proceeding.*

## 3. Deployment 🚀

**Fazt** is a single-binary application. Deployment involves building the binary
locally and shipping it to the server.

### 3.1 Build Locally
Compile the binary for Linux AMD64:
```bash
GOOS=linux GOARCH=amd64 go build -o fazt ./cmd/server
```

### 3.2 Upload
Transfer the binary to your server via SCP:
```bash
scp fazt root@YOUR_DROPLET_IP:/root/
```

### 3.3 Install Service
SSH into your server and run the auto-provisioning command. This handles:
1.  Creating a dedicated `fazt` system user.
2.  Setting capabilities (bind ports 80/443 without root).
3.  Generating a systemd unit file.
4.  Initializing the database and configuration.

```bash
ssh root@YOUR_DROPLET_IP

# Run the installer (Interactive)
./fazt service install \
  --domain https://example.com \
  --email admin@example.com \
  --https
```

*   **Flags**:
    *   `--domain`: Your main domain (e.g., `https://paas.net`).
    *   `--email`: Required for Let's Encrypt notifications.
    *   `--https`: Enables automatic SSL provisioning.

## 4. Firewall Hardening (UFW) 🛡️

Secure your server by restricting incoming traffic to essential ports.

```bash
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow http
ufw allow https
ufw enable
```

## 5. Verification ✅

1.  **Dashboard**: Visit `https://example.com`. You should see the login screen.
    *   Default User: `admin`
    *   Password: *Printed during the install step.*
2.  **Service Status**: Check the daemon on the server:
    ```bash
    fazt service status
    # or
    systemctl status fazt
    ```
3.  **Logs**: Monitor real-time logs:
    ```bash
    fazt service logs
    ```

## 6. Maintenance 🔧

### Updating
To update Fazt, simply upload the new binary and restart the service:

```bash
# Local
GOOS=linux GOARCH=amd64 go build -o fazt ./cmd/server
scp fazt root@YOUR_DROPLET_IP:/usr/local/bin/fazt

# Remote
ssh root@YOUR_DROPLET_IP "systemctl restart fazt"
```

### Backups
All state is contained in a single SQLite file. Backup is trivial:

```bash
# Download a snapshot of the database
scp root@YOUR_DROPLET_IP:/home/fazt/.config/fazt/data.db ./backup.db
```