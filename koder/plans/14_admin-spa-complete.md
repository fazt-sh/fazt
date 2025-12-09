# Admin SPA - Complete Implementation Plan

**Date**: December 9, 2025
**Status**: Active - Ready for Implementation
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

### ✅ Phase 1 Complete

| Component | Status | Notes |
|-----------|--------|-------|
| Project Setup | ✅ Done | Vite + React + TypeScript |
| Tailwind CSS | ✅ Done | Custom design system |
| Folder Structure | ✅ Done | Components, pages, hooks, lib |
| UI Primitives | ✅ Done | Button, Input, Card, Badge, etc. |
| App Shell | ✅ Done | Navbar, Sidebar (with sub-menus) |
| Theme System | ✅ Done | Light/dark with CSS variables |
| Auth Context | ✅ Done | Login, logout, session |
| Mock Context | ✅ Done | URL param + localStorage |
| TanStack Query | ✅ Done | API caching configured |
| Router | ✅ Done | Hash-based routing |
| PWA | ✅ Done | Manifest, service worker |
| Login Page | ✅ Done | Functional |
| Dashboard | ✅ Done | Stats, world map, traffic |
| Sites Page | ✅ Done | List view |
| Design System | ✅ Done | Component showcase |
| Datamaps | ✅ Done | World map with theme support |

### 🟡 Needs Work

| Component | Status | Notes |
|-----------|--------|-------|
| Route Hierarchy | 🟡 Partial | Flat structure, needs UX-first redesign |
| Sidebar Menu | 🟡 Partial | Has sub-menus, needs new structure |
| Breadcrumbs | ❌ Missing | Need to implement |
| 404 Page | ❌ Missing | Need to implement |
| Placeholder Pages | ❌ Missing | ~20 pages need placeholders |
| Site Detail | ❌ Missing | Individual site management |
| System Pages | ❌ Missing | Stats, limits, logs, backup, health |
| Apps Pages | ❌ Missing | Webhooks, redirects, tunnel, etc. |
| Security Pages | ❌ Missing | SSH, tokens, password |
| External Pages | ❌ Missing | Cloudflare, Litestream |

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
- **React Context** - Auth, theme, mock mode
- **React Hook Form** - Form handling (when needed)

### UI
- **Lucide React** - Icons
- **Headless UI** - Accessible primitives
- **Custom Components** - Design system

### PWA
- **vite-plugin-pwa** - Service worker
- **Workbox** - Caching

---

## 4. Design System (Current)

### 4.1 Aesthetic
**"Technical Precision"** - Linear/Vercel inspired
- Clean, minimal interfaces
- Subtle borders and shadows
- Purposeful whitespace
- Monospace accents for data

### 4.2 Colors (CSS Variables)

```css
/* Light Mode */
--bg-primary: #ffffff;
--bg-secondary: #f9fafb;
--bg-surface: #ffffff;
--text-primary: #111827;
--text-secondary: #6b7280;
--text-tertiary: #9ca3af;
--border-default: #e5e7eb;
--accent: #f97316;  /* Orange */

/* Dark Mode */
--bg-primary: #0a0a0a;
--bg-secondary: #111111;
--bg-surface: #1a1a1a;
--text-primary: #fafafa;
--text-secondary: #a1a1aa;
--text-tertiary: #71717a;
--border-default: #27272a;
--accent: #f97316;
```

### 4.3 Typography
- **Primary**: Inter
- **Monospace**: Consolas, Monaco, monospace
- **Scale**: text-xs through text-4xl (Tailwind default)

### 4.4 Components Built
- Button (primary, secondary, ghost, danger)
- Input (text, password, with icons)
- Card (default, bordered)
- Badge (success, error, warning, info)
- Skeleton (text, rect, circle)
- Modal (sm, md, lg)
- Table (with sorting, pagination)
- Dropdown
- Tabs

---

## 5. Route Hierarchy (UX-First)

### 5.1 Design Philosophy
- **Routes are for users** - Organized by workflows, not API structure
- **Pages can call any APIs** - No 1:1 route-to-API mapping required
- **Logical groupings** - Sites, System, Apps, Security, External

### 5.2 Complete Route Map

```
/#/                          Dashboard
/#/sites                     Sites List (existing)
/#/sites/analytics           Aggregate Site Analytics
/#/sites/domains             Domain Management
/#/sites/create              Create New Site
/#/sites/:id                 Site Detail (future)

/#/system                    → Redirects to /system/stats
/#/system/stats              System Statistics
/#/system/limits             Resource Limits
/#/system/logs               System Logs
/#/system/backup             Backup Management
/#/system/health             Health Monitoring
/#/system/settings           System Settings
/#/system/design-system      Design System (keep in production)

/#/apps                      Apps List
/#/apps/webhooks             Webhooks
/#/apps/redirects            URL Redirects
/#/apps/tunnel               Tunnel Management
/#/apps/proxy                Proxy Configuration
/#/apps/botdaddy             BotDaddy (Telegram Bot)

/#/security                  → Redirects to /security/ssh
/#/security/ssh              SSH Keys
/#/security/tokens           Auth Tokens
/#/security/password         Password Settings

/#/external                  → Redirects to /external/cloudflare
/#/external/cloudflare       Cloudflare Integration
/#/external/litestream       Litestream Integration

/#/login                     Login (outside AppShell)
/#/*                         404 Not Found
```

### 5.3 Route-to-API Mapping

Routes don't need to mirror APIs. Here's how pages map to endpoints:

| Route | API Endpoints Called |
|-------|---------------------|
| `/` (Dashboard) | `/api/system/health`, `/api/sites`, `/api/analytics/stats` |
| `/sites` | `/api/sites` |
| `/sites/analytics` | `/api/analytics/stats`, `/api/analytics/domains` |
| `/sites/:id` | `/api/sites/:id`, `/api/env/:site`, `/api/keys` |
| `/system/stats` | `/api/system/health`, `/api/system/config` |
| `/system/health` | `/api/system/health` |
| `/apps/webhooks` | `/api/webhooks` |
| `/apps/redirects` | `/api/redirects` |

---

## 6. Implementation Plan

### Phase 2A: Route Hierarchy (CURRENT)

**Goal**: Implement new route structure with placeholders for all pages

**Files to Create**:

#### 6.1 Breadcrumbs Component

**File**: `admin/src/components/ui/Breadcrumbs.tsx`

```tsx
import { Link, useLocation } from 'react-router-dom';
import { ChevronRight, Home } from 'lucide-react';

const routeLabels: Record<string, string> = {
  sites: 'Sites',
  analytics: 'Analytics',
  domains: 'Domains',
  create: 'Create',
  system: 'System',
  stats: 'Statistics',
  limits: 'Limits',
  logs: 'Logs',
  backup: 'Backup',
  health: 'Health',
  settings: 'Settings',
  'design-system': 'Design System',
  apps: 'Apps',
  webhooks: 'Webhooks',
  redirects: 'Redirects',
  tunnel: 'Tunnel',
  proxy: 'Proxy',
  botdaddy: 'BotDaddy',
  security: 'Security',
  ssh: 'SSH Keys',
  tokens: 'Auth Tokens',
  password: 'Password',
  external: 'External',
  cloudflare: 'Cloudflare',
  litestream: 'Litestream',
};

export function Breadcrumbs() {
  const location = useLocation();
  const pathSegments = location.pathname.split('/').filter(Boolean);

  if (pathSegments.length === 0) return null;

  const breadcrumbs = pathSegments.map((segment, index) => {
    const path = '/' + pathSegments.slice(0, index + 1).join('/');
    const label = routeLabels[segment] || segment.charAt(0).toUpperCase() + segment.slice(1);
    return { label, path };
  });

  return (
    <nav className="flex items-center space-x-1 text-sm text-secondary mb-4">
      <Link to="/" className="flex items-center hover:text-primary transition-colors">
        <Home className="w-4 h-4" />
      </Link>
      {breadcrumbs.map((crumb, index) => (
        <div key={crumb.path} className="flex items-center">
          <ChevronRight className="w-4 h-4 mx-1 text-tertiary" />
          {index === breadcrumbs.length - 1 ? (
            <span className="text-primary font-medium">{crumb.label}</span>
          ) : (
            <Link to={crumb.path} className="hover:text-primary transition-colors">
              {crumb.label}
            </Link>
          )}
        </div>
      ))}
    </nav>
  );
}
```

#### 6.2 Layout Components

Create 5 layout components - all identical structure:

**File**: `admin/src/components/layout/SitesLayout.tsx`
```tsx
import { Outlet } from 'react-router-dom';
import { Breadcrumbs } from '../ui/Breadcrumbs';

export function SitesLayout() {
  return (
    <div>
      <Breadcrumbs />
      <Outlet />
    </div>
  );
}
```

Copy for: `SystemLayout.tsx`, `AppsLayout.tsx`, `SecurityLayout.tsx`, `ExternalLayout.tsx`

#### 6.3 404 Page

**File**: `admin/src/pages/NotFound.tsx`

```tsx
import { Link } from 'react-router-dom';
import { AlertCircle, Home } from 'lucide-react';

export function NotFound() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh]">
      <AlertCircle className="w-16 h-16 text-tertiary mb-4" />
      <h1 className="text-3xl font-bold text-primary mb-2">404 - Page Not Found</h1>
      <p className="text-secondary mb-6">The page you're looking for doesn't exist.</p>
      <Link
        to="/"
        className="inline-flex items-center gap-2 px-4 py-2 bg-accent text-white rounded-lg hover:bg-accent/90 transition-colors"
      >
        <Home className="w-4 h-4" />
        Back to Dashboard
      </Link>
    </div>
  );
}
```

#### 6.4 Placeholder Component

**File**: `admin/src/components/PlaceholderPage.tsx`

```tsx
import { LucideIcon, Construction } from 'lucide-react';

interface PlaceholderPageProps {
  title: string;
  description: string;
  icon: LucideIcon;
}

export function PlaceholderPage({ title, description, icon: Icon }: PlaceholderPageProps) {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-primary">{title}</h1>
        <p className="text-secondary mt-1">{description}</p>
      </div>

      <div className="bg-surface border border-default rounded-xl p-12">
        <div className="flex flex-col items-center justify-center text-center">
          <div className="w-16 h-16 rounded-2xl bg-accent/10 flex items-center justify-center mb-4">
            <Icon className="w-8 h-8 text-accent" />
          </div>
          <h2 className="text-xl font-medium text-primary mb-2">{title}</h2>
          <p className="text-secondary max-w-md mb-6">
            This page is under construction. The functionality will be available in a future update.
          </p>
          <div className="flex items-center gap-2 text-sm text-tertiary">
            <Construction className="w-4 h-4" />
            <span>Coming Soon</span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="bg-surface border border-default rounded-lg p-4">
            <div className="w-8 h-8 rounded-lg bg-tertiary/10 mb-3" />
            <div className="h-4 w-24 bg-tertiary/20 rounded mb-2" />
            <div className="h-3 w-32 bg-tertiary/10 rounded" />
          </div>
        ))}
      </div>
    </div>
  );
}
```

#### 6.5 All Placeholder Pages

Create these 18 files in their respective directories:

**Sites Group** (`admin/src/pages/sites/`):
- `SitesAnalytics.tsx` - icon: `BarChart3`, title: "Site Analytics"
- `SitesDomains.tsx` - icon: `Globe`, title: "Domain Management"
- `CreateSite.tsx` - icon: `Plus`, title: "Create Site"

**System Group** (`admin/src/pages/system/`):
- `SystemStats.tsx` - icon: `Activity`, title: "System Statistics"
- `SystemLimits.tsx` - icon: `Gauge`, title: "Resource Limits"
- `SystemLogs.tsx` - icon: `ScrollText`, title: "System Logs"
- `SystemBackup.tsx` - icon: `Archive`, title: "Backup Management"
- `SystemHealth.tsx` - icon: `HeartPulse`, title: "Health Monitoring"
- `SystemSettings.tsx` - icon: `Settings`, title: "System Settings"

**Apps Group** (`admin/src/pages/apps/`):
- `AppsList.tsx` - icon: `Package`, title: "Installed Apps"
- `Webhooks.tsx` - icon: `Webhook`, title: "Webhooks"
- `Redirects.tsx` - icon: `ExternalLink`, title: "URL Redirects"
- `Tunnel.tsx` - icon: `Route`, title: "Tunnel Management"
- `Proxy.tsx` - icon: `Network`, title: "Proxy Configuration"
- `BotDaddy.tsx` - icon: `Bot`, title: "BotDaddy"

**Security Group** (`admin/src/pages/security/`):
- `SecuritySSH.tsx` - icon: `Key`, title: "SSH Keys"
- `SecurityTokens.tsx` - icon: `KeyRound`, title: "Auth Tokens"
- `SecurityPassword.tsx` - icon: `Lock`, title: "Password Settings"

**External Group** (`admin/src/pages/external/`):
- `ExternalCloudflare.tsx` - icon: `Cloud`, title: "Cloudflare"
- `ExternalLitestream.tsx` - icon: `Database`, title: "Litestream"

**Template for each**:
```tsx
import { ICON_NAME } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function PAGE_NAME() {
  return (
    <PlaceholderPage
      title="TITLE"
      description="DESCRIPTION"
      icon={ICON_NAME}
    />
  );
}
```

#### 6.6 Updated App.tsx

**File**: `admin/src/App.tsx` (replace entire file)

```tsx
import { HashRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from './context/ThemeContext';
import { MockProvider } from './context/MockContext';
import { AuthProvider, useAuth } from './context/AuthContext';
import { queryClient } from './lib/queryClient';

// Layouts
import { AppShell } from './components/layout/AppShell';
import { SitesLayout } from './components/layout/SitesLayout';
import { SystemLayout } from './components/layout/SystemLayout';
import { AppsLayout } from './components/layout/AppsLayout';
import { SecurityLayout } from './components/layout/SecurityLayout';
import { ExternalLayout } from './components/layout/ExternalLayout';

// Main Pages
import { Login } from './pages/Login';
import { Dashboard } from './pages/Dashboard';
import { Sites } from './pages/Sites';
import { DesignSystem } from './pages/DesignSystem';
import { NotFound } from './pages/NotFound';

// Sites Pages
import { SitesAnalytics } from './pages/sites/SitesAnalytics';
import { SitesDomains } from './pages/sites/SitesDomains';
import { CreateSite } from './pages/sites/CreateSite';

// System Pages
import { SystemStats } from './pages/system/SystemStats';
import { SystemLimits } from './pages/system/SystemLimits';
import { SystemLogs } from './pages/system/SystemLogs';
import { SystemBackup } from './pages/system/SystemBackup';
import { SystemHealth } from './pages/system/SystemHealth';
import { SystemSettings } from './pages/system/SystemSettings';

// Apps Pages
import { AppsList } from './pages/apps/AppsList';
import { Webhooks } from './pages/apps/Webhooks';
import { Redirects } from './pages/apps/Redirects';
import { Tunnel } from './pages/apps/Tunnel';
import { Proxy } from './pages/apps/Proxy';
import { BotDaddy } from './pages/apps/BotDaddy';

// Security Pages
import { SecuritySSH } from './pages/security/SecuritySSH';
import { SecurityTokens } from './pages/security/SecurityTokens';
import { SecurityPassword } from './pages/security/SecurityPassword';

// External Pages
import { ExternalCloudflare } from './pages/external/ExternalCloudflare';
import { ExternalLitestream } from './pages/external/ExternalLitestream';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

function AppRoutes() {
  const { isAuthenticated } = useAuth();

  return (
    <Routes>
      <Route path="/login" element={isAuthenticated ? <Navigate to="/" replace /> : <Login />} />

      <Route path="/" element={<ProtectedRoute><AppShell /></ProtectedRoute>}>
        <Route index element={<Dashboard />} />

        <Route path="sites" element={<SitesLayout />}>
          <Route index element={<Sites />} />
          <Route path="analytics" element={<SitesAnalytics />} />
          <Route path="domains" element={<SitesDomains />} />
          <Route path="create" element={<CreateSite />} />
        </Route>

        <Route path="system" element={<SystemLayout />}>
          <Route index element={<Navigate to="stats" replace />} />
          <Route path="stats" element={<SystemStats />} />
          <Route path="limits" element={<SystemLimits />} />
          <Route path="logs" element={<SystemLogs />} />
          <Route path="backup" element={<SystemBackup />} />
          <Route path="health" element={<SystemHealth />} />
          <Route path="settings" element={<SystemSettings />} />
          <Route path="design-system" element={<DesignSystem />} />
        </Route>

        <Route path="apps" element={<AppsLayout />}>
          <Route index element={<AppsList />} />
          <Route path="webhooks" element={<Webhooks />} />
          <Route path="redirects" element={<Redirects />} />
          <Route path="tunnel" element={<Tunnel />} />
          <Route path="proxy" element={<Proxy />} />
          <Route path="botdaddy" element={<BotDaddy />} />
        </Route>

        <Route path="security" element={<SecurityLayout />}>
          <Route index element={<Navigate to="ssh" replace />} />
          <Route path="ssh" element={<SecuritySSH />} />
          <Route path="tokens" element={<SecurityTokens />} />
          <Route path="password" element={<SecurityPassword />} />
        </Route>

        <Route path="external" element={<ExternalLayout />}>
          <Route index element={<Navigate to="cloudflare" replace />} />
          <Route path="cloudflare" element={<ExternalCloudflare />} />
          <Route path="litestream" element={<ExternalLitestream />} />
        </Route>

        <Route path="*" element={<NotFound />} />
      </Route>
    </Routes>
  );
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <MockProvider>
          <AuthProvider>
            <HashRouter>
              <AppRoutes />
            </HashRouter>
          </AuthProvider>
        </MockProvider>
      </ThemeProvider>
    </QueryClientProvider>
  );
}

export default App;
```

#### 6.7 Updated Sidebar Menu

Update `admin/src/components/layout/Sidebar.tsx` with new menu structure:

```tsx
const menuItems = [
  {
    label: 'Dashboard',
    icon: LayoutDashboard,
    path: '/',
    exact: true,
  },
  {
    label: 'Sites',
    icon: Globe,
    path: '/sites',
    children: [
      { label: 'All Sites', path: '/sites', icon: Globe },
      { label: 'Analytics', path: '/sites/analytics', icon: BarChart3 },
      { label: 'Domains', path: '/sites/domains', icon: GlobeLock },
      { label: 'Create Site', path: '/sites/create', icon: Plus },
    ],
  },
  {
    label: 'System',
    icon: Settings,
    path: '/system',
    children: [
      { label: 'Statistics', path: '/system/stats', icon: Activity },
      { label: 'Resource Limits', path: '/system/limits', icon: Gauge },
      { label: 'Logs', path: '/system/logs', icon: ScrollText },
      { label: 'Backup', path: '/system/backup', icon: Archive },
      { label: 'Health', path: '/system/health', icon: HeartPulse },
      { label: 'Settings', path: '/system/settings', icon: Settings },
      { label: 'Design System', path: '/system/design-system', icon: Palette },
    ],
  },
  {
    label: 'Apps',
    icon: Package,
    path: '/apps',
    children: [
      { label: 'All Apps', path: '/apps', icon: Package },
      { label: 'Webhooks', path: '/apps/webhooks', icon: Webhook },
      { label: 'Redirects', path: '/apps/redirects', icon: ExternalLink },
      { label: 'Tunnel', path: '/apps/tunnel', icon: Route },
      { label: 'Proxy', path: '/apps/proxy', icon: Network },
      { label: 'BotDaddy', path: '/apps/botdaddy', icon: Bot },
    ],
  },
  {
    label: 'Security',
    icon: Shield,
    path: '/security',
    children: [
      { label: 'SSH Keys', path: '/security/ssh', icon: Key },
      { label: 'Auth Tokens', path: '/security/tokens', icon: KeyRound },
      { label: 'Password', path: '/security/password', icon: Lock },
    ],
  },
  {
    label: 'External',
    icon: Link2,
    path: '/external',
    children: [
      { label: 'Cloudflare', path: '/external/cloudflare', icon: Cloud },
      { label: 'Litestream', path: '/external/litestream', icon: Database },
    ],
  },
];
```

---

### Phase 2B: Core Pages (After Phase 2A)

**Goal**: Implement real functionality for high-priority pages

**Priority Order**:

1. **Site Detail** (`/sites/:id`) - Most requested feature
   - Overview tab with site info
   - Environment variables
   - API keys
   - Logs tab
   - Settings tab

2. **Webhooks** (`/apps/webhooks`) - Full CRUD
   - List webhooks
   - Create webhook
   - Edit webhook
   - Delete webhook

3. **Redirects** (`/apps/redirects`) - Full CRUD
   - List redirects
   - Create redirect
   - Edit redirect
   - Delete redirect
   - Copy link

4. **System Stats** (`/system/stats`) - Dashboard-like overview
   - System health
   - Resource usage
   - Database size

5. **System Health** (`/system/health`) - Health monitoring
   - Service status
   - Uptime
   - Alerts

---

### Phase 3: Polish & Production

**Goal**: Production-ready

**Deliverables**:
1. Real API integration (replace mock)
2. Form validation
3. Error handling
4. Loading states everywhere
5. Toast notifications
6. Keyboard shortcuts
7. Performance optimization
8. Accessibility audit

---

## 7. File Summary

### Files to Create (Phase 2A)

| # | File | Type |
|---|------|------|
| 1 | `components/ui/Breadcrumbs.tsx` | Component |
| 2 | `components/PlaceholderPage.tsx` | Component |
| 3 | `components/layout/SitesLayout.tsx` | Layout |
| 4 | `components/layout/SystemLayout.tsx` | Layout |
| 5 | `components/layout/AppsLayout.tsx` | Layout |
| 6 | `components/layout/SecurityLayout.tsx` | Layout |
| 7 | `components/layout/ExternalLayout.tsx` | Layout |
| 8 | `pages/NotFound.tsx` | Page |
| 9 | `pages/sites/SitesAnalytics.tsx` | Page |
| 10 | `pages/sites/SitesDomains.tsx` | Page |
| 11 | `pages/sites/CreateSite.tsx` | Page |
| 12 | `pages/system/SystemStats.tsx` | Page |
| 13 | `pages/system/SystemLimits.tsx` | Page |
| 14 | `pages/system/SystemLogs.tsx` | Page |
| 15 | `pages/system/SystemBackup.tsx` | Page |
| 16 | `pages/system/SystemHealth.tsx` | Page |
| 17 | `pages/system/SystemSettings.tsx` | Page |
| 18 | `pages/apps/AppsList.tsx` | Page |
| 19 | `pages/apps/Webhooks.tsx` | Page |
| 20 | `pages/apps/Redirects.tsx` | Page |
| 21 | `pages/apps/Tunnel.tsx` | Page |
| 22 | `pages/apps/Proxy.tsx` | Page |
| 23 | `pages/apps/BotDaddy.tsx` | Page |
| 24 | `pages/security/SecuritySSH.tsx` | Page |
| 25 | `pages/security/SecurityTokens.tsx` | Page |
| 26 | `pages/security/SecurityPassword.tsx` | Page |
| 27 | `pages/external/ExternalCloudflare.tsx` | Page |
| 28 | `pages/external/ExternalLitestream.tsx` | Page |

### Files to Modify (Phase 2A)

| # | File | Changes |
|---|------|---------|
| 1 | `App.tsx` | Replace with new routing structure |
| 2 | `components/layout/Sidebar.tsx` | Update menuItems array |

**Total**: 28 new files + 2 modified = 30 file operations

---

## 8. Implementation Order

Execute in this exact sequence:

1. Create directories under `pages/`: `sites/`, `system/`, `apps/`, `security/`, `external/`
2. Create `Breadcrumbs.tsx`
3. Create `PlaceholderPage.tsx`
4. Create 5 layout components
5. Create `NotFound.tsx`
6. Create all 18 placeholder pages
7. Update `App.tsx`
8. Update `Sidebar.tsx` menuItems
9. Test all routes work
10. Verify sidebar highlighting
11. Test breadcrumbs on all pages

---

## 9. Verification Checklist

After implementation, test each route:

| Route | Expected |
|-------|----------|
| `/#/` | Dashboard |
| `/#/sites` | Sites list |
| `/#/sites/analytics` | Placeholder |
| `/#/sites/domains` | Placeholder |
| `/#/sites/create` | Placeholder |
| `/#/system` | Redirects to /system/stats |
| `/#/system/stats` | Placeholder |
| `/#/system/limits` | Placeholder |
| `/#/system/logs` | Placeholder |
| `/#/system/backup` | Placeholder |
| `/#/system/health` | Placeholder |
| `/#/system/settings` | Placeholder |
| `/#/system/design-system` | Design System |
| `/#/apps` | Placeholder |
| `/#/apps/webhooks` | Placeholder |
| `/#/apps/redirects` | Placeholder |
| `/#/apps/tunnel` | Placeholder |
| `/#/apps/proxy` | Placeholder |
| `/#/apps/botdaddy` | Placeholder |
| `/#/security` | Redirects to /security/ssh |
| `/#/security/ssh` | Placeholder |
| `/#/security/tokens` | Placeholder |
| `/#/security/password` | Placeholder |
| `/#/external` | Redirects to /external/cloudflare |
| `/#/external/cloudflare` | Placeholder |
| `/#/external/litestream` | Placeholder |
| `/#/nonexistent` | 404 page |

---

## 10. API Reference

### Available Endpoints (~30)

**Auth**:
- `POST /api/auth/login`
- `POST /api/auth/logout`
- `GET /api/user/me`
- `GET /api/auth/status`

**Sites**:
- `GET /api/sites`
- `POST /api/sites`
- `GET /api/sites/:id`
- `PUT /api/sites/:id`
- `DELETE /api/sites/:id`
- `POST /api/deploy`

**Environment**:
- `GET /api/env/:site`
- `POST /api/env/:site`
- `DELETE /api/env/:site/:key`

**API Keys**:
- `GET /api/keys`
- `POST /api/keys`
- `DELETE /api/keys/:id`

**Analytics**:
- `GET /api/analytics/stats`
- `GET /api/analytics/events`
- `GET /api/analytics/domains`
- `GET /api/analytics/tags`

**Redirects**:
- `GET /api/redirects`
- `POST /api/redirects`
- `PUT /api/redirects/:id`
- `DELETE /api/redirects/:id`

**Webhooks**:
- `GET /api/webhooks`
- `POST /api/webhooks`
- `PUT /api/webhooks/:id`
- `DELETE /api/webhooks/:id`

**System**:
- `GET /api/system/health`
- `GET /api/system/config`
- `PUT /api/system/config`

**Logs**:
- `GET /api/sites/:site/logs`
- `GET /api/deployments`

---

## 11. Development Commands

```bash
# Start dev server
cd /home/testman/workspace/admin
npm run dev -- --port 37180 --host 0.0.0.0

# Build for production
npm run build

# Copy to Go embed location
cd /home/testman/workspace
rm -rf internal/assets/system/admin
cp -r admin/dist internal/assets/system/admin

# Build Go binary
go build -o fazt ./cmd/server
```

---

## 12. Success Criteria

### Phase 2A Complete When:
- [ ] All 25 routes accessible
- [ ] All routes show appropriate content (pages or placeholders)
- [ ] Breadcrumbs work on all nested pages
- [ ] Sidebar highlights correctly for all routes
- [ ] 404 page shows for invalid routes
- [ ] No console errors
- [ ] Build succeeds

### Project Complete When:
- [ ] All core pages have real functionality
- [ ] Real API integration working
- [ ] Form validation implemented
- [ ] Error handling implemented
- [ ] Loading states everywhere
- [ ] PWA installable
- [ ] Lighthouse score > 90

---

**Plan Status**: ✅ ACTIVE
**Current Phase**: 2A (Route Hierarchy)
**Next Action**: Create files as specified in Section 6

---

**Remember**: Routes are for users. APIs are for data. Build to delight.
