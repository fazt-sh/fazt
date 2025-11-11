# Changelog

All notable changes to Command Center will be documented in this file.

## [0.1.0] - 2025-11-11

### Added
- Initial release of Command Center
- Universal tracking endpoint with domain auto-detection
- 1x1 transparent GIF pixel tracking
- URL redirect service with click tracking
- Webhook receiver with HMAC SHA256 validation
- Real-time dashboard with interactive charts
- Analytics page with filtering (domain, source, search)
- Redirects management interface
- Webhooks configuration interface
- Settings page with integration snippets
- PWA support with service worker
- Client-side tracking script (track.min.js)
- Light/dark theme toggle with persistence
- SQLite database with WAL mode
- ntfy.sh integration for notifications
- RESTful API with 8 endpoints
- Comprehensive test scripts
- Production-ready deployment configuration

### Features
- **Backend**: Go + SQLite with proper indexing
- **Frontend**: Tabler UI with Chart.js visualizations
- **Database**: 4 tables (events, redirects, webhooks, notifications)
- **API**: Complete CRUD operations for all resources
- **Security**: HMAC validation, input sanitization, prepared statements
- **Performance**: Database indexing, service worker caching, auto-refresh

### Documentation
- Complete README with installation instructions
- API endpoint documentation
- Deployment guide (systemd, nginx)
- Usage examples for all tracking methods
- Troubleshooting section

### Testing
- 4 comprehensive test scripts
- All endpoints tested and validated
- Mock data generator for development

---

**Total Commits**: 13
**Lines of Code**: ~5000+
**Build Time**: ~8 hours (autonomous session)
