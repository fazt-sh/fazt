# Admin UI Parity Plan

**Goal**: Take admin from 15% to 80% completeness
**Status**: In Progress
**Created**: 2026-02-02

## Current State Analysis

### Backend Coverage

**Total handlers**: 69 handlers across 22 files

| Handler File | Handlers | Purpose |
|--------------|----------|---------|
| agent_handler.go | 8 | Agent info, storage, snapshots, logs, errors |
| aliases_handler.go | 8 | List, detail, create, update, delete, reserve, swap, split |
| api.go | 6 | Stats, events, domains, tags, redirects, webhooks |
| apps_handler.go | 8 | List, detail, delete, files, source, file content, install, create |
| apps_handler_v2.go | 8 | V2 CRUD, fork, lineage, forks |
| auth_handlers.go | 4 | User me, login, logout, auth status |
| system.go | 6 | Health, limits, cache, db, config, capacity |
| sql.go | 1 | SQL query execution |
| logs.go + logs_stream.go | 2 | Log viewing and streaming |
| upgrade_handler.go | 1 | Binary upgrade |
| deploy.go | 1 | Deployment |
| Others | 16 | Webhooks, redirects, tracking, etc. |

### SDK Coverage

**Current methods**: ~27 methods

| Namespace | Methods | Coverage |
|-----------|---------|----------|
| auth | session, signOut | ✅ Complete |
| apps | list, get, files, source, file, delete, create, install | ✅ Core complete |
| aliases | list, get, create, update, delete, reserve, swap, split | ✅ Complete |
| system | health, config, limits, cache, db, capacity | ✅ Complete |
| stats | overview, app | ✅ Complete |
| templates | list | ✅ Complete |

**Missing SDK methods** (needed for parity):
- Events: `events.list()`
- Logs: `logs.list()`, `logs.stream()`
- SQL: `sql.query()`
- App V2: `apps.fork()`, `apps.lineage()`, `apps.forks()`, `apps.update()`
- Agent: `agent.snapshot()`, `agent.restore()`, `agent.snapshots()`

### Page Coverage

**Current pages** (1,767 lines total):

| Page | Lines | Status | Functionality |
|------|-------|--------|---------------|
| dashboard.js | 362 | ✅ Working | Stats, recent apps, activity |
| apps.js | 679 | ✅ Working | List + detail view |
| aliases.js | 170 | ✅ Working | List view only |
| system.js | 174 | ✅ Working | Health, memory, DB stats |
| settings.js | 152 | ⚠️ Partial | Theme/palette only, no server settings |
| design-system.js | 230 | 🔧 Reference | CSS/component showcase |

**Missing pages**:
1. ❌ Logs page
2. ❌ Events/activity page (dedicated view)
3. ❌ SQL browser
4. ❌ Backups/snapshots page
5. ❌ Create app modal
6. ❌ Edit app modal
7. ❌ Create/edit alias modal
8. ❌ Server config settings

### Mock Fixtures

**Current fixtures**:
- apps.json
- aliases.json
- health.json
- user.json

**Missing fixtures**:
- events.json
- logs.json
- config.json (server config)
- snapshots.json

---

## Parity Roadmap

### Phase 1: Core CRUD Completion (40% → 55%)

**Task 1.1: Add Create/Edit Modals**
- [ ] Create App modal (from template or blank)
- [ ] Create Alias modal
- [ ] Edit Alias modal
- [ ] Delete confirmation modals (consistent UX)

**Task 1.2: Enhance Existing Pages**
- [ ] Apps: Add "New App" button → modal
- [ ] Apps detail: Add "Edit" button → modal
- [ ] Aliases: Add "New Alias" button → modal
- [ ] Aliases: Add edit/delete actions

### Phase 2: Activity & Observability (55% → 70%)

**Task 2.1: Events/Activity Page**
- [ ] Add SDK: `events.list(options)`
- [ ] Add fixture: events.json
- [ ] Create page: events.js
- [ ] Add route: /events
- [ ] Features: Filter by type, app, date range

**Task 2.2: Logs Page**
- [ ] Add SDK: `logs.list(options)`, `logs.stream()`
- [ ] Add fixture: logs.json
- [ ] Create page: logs.js
- [ ] Add route: /logs
- [ ] Features: Filter by app, level, search text

**Task 2.3: Enhanced Dashboard**
- [ ] Recent events widget (last 10 events)
- [ ] System health sparkline
- [ ] Quick actions panel (deploy, create app)

### Phase 3: Administration (70% → 80%)

**Task 3.1: Server Settings Page**
- [ ] Add SDK: `system.updateConfig()`
- [ ] Expand settings.js with "Server" tab
- [ ] Config options: Domain, HTTPS, notifications
- [ ] Display-only for sensitive config

**Task 3.2: Backups Page**
- [ ] Add SDK: `agent.snapshot()`, `agent.restore()`, `agent.snapshots()`
- [ ] Create page: backups.js
- [ ] Add route: /backups
- [ ] Features: Create backup, restore, download

**Task 3.3: SQL Browser (Stretch)**
- [ ] Add SDK: `sql.query()`
- [ ] Create page: sql.js
- [ ] Features: Query input, results table, export

---

## Implementation Order (Prioritized)

### Week 1: Core Completion

**Day 1-2: Modals & CRUD**
1. Create App modal
2. Create Alias modal
3. Edit Alias modal

**Day 3-4: Events Page**
4. SDK + fixtures for events
5. Events page implementation
6. Dashboard events widget

**Day 5: Polish**
7. Error states
8. Loading states
9. Empty states

### Week 2: Observability

**Day 1-2: Logs**
10. SDK + fixtures for logs
11. Logs page implementation

**Day 3-4: Server Settings**
12. Server config in settings
13. Display system info

**Day 5: Integration Testing**
14. Test all flows in real mode
15. Fix any API mismatches

---

## Files to Create/Modify

### New Files
```
admin/src/pages/events.js        # Events page
admin/src/pages/logs.js          # Logs page
admin/src/pages/backups.js       # Backups page (stretch)
admin/packages/fazt-sdk/fixtures/events.json
admin/packages/fazt-sdk/fixtures/logs.json
admin/packages/fazt-sdk/fixtures/config.json
```

### Modified Files
```
admin/packages/fazt-sdk/index.js  # Add events, logs, agent SDK methods
admin/packages/fazt-sdk/mock.js   # Add mock routes
admin/src/routes.js               # Add new routes
admin/src/stores/data.js          # Add events, logs stores
admin/src/pages/apps.js           # Add create modal
admin/src/pages/aliases.js        # Add create/edit modals
admin/src/pages/settings.js       # Add server config tab
admin/src/pages/dashboard.js      # Add events widget
admin/index.html                  # Add modal HTML, nav items
```

---

## Success Criteria

**80% Parity means:**
- [x] View all apps with filtering
- [x] View app details with files
- [x] Delete apps
- [ ] Create apps
- [x] View all aliases
- [ ] Create/edit/delete aliases (full CRUD)
- [x] View system health
- [ ] View server configuration
- [ ] View activity/events
- [ ] View logs
- [x] Theme/palette preferences
- [ ] Server settings management

**Nice to have (90%+):**
- [ ] SQL browser
- [ ] Backups management
- [ ] Real-time log streaming
- [ ] Deploy from UI

---

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Backend API missing | Check handlers before implementation |
| Mock/real drift | Test both modes for each feature |
| Scope creep | Stick to 80% target, defer extras |
| UI consistency | Use design system patterns |

---

## Progress Tracking

| Task | Status | Notes |
|------|--------|-------|
| Analysis complete | ✅ | This document |
| Create App modal | ⏳ | Next |
| Create Alias modal | ⏳ | |
| Events page | ⏳ | |
| Logs page | ⏳ | |
| Server settings | ⏳ | |
| Integration testing | ⏳ | |

---

## References

- [Admin UI Architecture](../knowledge-base/workflows/admin-ui/architecture.md)
- [Adding Features Workflow](../knowledge-base/workflows/admin-ui/adding-features.md)
- [Testing Strategy](../knowledge-base/workflows/admin-ui/testing.md)
- [Feature Checklist](../knowledge-base/workflows/admin-ui/checklist.md)
