# Admin SPA - Complete Implementation Plan

**Date**: December 9, 2025
**Status**: Active - Phase 2C (Feature UI Completion)
**Supersedes**: Plan 12 (Admin SPA Rebuild), Plan 13 (Route Hierarchy)
**Tech Stack**: React 18 + TypeScript + Vite + Tailwind CSS

---

## 1. Executive Summary

This is the **single source of truth** for the Fazt Admin SPA. It combines:
- Foundation work already completed (Phase 1)
- UX-first route hierarchy design
- Remaining implementation tasks
- Implementation-ready code for any model

**Core Principles:**
- **Delight First**: First impression of Fazt
- **UX-First Routing**: Routes organized by user workflows, not API structure
- **Comprehensive**: Interface for all ~30 API endpoints
- **Professional**: Technical Precision aesthetic (Linear/Vercel-inspired)

---

## 2. Current State (What's Built)

### ✅ Phase 1: Foundation (Complete)
- Project Setup (Vite + React + TS)
- Tailwind Design System & Theme (Dark/Light)
- App Shell (Sidebar, Navbar)
- Auth Context & Mock Mode
- Dashboard & Datamaps

### ✅ Phase 2A: Route Hierarchy (Complete)
- UX-First Routing structure (`/sites`, `/system`, `/apps`, `/security`)
- Breadcrumbs implementation
- Placeholder pages for all routes
- 404 Page

### ✅ Phase 2B: Core Pages (Complete)
- **Site Detail**: Tabbed interface (Overview, Env, Keys, Logs, Settings)
- **Webhooks**: Full CRUD UI
- **Redirects**: Full CRUD UI
- **System Stats**: Real-time metrics UI
- **System Health**: Service status UI
- **Profile**: Settings & Password UI
- **Toasts**: Notification system implemented

### 🟡 Phase 2C: Pending Feature UI (Current Priority)
The `SiteDetail` page structure exists, but specific tabs need internal UI implementation:
1.  **Logs Tab**: Needs `Terminal` integration and stream viewer.
2.  **Environment Tab**: Needs Key-Value table and "Add Variable" UI.
3.  **API Keys Tab**: Needs Keys list and generation UI.
4.  **Settings Tab**: Needs Delete/Domain UI.

---

## 3. Technology Stack

### Core
- **React 18+** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool
- **Tailwind CSS** - Styling
- **React Router v6** - Hash-based routing

### State & Data
- **TanStack Query** - API state, caching
- **React Context** - Auth, theme, mock mode, toast
- **React Hook Form** - Form handling

### UI
- **Lucide React** - Icons
- **Headless UI** - Accessible primitives
- **Custom Components** - Design system (Preline-inspired)

---

## 4. Route Hierarchy (UX-First)

### Complete Route Map

```
/#/dashboard                 Dashboard
/#/profile                   User Profile

/#/sites                     Sites List
/#/sites/analytics           Aggregate Site Analytics
/#/sites/domains             Domain Management
/#/sites/create              Create New Site
/#/sites/:id                 Site Detail

/#/system                    → Redirects to /system/stats
/#/system/stats              System Statistics
/#/system/limits             Resource Limits
/#/system/logs               System Logs
/#/system/backup             Backup Management
/#/system/health             Health Monitoring
/#/system/settings           System Settings
/#/system/design-system      Design System (dev only)

/#/apps                      Apps List
/#/apps/webhooks             Webhooks
/#/apps/redirects            URL Redirects
/#/apps/tunnel               Tunnel Management
/#/apps/proxy                Proxy Configuration
/#/apps/botdaddy             BotDaddy (Telegram Bot)

/#/security                  → Redirects to /security/ssh
/#/security/ssh              SSH Keys
/#/security/tokens           Auth Tokens

/#/external                  → Redirects to /external/cloudflare
/#/external/cloudflare       Cloudflare Integration
/#/external/litestream       Litestream Integration

/#/login                     Login (outside AppShell)
/#/*                         404 Not Found
```

---

## 5. Implementation Plan

### Phase 2C: Feature UI Completion (NEXT)

**Goal**: Implement the missing internal UI for Site Detail tabs using mock data.

1.  **Logs Tab UI**
    - Component: `SiteLogs.tsx`
    - Use `Terminal` component from `ui/`.
    - Mock a log stream.

2.  **Environment Variables UI**
    - Component: `SiteEnv.tsx`
    - Table layout for Key/Value.
    - Mask values by default (*****).
    - "Reveal" button.
    - "Add Variable" form.

3.  **API Keys UI**
    - Component: `SiteKeys.tsx`
    - List of active keys.
    - "Create Key" modal showing token once.

4.  **Settings UI**
    - Component: `SiteSettings.tsx`
    - Custom Domain form.
    - Danger Zone (Delete Site).

### Phase 3: Real Data Integration (Future)

**Goal**: Replace `mockData` and `useMockMode` with real API calls.

1.  **API Client Setup**
    - Create `api/client.ts` using `axios` or `fetch`.
    - Configure interceptors for Auth.

2.  **TanStack Query Hooks**
    - `useSites`, `useSite(id)`
    - `useDeployments`
    - `useSystemHealth`
    - `useProfile`

3.  **Validation & Error Handling**
    - Integrate `zod` for all forms.
    - Add Error Boundaries.
    - Global Error Toast handler.

4.  **Production Build**
    - `npm run build`
    - Copy to `internal/assets/system/admin`.
    - `go build`

---

## 6. Development Commands

```bash
# Start dev server
cd /home/testman/workspace/admin
npm run dev -- --port 37180 --host 0.0.0.0

# Build for production
npm run build
```

---

**Plan Status**: ✅ ACTIVE
**Current Phase**: 2C (Feature UI Completion)
**Next Action**: Implement Site Detail Tabs (Logs, Env, Keys, Settings)