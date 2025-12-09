import { createContext, useContext, useEffect, useState } from 'react';
import type { ReactNode } from 'react';

interface MockContextValue {
  enabled: boolean;
  loading: boolean;
  setEnabled: (enabled: boolean) => void;
}

const MockContext = createContext<MockContextValue | undefined>(undefined);

export function MockProvider({ children }: { children: ReactNode }) {
  const [loading, setLoading] = useState(true);
  const [enabled, setEnabled] = useState(() => {
    // Check URL param
    const params = new URLSearchParams(window.location.search);
    if (params.has('mock-data')) {
      return true;
    }
    // Check localStorage
    const stored = localStorage.getItem('mockMode');
    return stored === 'true';
  });

  useEffect(() => {
    // Simulate loading delay
    const timer = setTimeout(() => {
      setLoading(false);
    }, 500);

    localStorage.setItem('mockMode', String(enabled));

    // Expose to window for debugging
    (window as any).mockMode = {
      enabled,
      loading,
      enable: () => setEnabled(true),
      disable: () => setEnabled(false),
    };

    if (enabled) {
      console.log('[Mock Mode] Enabled - All API calls will use mock data');
    }

    return () => clearTimeout(timer);
  }, [enabled, loading]);

  return (
    <MockContext.Provider value={{ enabled, loading, setEnabled }}>
      {children}
    </MockContext.Provider>
  );
}

export function useMockMode() {
  const context = useContext(MockContext);
  if (!context) {
    throw new Error('useMockMode must be used within MockProvider');
  }
  return context;
}
