# [Fazt ⚡](https://fazt-sh.github.io/fazt/)

**The "Cartridge" Personal Cloud Platform.**
One Binary. One Database. Zero Dependencies.

## 🚀 Install

Run this on your VPS (Ubuntu/Debian/Arch):

```bash
curl -s https://fazt-sh.github.io/fazt/install.sh |
  sudo bash
```

**Production Setup (Auto HTTPS):**

```bash
sudo fazt service install \
  --domain example.com \
  --email you@mail.com \
  --https
```

## 🧠 Philosophy

Fazt follows the **Cartridge Architecture**:

- **State is Precious**:
  All data lives in one SQLite file (`data.db`).
  - Sites, Analytics, Users, SSL Certs.
  - Backup one file = Backup everything.

- **Binary is Disposable**:
  The `fazt` executable is stateless.
  - Contains runtime, migrations, UI.
  - Upgrade? Just replace the binary.

- **Zero Dependencies**:
  No Nginx. No Docker. No Node.js.
  - Native Let's Encrypt (HTTPS).
  - Built-in JS Serverless Runtime.
  - Virtual Filesystem (VFS).

## ✨ Features

- **PaaS**: Deploy static sites & JS functions.
- **Analytics**: Built-in privacy-first tracking.
- **Routing**: Auto subdomains (`blog.domain.com`).
- **Dashboard**: Real-time metrics & management.

## 🛠️ CLI

- `fazt deploy`
  Push your local site to the cloud.

- `fazt client set-auth-token`
  Authenticate your local machine.

- `fazt server set-credentials`
  Update admin password (useful if forgotten).

- `fazt server reset-admin`
  Reset Admin Dashboard UI (after upgrade).

- `fazt service logs`
  Tail system logs.

## 🏛️ Architecture Details

*   **Virtual Filesystem (VFS)**: All hosted sites live in the `data.db`.
*   **System Sites**: The Admin Dashboard (`admin.<domain>`), Landing Page (`root`), and 404 Page are standard sites in the VFS. They are seeded from the binary on startup.
*   **Routing**:
    *   `admin.*`: Admin Dashboard
    *   `root.*` / `(domain)`: Landing Page
    *   `*`: User Sites

## License
MIT