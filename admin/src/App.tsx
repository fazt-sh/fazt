import { HashRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from './context/ThemeContext';
import { MockProvider } from './context/MockContext';
import { AuthProvider, useAuth } from './context/AuthContext';
import { ToastProvider } from './context/ToastContext';
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
import { Profile } from './pages/Profile';
import { DesignSystem } from './pages/DesignSystem';
import { NotFound } from './pages/NotFound';

// Sites Pages
import { SitesAnalytics } from './pages/sites/SitesAnalytics';
import { SitesDomains } from './pages/sites/SitesDomains';
import { CreateSite } from './pages/sites/CreateSite';
import { SiteDetail } from './pages/sites/SiteDetail';

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
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="profile" element={<Profile />} />

        <Route path="sites" element={<SitesLayout />}>
          <Route index element={<Sites />} />
          <Route path="analytics" element={<SitesAnalytics />} />
          <Route path="domains" element={<SitesDomains />} />
          <Route path="create" element={<CreateSite />} />
          <Route path=":id" element={<SiteDetail />} />
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
          <ToastProvider>
            <AuthProvider>
              <HashRouter>
                <AppRoutes />
              </HashRouter>
            </AuthProvider>
          </ToastProvider>
        </MockProvider>
      </ThemeProvider>
    </QueryClientProvider>
  );
}

export default App;