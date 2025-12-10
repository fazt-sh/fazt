# Route Hierarchy Implementation Plan (FINAL)

**Date**: December 9, 2025
**Status**: ⚠️ SUPERSEDED by Plan 14
**Superseded By**: `koder/plans/14_admin-spa-complete.md`
**Approach**: UX-First (Routes optimized for user workflows)

> **Note**: This plan's route hierarchy has been merged into Plan 14, which is now the single source of truth for the Admin SPA.

---

## 1. Decisions (FINAL)

| Decision | Answer |
|----------|--------|
| Backwards Compatibility | ❌ No redirects needed (no users yet) |
| Analytics Location | `/sites/analytics` only (Option A) |
| Design System | ✅ Keep in production at `/system/design-system` |
| Breadcrumbs | ✅ Implement |
| Implementation | All routes with pretty placeholders first |

---

## 2. Complete Route Structure

```
/#/                          Dashboard
/#/sites                     Sites List
/#/sites/analytics           Aggregate Site Analytics
/#/sites/domains             Domain Management
/#/sites/create              Create New Site
/#/sites/:id                 Site Detail (tabs inside)

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

/#/login                     Login Page (outside AppShell)
/#/*                         404 Not Found
```

---

## 3. File Structure to Create

```
admin/src/
├── components/
│   ├── layout/
│   │   ├── SitesLayout.tsx       (NEW)
│   │   ├── SystemLayout.tsx      (NEW)
│   │   ├── AppsLayout.tsx        (NEW)
│   │   ├── SecurityLayout.tsx    (NEW)
│   │   └── ExternalLayout.tsx    (NEW)
│   └── ui/
│       └── Breadcrumbs.tsx       (NEW)
├── pages/
│   ├── NotFound.tsx              (NEW)
│   ├── sites/
│   │   ├── SitesAnalytics.tsx    (NEW)
│   │   ├── SitesDomains.tsx      (NEW)
│   │   └── CreateSite.tsx        (NEW)
│   ├── system/
│   │   ├── SystemStats.tsx       (NEW)
│   │   ├── SystemLimits.tsx      (NEW)
│   │   ├── SystemLogs.tsx        (NEW)
│   │   ├── SystemBackup.tsx      (NEW)
│   │   ├── SystemHealth.tsx      (NEW)
│   │   └── SystemSettings.tsx    (NEW)
│   ├── apps/
│   │   ├── AppsList.tsx          (NEW)
│   │   ├── Webhooks.tsx          (NEW - full page)
│   │   ├── Redirects.tsx         (NEW - full page)
│   │   ├── Tunnel.tsx            (NEW)
│   │   ├── Proxy.tsx             (NEW)
│   │   └── BotDaddy.tsx          (NEW)
│   ├── security/
│   │   ├── SecuritySSH.tsx       (NEW)
│   │   ├── SecurityTokens.tsx    (NEW)
│   │   └── SecurityPassword.tsx  (NEW)
│   └── external/
│       ├── ExternalCloudflare.tsx (NEW)
│       └── ExternalLitestream.tsx (NEW)
└── App.tsx                       (MODIFY)
```

---

## 4. Implementation Steps

### Step 1: Create Breadcrumbs Component

**File**: `admin/src/components/ui/Breadcrumbs.tsx`

```tsx
import { Link, useLocation } from 'react-router-dom';
import { ChevronRight, Home } from 'lucide-react';

// Human-readable labels for route segments
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

  // Don't show breadcrumbs on dashboard
  if (pathSegments.length === 0) {
    return null;
  }

  const breadcrumbs = pathSegments.map((segment, index) => {
    const path = '/' + pathSegments.slice(0, index + 1).join('/');
    const label = routeLabels[segment] || segment.charAt(0).toUpperCase() + segment.slice(1);
    return { label, path };
  });

  return (
    <nav className="flex items-center space-x-1 text-sm text-secondary mb-4">
      <Link
        to="/"
        className="flex items-center hover:text-primary transition-colors"
      >
        <Home className="w-4 h-4" />
      </Link>
      {breadcrumbs.map((crumb, index) => (
        <div key={crumb.path} className="flex items-center">
          <ChevronRight className="w-4 h-4 mx-1 text-tertiary" />
          {index === breadcrumbs.length - 1 ? (
            <span className="text-primary font-medium">{crumb.label}</span>
          ) : (
            <Link
              to={crumb.path}
              className="hover:text-primary transition-colors"
            >
              {crumb.label}
            </Link>
          )}
        </div>
      ))}
    </nav>
  );
}
```

---

### Step 2: Create Layout Components

Each layout is a simple pass-through with Breadcrumbs.

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

**File**: `admin/src/components/layout/SystemLayout.tsx`

```tsx
import { Outlet } from 'react-router-dom';
import { Breadcrumbs } from '../ui/Breadcrumbs';

export function SystemLayout() {
  return (
    <div>
      <Breadcrumbs />
      <Outlet />
    </div>
  );
}
```

**File**: `admin/src/components/layout/AppsLayout.tsx`

```tsx
import { Outlet } from 'react-router-dom';
import { Breadcrumbs } from '../ui/Breadcrumbs';

export function AppsLayout() {
  return (
    <div>
      <Breadcrumbs />
      <Outlet />
    </div>
  );
}
```

**File**: `admin/src/components/layout/SecurityLayout.tsx`

```tsx
import { Outlet } from 'react-router-dom';
import { Breadcrumbs } from '../ui/Breadcrumbs';

export function SecurityLayout() {
  return (
    <div>
      <Breadcrumbs />
      <Outlet />
    </div>
  );
}
```

**File**: `admin/src/components/layout/ExternalLayout.tsx`

```tsx
import { Outlet } from 'react-router-dom';
import { Breadcrumbs } from '../ui/Breadcrumbs';

export function ExternalLayout() {
  return (
    <div>
      <Breadcrumbs />
      <Outlet />
    </div>
  );
}
```

---

### Step 3: Create 404 Page

**File**: `admin/src/pages/NotFound.tsx`

```tsx
import { Link } from 'react-router-dom';
import { AlertCircle, Home } from 'lucide-react';

export function NotFound() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh]">
      <AlertCircle className="w-16 h-16 text-tertiary mb-4" />
      <h1 className="text-3xl font-bold text-primary mb-2">
        404 - Page Not Found
      </h1>
      <p className="text-secondary mb-6">
        The page you're looking for doesn't exist.
      </p>
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

---

### Step 4: Create Placeholder Page Component

**File**: `admin/src/components/PlaceholderPage.tsx`

A reusable pretty placeholder for pages not yet implemented.

```tsx
import { LucideIcon, Construction, ArrowLeft } from 'lucide-react';
import { Link } from 'react-router-dom';

interface PlaceholderPageProps {
  title: string;
  description: string;
  icon: LucideIcon;
  parentPath?: string;
  parentLabel?: string;
}

export function PlaceholderPage({
  title,
  description,
  icon: Icon,
  parentPath,
  parentLabel
}: PlaceholderPageProps) {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-primary">{title}</h1>
          <p className="text-secondary mt-1">{description}</p>
        </div>
        {parentPath && (
          <Link
            to={parentPath}
            className="inline-flex items-center gap-2 text-sm text-secondary hover:text-primary transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            {parentLabel || 'Back'}
          </Link>
        )}
      </div>

      {/* Placeholder Content */}
      <div className="bg-surface border border-default rounded-xl p-12">
        <div className="flex flex-col items-center justify-center text-center">
          <div className="w-16 h-16 rounded-2xl bg-accent/10 flex items-center justify-center mb-4">
            <Icon className="w-8 h-8 text-accent" />
          </div>
          <h2 className="text-xl font-medium text-primary mb-2">
            {title}
          </h2>
          <p className="text-secondary max-w-md mb-6">
            This page is under construction. The functionality will be available in a future update.
          </p>
          <div className="flex items-center gap-2 text-sm text-tertiary">
            <Construction className="w-4 h-4" />
            <span>Coming Soon</span>
          </div>
        </div>
      </div>

      {/* Feature Preview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-surface border border-default rounded-lg p-4">
          <div className="w-8 h-8 rounded-lg bg-blue-500/10 flex items-center justify-center mb-3">
            <div className="w-4 h-4 rounded bg-blue-500/30" />
          </div>
          <div className="h-4 w-24 bg-tertiary/20 rounded mb-2" />
          <div className="h-3 w-32 bg-tertiary/10 rounded" />
        </div>
        <div className="bg-surface border border-default rounded-lg p-4">
          <div className="w-8 h-8 rounded-lg bg-green-500/10 flex items-center justify-center mb-3">
            <div className="w-4 h-4 rounded bg-green-500/30" />
          </div>
          <div className="h-4 w-20 bg-tertiary/20 rounded mb-2" />
          <div className="h-3 w-28 bg-tertiary/10 rounded" />
        </div>
        <div className="bg-surface border border-default rounded-lg p-4">
          <div className="w-8 h-8 rounded-lg bg-purple-500/10 flex items-center justify-center mb-3">
            <div className="w-4 h-4 rounded bg-purple-500/30" />
          </div>
          <div className="h-4 w-28 bg-tertiary/20 rounded mb-2" />
          <div className="h-3 w-24 bg-tertiary/10 rounded" />
        </div>
      </div>
    </div>
  );
}
```

---

### Step 5: Create All Placeholder Pages

Create these files with the exact content shown:

#### Sites Group

**File**: `admin/src/pages/sites/SitesAnalytics.tsx`

```tsx
import { BarChart3 } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function SitesAnalytics() {
  return (
    <PlaceholderPage
      title="Site Analytics"
      description="Aggregate analytics across all your sites"
      icon={BarChart3}
    />
  );
}
```

**File**: `admin/src/pages/sites/SitesDomains.tsx`

```tsx
import { Globe } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function SitesDomains() {
  return (
    <PlaceholderPage
      title="Domain Management"
      description="Manage custom domains for your sites"
      icon={Globe}
    />
  );
}
```

**File**: `admin/src/pages/sites/CreateSite.tsx`

```tsx
import { Plus } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function CreateSite() {
  return (
    <PlaceholderPage
      title="Create Site"
      description="Create a new site to deploy your content"
      icon={Plus}
      parentPath="/sites"
      parentLabel="Back to Sites"
    />
  );
}
```

#### System Group

**File**: `admin/src/pages/system/SystemStats.tsx`

```tsx
import { Activity } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function SystemStats() {
  return (
    <PlaceholderPage
      title="System Statistics"
      description="Overview of system performance and resource usage"
      icon={Activity}
    />
  );
}
```

**File**: `admin/src/pages/system/SystemLimits.tsx`

```tsx
import { Gauge } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function SystemLimits() {
  return (
    <PlaceholderPage
      title="Resource Limits"
      description="Configure safeguards and resource limits"
      icon={Gauge}
    />
  );
}
```

**File**: `admin/src/pages/system/SystemLogs.tsx`

```tsx
import { ScrollText } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function SystemLogs() {
  return (
    <PlaceholderPage
      title="System Logs"
      description="View system logs and debug information"
      icon={ScrollText}
    />
  );
}
```

**File**: `admin/src/pages/system/SystemBackup.tsx`

```tsx
import { Archive } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function SystemBackup() {
  return (
    <PlaceholderPage
      title="Backup Management"
      description="Create and restore database backups"
      icon={Archive}
    />
  );
}
```

**File**: `admin/src/pages/system/SystemHealth.tsx`

```tsx
import { HeartPulse } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function SystemHealth() {
  return (
    <PlaceholderPage
      title="Health Monitoring"
      description="Monitor system health and service status"
      icon={HeartPulse}
    />
  );
}
```

**File**: `admin/src/pages/system/SystemSettings.tsx`

```tsx
import { Settings } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function SystemSettings() {
  return (
    <PlaceholderPage
      title="System Settings"
      description="Configure system-wide settings"
      icon={Settings}
    />
  );
}
```

#### Apps Group

**File**: `admin/src/pages/apps/AppsList.tsx`

```tsx
import { Package } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function AppsList() {
  return (
    <PlaceholderPage
      title="Installed Apps"
      description="Manage your installed applications"
      icon={Package}
    />
  );
}
```

**File**: `admin/src/pages/apps/Webhooks.tsx`

```tsx
import { Webhook } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function Webhooks() {
  return (
    <PlaceholderPage
      title="Webhooks"
      description="Manage webhook endpoints for integrations"
      icon={Webhook}
    />
  );
}
```

**File**: `admin/src/pages/apps/Redirects.tsx`

```tsx
import { ExternalLink } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function Redirects() {
  return (
    <PlaceholderPage
      title="URL Redirects"
      description="Manage URL shortcuts and redirects"
      icon={ExternalLink}
    />
  );
}
```

**File**: `admin/src/pages/apps/Tunnel.tsx`

```tsx
import { Route } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function Tunnel() {
  return (
    <PlaceholderPage
      title="Tunnel Management"
      description="Configure secure tunnels for local development"
      icon={Route}
    />
  );
}
```

**File**: `admin/src/pages/apps/Proxy.tsx`

```tsx
import { Network } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function Proxy() {
  return (
    <PlaceholderPage
      title="Proxy Configuration"
      description="Set up reverse proxy rules"
      icon={Network}
    />
  );
}
```

**File**: `admin/src/pages/apps/BotDaddy.tsx`

```tsx
import { Bot } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function BotDaddy() {
  return (
    <PlaceholderPage
      title="BotDaddy"
      description="Telegram bot server management"
      icon={Bot}
    />
  );
}
```

#### Security Group

**File**: `admin/src/pages/security/SecuritySSH.tsx`

```tsx
import { Key } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function SecuritySSH() {
  return (
    <PlaceholderPage
      title="SSH Keys"
      description="Manage SSH keys for secure access"
      icon={Key}
    />
  );
}
```

**File**: `admin/src/pages/security/SecurityTokens.tsx`

```tsx
import { KeyRound } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function SecurityTokens() {
  return (
    <PlaceholderPage
      title="Auth Tokens"
      description="Manage API authentication tokens"
      icon={KeyRound}
    />
  );
}
```

**File**: `admin/src/pages/security/SecurityPassword.tsx`

```tsx
import { Lock } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function SecurityPassword() {
  return (
    <PlaceholderPage
      title="Password Settings"
      description="Update your password and security settings"
      icon={Lock}
    />
  );
}
```

#### External Group

**File**: `admin/src/pages/external/ExternalCloudflare.tsx`

```tsx
import { Cloud } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function ExternalCloudflare() {
  return (
    <PlaceholderPage
      title="Cloudflare"
      description="Configure Cloudflare integration"
      icon={Cloud}
    />
  );
}
```

**File**: `admin/src/pages/external/ExternalLitestream.tsx`

```tsx
import { Database } from 'lucide-react';
import { PlaceholderPage } from '../../components/PlaceholderPage';

export function ExternalLitestream() {
  return (
    <PlaceholderPage
      title="Litestream"
      description="Configure Litestream for database replication"
      icon={Database}
    />
  );
}
```

---

### Step 6: Update App.tsx

**File**: `admin/src/App.tsx`

Replace the entire file with this content:

```tsx
import { HashRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from './context/ThemeContext';
import { MockProvider } from './context/MockContext';
import { AuthProvider, useAuth } from './context/AuthContext';
import { queryClient } from './lib/queryClient';

// Layout
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

// Protected Route wrapper
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

function AppRoutes() {
  const { isAuthenticated } = useAuth();

  return (
    <Routes>
      {/* Login (outside AppShell) */}
      <Route path="/login" element={isAuthenticated ? <Navigate to="/" replace /> : <Login />} />

      {/* Protected Routes */}
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <AppShell />
          </ProtectedRoute>
        }
      >
        {/* Dashboard */}
        <Route index element={<Dashboard />} />

        {/* Sites Group */}
        <Route path="sites" element={<SitesLayout />}>
          <Route index element={<Sites />} />
          <Route path="analytics" element={<SitesAnalytics />} />
          <Route path="domains" element={<SitesDomains />} />
          <Route path="create" element={<CreateSite />} />
        </Route>

        {/* System Group */}
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

        {/* Apps Group */}
        <Route path="apps" element={<AppsLayout />}>
          <Route index element={<AppsList />} />
          <Route path="webhooks" element={<Webhooks />} />
          <Route path="redirects" element={<Redirects />} />
          <Route path="tunnel" element={<Tunnel />} />
          <Route path="proxy" element={<Proxy />} />
          <Route path="botdaddy" element={<BotDaddy />} />
        </Route>

        {/* Security Group */}
        <Route path="security" element={<SecurityLayout />}>
          <Route index element={<Navigate to="ssh" replace />} />
          <Route path="ssh" element={<SecuritySSH />} />
          <Route path="tokens" element={<SecurityTokens />} />
          <Route path="password" element={<SecurityPassword />} />
        </Route>

        {/* External Group */}
        <Route path="external" element={<ExternalLayout />}>
          <Route index element={<Navigate to="cloudflare" replace />} />
          <Route path="cloudflare" element={<ExternalCloudflare />} />
          <Route path="litestream" element={<ExternalLitestream />} />
        </Route>

        {/* 404 Catch-all */}
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

---

### Step 7: Update Sidebar Menu

**File**: `admin/src/components/layout/Sidebar.tsx`

Update the `menuItems` array to use this exact structure:

```tsx
import {
  LayoutDashboard,
  Globe,
  Settings,
  Package,
  Shield,
  Link2,
  BarChart3,
  GlobeLock,
  Plus,
  Activity,
  Gauge,
  ScrollText,
  Archive,
  HeartPulse,
  Palette,
  Webhook,
  ExternalLink,
  Route,
  Network,
  Bot,
  Key,
  KeyRound,
  Lock,
  Cloud,
  Database,
} from 'lucide-react';

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

## 5. Verification Checklist

After implementation, verify each route works:

| Route | Expected Behavior |
|-------|-------------------|
| `/#/` | Dashboard loads |
| `/#/sites` | Sites list (existing page) |
| `/#/sites/analytics` | Pretty placeholder |
| `/#/sites/domains` | Pretty placeholder |
| `/#/sites/create` | Pretty placeholder with "Back to Sites" |
| `/#/system` | Redirects to `/system/stats` |
| `/#/system/stats` | Pretty placeholder |
| `/#/system/limits` | Pretty placeholder |
| `/#/system/logs` | Pretty placeholder |
| `/#/system/backup` | Pretty placeholder |
| `/#/system/health` | Pretty placeholder |
| `/#/system/settings` | Pretty placeholder |
| `/#/system/design-system` | Design System (existing page) |
| `/#/apps` | Pretty placeholder |
| `/#/apps/webhooks` | Pretty placeholder |
| `/#/apps/redirects` | Pretty placeholder |
| `/#/apps/tunnel` | Pretty placeholder |
| `/#/apps/proxy` | Pretty placeholder |
| `/#/apps/botdaddy` | Pretty placeholder |
| `/#/security` | Redirects to `/security/ssh` |
| `/#/security/ssh` | Pretty placeholder |
| `/#/security/tokens` | Pretty placeholder |
| `/#/security/password` | Pretty placeholder |
| `/#/external` | Redirects to `/external/cloudflare` |
| `/#/external/cloudflare` | Pretty placeholder |
| `/#/external/litestream` | Pretty placeholder |
| `/#/nonexistent` | 404 page |
| `/#/login` | Login page (outside AppShell) |

---

## 6. Sidebar Behavior

| Action | Expected |
|--------|----------|
| Click "Dashboard" | Navigate to `/`, highlight Dashboard |
| Click "Sites" | Expand sub-menu |
| Click "All Sites" | Navigate to `/sites`, highlight Sites > All Sites |
| Navigate to `/sites/analytics` | Sites expanded, Analytics highlighted |
| Navigate to `/system/stats` | System expanded, Statistics highlighted |
| Navigate to `/` | All sub-menus can be collapsed, only Dashboard highlighted |

---

## 7. Files Summary

### New Files to Create (24 files):

**Components (6 files):**
1. `admin/src/components/ui/Breadcrumbs.tsx`
2. `admin/src/components/PlaceholderPage.tsx`
3. `admin/src/components/layout/SitesLayout.tsx`
4. `admin/src/components/layout/SystemLayout.tsx`
5. `admin/src/components/layout/AppsLayout.tsx`
6. `admin/src/components/layout/SecurityLayout.tsx`
7. `admin/src/components/layout/ExternalLayout.tsx`

**Pages (18 files):**
1. `admin/src/pages/NotFound.tsx`
2. `admin/src/pages/sites/SitesAnalytics.tsx`
3. `admin/src/pages/sites/SitesDomains.tsx`
4. `admin/src/pages/sites/CreateSite.tsx`
5. `admin/src/pages/system/SystemStats.tsx`
6. `admin/src/pages/system/SystemLimits.tsx`
7. `admin/src/pages/system/SystemLogs.tsx`
8. `admin/src/pages/system/SystemBackup.tsx`
9. `admin/src/pages/system/SystemHealth.tsx`
10. `admin/src/pages/system/SystemSettings.tsx`
11. `admin/src/pages/apps/AppsList.tsx`
12. `admin/src/pages/apps/Webhooks.tsx`
13. `admin/src/pages/apps/Redirects.tsx`
14. `admin/src/pages/apps/Tunnel.tsx`
15. `admin/src/pages/apps/Proxy.tsx`
16. `admin/src/pages/apps/BotDaddy.tsx`
17. `admin/src/pages/security/SecuritySSH.tsx`
18. `admin/src/pages/security/SecurityTokens.tsx`
19. `admin/src/pages/security/SecurityPassword.tsx`
20. `admin/src/pages/external/ExternalCloudflare.tsx`
21. `admin/src/pages/external/ExternalLitestream.tsx`

### Files to Modify (2 files):
1. `admin/src/App.tsx` - Replace entire content
2. `admin/src/components/layout/Sidebar.tsx` - Update menuItems array

---

## 8. Implementation Order

Execute in this exact order:

1. Create directories: `sites/`, `system/`, `apps/`, `security/`, `external/` under `pages/`
2. Create `Breadcrumbs.tsx`
3. Create `PlaceholderPage.tsx`
4. Create all 5 layout components
5. Create `NotFound.tsx`
6. Create all 18 placeholder pages
7. Update `App.tsx`
8. Update `Sidebar.tsx` menuItems
9. Test all routes
10. Verify sidebar highlighting

---

**Plan Status**: ✅ FINAL - Ready for Implementation
**Estimated Files**: 24 new + 2 modified = 26 total
**Complexity**: Low (mostly copy-paste with small variations)
