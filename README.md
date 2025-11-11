# Command Center (CC) v0.1.0

A unified analytics, monitoring, and tracking platform with real-time dashboard capabilities.

## ✨ Features

- **Universal Tracking Endpoint** - Auto-detects domains and tracks pageviews, clicks, and events
- **Tracking Pixel** - 1x1 transparent GIF for email/image tracking
- **Redirect Service** - URL shortening with click tracking
- **Webhook Receiver** - Accept webhook events from external services with HMAC validation
- **Real-time Dashboard** - Interactive charts, filtering, and live updates
- **PWA Support** - Installable progressive web app with offline support
- **Push Notifications** - ntfy.sh integration for traffic spikes and alerts

## 🚀 Quick Start

### Prerequisites

- Go 1.20+ (with CGO support for SQLite)
- Linux/macOS or Windows with WSL

### Installation

```bash
# Clone the repository
git clone https://github.com/jikkuatwork/command-center.git
cd command-center

# Build the server
make build

# Or build for local OS
make build-local

# Run the server
./cc-server
```

The server will start on **port 4698**. Access the dashboard at `http://localhost:4698`

### Configuration

Copy `.env.example` to `.env` and configure:

```bash
PORT=4698
DB_PATH=./cc.db
NTFY_TOPIC=your-topic      # Optional: for push notifications
NTFY_URL=https://ntfy.sh   # Optional: ntfy.sh server
ENV=development            # development or production
```

## 📊 Usage

### Tracking Website Analytics

Add the tracking script to your website:

```html
<script src="https://cc.toolbomber.com/static/js/track.min.js"></script>
```

With configuration:

```html
<script>
  window.CC_CONFIG = {
    domain: 'my-website',
    tags: ['production', 'app']
  };
</script>
<script src="https://cc.toolbomber.com/static/js/track.min.js"></script>
```

### Tracking Pixel

For email or image-based tracking:

```html
<img src="https://cc.toolbomber.com/pixel.gif?domain=newsletter&tags=email,campaign-01" style="display:none">
```

### URL Redirects with Tracking

Create redirects in the dashboard, then use:

```
https://cc.toolbomber.com/r/your-slug?tags=twitter,promo
```

### Webhooks

Configure webhooks in the dashboard, then send POST requests:

```bash
curl -X POST https://cc.toolbomber.com/webhook/your-endpoint \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Signature: your-hmac-signature" \
  -d '{"event":"deploy","status":"success"}'
```

## 🏗️ Architecture

### Backend (Go)
- **SQLite Database** with WAL mode for concurrent access
- **RESTful API** with 8 endpoints
- **Middleware**: CORS, logging, recovery
- **Real-time notifications** via ntfy.sh

### Frontend (JavaScript)
- **Tabler UI Framework** for clean, professional interface
- **Chart.js** for interactive visualizations
- **Single Page Application** with client-side routing
- **Service Worker** for PWA functionality

### Database Schema

- `events` - All tracked events (pageviews, clicks, etc.)
- `redirects` - URL shortening with click counts
- `webhooks` - Webhook endpoint configurations
- `notifications` - Push notification history

## 📱 Dashboard Pages

1. **Dashboard** - Overview with stats cards, charts, and recent events
2. **Analytics** - Detailed event filtering and search
3. **Redirects** - Create and manage tracked short URLs
4. **Webhooks** - Configure webhook endpoints
5. **Settings** - Integration snippets and preferences

## 🔧 API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/track` | POST | Track events via JSON |
| `/pixel.gif` | GET | 1x1 tracking pixel |
| `/r/{slug}` | GET | Redirect with tracking |
| `/webhook/{endpoint}` | POST | Receive webhooks |
| `/api/stats` | GET | Dashboard statistics |
| `/api/events` | GET | Paginated events (filterable) |
| `/api/domains` | GET | List domains |
| `/api/tags` | GET | List tags |
| `/api/redirects` | GET/POST | Manage redirects |
| `/api/webhooks` | GET/POST | Manage webhooks |

## 🧪 Testing

```bash
# Test tracking endpoint
./test_track.sh

# Test pixel and redirects
./test_pixel_redirect.sh

# Test webhooks
./test_webhook.sh

# Test API endpoints
./test_api.sh
```

## 📦 Deployment

### Building for Production

```bash
# Build Linux x64 binary
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -o cc-server ./cmd/server

# Create release package
make release
```

### Systemd Service

Create `/etc/systemd/system/command-center.service`:

```ini
[Unit]
Description=Command Center Analytics
After=network.target

[Service]
Type=simple
User=youruser
WorkingDirectory=/opt/command-center
ExecStart=/opt/command-center/cc-server
Restart=always
RestartSec=5
Environment="PORT=4698"
Environment="ENV=production"

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable command-center
sudo systemctl start command-center
```

### Nginx Reverse Proxy

```nginx
server {
    listen 80;
    server_name cc.toolbomber.com;

    location / {
        proxy_pass http://localhost:4698;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 🔐 Security

- **HMAC SHA256** webhook signature validation
- **Input sanitization** on all endpoints
- **Request size limits** (10KB max)
- **SQL injection protection** via prepared statements
- **CORS** configurable per environment

## 🎨 Theming

Command Center supports light and dark modes with theme persistence. Toggle via the dashboard or programmatically:

```javascript
// Set theme
document.body.setAttribute('data-bs-theme', 'dark');
localStorage.setItem('cc-theme', 'dark');
```

## 📈 Performance

- **SQLite WAL mode** for concurrent reads/writes
- **Database indexing** on frequently queried columns
- **Service Worker** caching for static assets
- **Auto-refresh** every 30 seconds for dashboard
- **Response compression** via gzip

## 🐛 Troubleshooting

### Server won't start
```bash
# Check if port is in use
lsof -i :4698

# Check logs
journalctl -u command-center -f
```

### Database errors
```bash
# Reset database (WARNING: deletes all data)
rm cc.db cc.db-shm cc.db-wal

# Rebuild
./cc-server
```

### Service worker issues
Clear browser cache and re-register:
```javascript
navigator.serviceWorker.getRegistrations().then(regs => {
  regs.forEach(reg => reg.unregister());
});
```

## 🤝 Contributing

This project was built as a v0.1 implementation during a single autonomous build session. Future improvements welcome!

## 📝 License

MIT License - see LICENSE file for details

## 🙏 Acknowledgments

- **Tabler** - UI Framework
- **Chart.js** - Data visualization
- **SQLite** - Embedded database
- **ntfy.sh** - Push notifications

---

**Built with ❤️ by Claude** | v0.1.0 | Port 4698
