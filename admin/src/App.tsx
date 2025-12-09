import { HashRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from './context/ThemeContext';
import { MockProvider } from './context/MockContext';
import { AuthProvider, useAuth } from './context/AuthContext';
import { queryClient } from './lib/queryClient';
import { AppShell } from './components/layout/AppShell';
import { Login } from './pages/Login';
import { Dashboard } from './pages/Dashboard';
import { Sites } from './pages/Sites';
import { DesignSystem } from './pages/DesignSystem';
import { DesignSystemPreline } from './pages/DesignSystemPreline';
import { PlaceholderPage } from './components/skeletons';

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
      <Route path="/login" element={isAuthenticated ? <Navigate to="/" replace /> : <Login />} />

      <Route
        path="/"
        element={
          <ProtectedRoute>
            <AppShell />
          </ProtectedRoute>
        }
      >
        <Route index element={<Dashboard />} />
        <Route path="sites" element={<Sites />} />
        <Route path="analytics" element={<PlaceholderPage title="Analytics" />} />
        <Route path="redirects" element={<PlaceholderPage title="Redirects" />} />
        <Route path="webhooks" element={<PlaceholderPage title="Webhooks" />} />
        <Route path="logs" element={<PlaceholderPage title="Logs" />} />
        <Route path="settings" element={<PlaceholderPage title="Settings" />} />
        <Route path="design-system" element={<DesignSystem />} />
        <Route path="design-system-preline" element={<DesignSystemPreline />} />
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
