# Future Roadmap & Random Ideas 💡

This document serves as the backlog for future features. Ideally, implement one item at a time.

## 1. Core Platform (Maintenance & Ops) 🛠️
1.  **`fazt upgrade`**: Auto-update mechanism.
    *   Fetch latest binary from GitHub Releases.
    *   **Critical**: Re-apply `setcap CAP_NET_BIND_SERVICE=+eip`.
    *   Restart systemd service.
2.  **`fazt backup`**: Database snapshotting.
    *   `fazt backup create`: Dump SQLite to a timestamped file.
    *   `fazt backup restore`: Safety checks + overwrite `data.db`.
    *   *Stretch*: S3 integration via Litestream.
3.  **Email Service**:
    *   **Inbound**: Receive emails for `admin@your-domain.com` (SMTP port 25 or Webhook).
    *   **Storage**: `inbox` table in SQLite.
    *   **Outbound**: Relay via Postmark/AWS SES.

## 2. "Personal Cloud" Apps (The Fun Stuff) ☁️
4.  **Ephemeral File Sharing**:
    *   "Pastebin/Snapdrop" for files.
    *   Upload -> Get Link -> Auto-delete after 24h.
    *   UI: Drag & Drop zone.
5.  **Live Scratchpad**:
    *   Real-time synced textarea (WebSocket).
    *   Use case: Copy text on phone, paste on desktop instantly.
6.  **Link Redirector**:
    *   Bit.ly clone (Short URLs).
    *   Analytics: Click counts, referrers (already partially supported).
7.  **TextDB**:
    *   Simple JSON document store over HTTP.
    *   `POST /db/collection` -> Save JSON.
    *   `GET /db/collection` -> List items.
8.  **WebDAV Server**:
    *   Mount `fazt` as a network drive on Windows/Mac.
    *   Back up photos/docs directly to your VPS.

## 3. Advanced Protocols & Integrations 🔌
9.  **Joplin Server**:
    *   Implement the sync API for Joplin Notes app.
10. **STUN/TURN Server**:
    *   Run a relay for P2P WebRTC (Video calling).
11. **PubSub Hub**:
    *   HTTP -> WebSocket message broker.
    *   IoT integration (sensors post to HTTP, dashboard listens via WS).

## 4. Developer Experience (Serverless) 💻
12. **Runtime V2 (JS)**:
    *   Add `fetch()` support (HTTP requests from JS).
    *   Add `db.get/set` (KV Store access).
    *   Add CRON jobs (`fazt.json` schedule).

## 5. For every new site launched, get the fazt dashboard these subdomains:
- dashboard: fazt.example.com
- lander example.com (we will use root.example.com file in homepage if uploading
  to homepage domain is hard for client?)
- 404 in 404.example.com // the same page will be shown in any unavailable
  sites; if this does not exist, we will show generic fazt 404

## 6. fazt locations

url: fazt.sh
github: github.com/fazt-sh
twitter: x.com/fazt_sh

Can we build a single install command soemthing like? curl -s fazt.sh/install.sh
| sh

## 7. Can't we avoid creating config directory and store that too to the db?

## 8. can we do a simple in memory cache for the sqlite file? Specifically the
sites that were just used?
