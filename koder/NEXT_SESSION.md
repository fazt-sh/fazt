# Current Task

**Project**: Admin SPA
**Location**: `/home/testman/workspace/admin`
**Plan**: `koder/plans/14_admin-spa-complete.md`
**Phase**: 2C - Feature UI Completion (Bridge to Phase 3)

## Status Update (Dec 9, 2025)

We have successfully completed the core shell and main CRUD pages. The app is fully navigable with a polished look and feel.

### ✅ Completed
- **Navigation**: Sidebar (collapsible), Navbar, Breadcrumbs, UX-first routing.
- **Global Systems**: Theme (Dark/Light), Toast Notifications, Mock Mode.
- **Pages (Full UI)**:
  - Dashboard (Stats, Datamap, Traffic Chart).
  - Sites List (Grid view).
  - Webhooks (List, Create, Edit, Delete).
  - Redirects (List, Create, Edit, Delete).
  - Profile (Tabs, Forms).
  - System Stats & Health.

### 🚧 Pending UI (Must complete before Data Integration)
The `SiteDetail` page has placeholder text for complex tabs. These need full React components (mocked first if necessary to verify UX).

1.  **Site Logs UI** (`/sites/:id` -> Logs Tab)
    - Needs to integrate the existing `Terminal` component.
    - Needs a log stream viewer interface.
2.  **Environment Variables UI** (`/sites/:id` -> Environment Tab)
    - Needs a key-value table/grid.
    - Needs "Add Variable" modal/inline row.
    - Needs "Reveal/Hide" secret toggle.
3.  **API Keys UI** (`/sites/:id` -> API Keys Tab)
    - Needs a list of keys.
    - Needs "Create Key" flow (showing key only once).
4.  **Site Settings UI** (`/sites/:id` -> Settings Tab)
    - Needs "Delete Site" danger zone.
    - Needs "Custom Domain" configuration form.

## Next Steps (Immediate)
1.  **Finish SiteDetail UI**: Implement the 4 missing tabs listed above using mock data.
2.  **Verify 100% UI**: Ensure every pixel of the interface is clickable and looks "production-ready" with mock data.

## Next Steps (After UI Complete)
1.  **Phase 3**: Real Data Integration.
    - Replace `mockData.ts` with TanStack Query hooks hitting the Go API.
    - Add Zod validation.
    - Add Error Boundaries.

## Dev Server
```bash
cd admin && npm run dev -- --port 37180 --host 0.0.0.0
```
(Currently running in background)
