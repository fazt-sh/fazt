import { createContext, useContext, useEffect, useState } from 'react';
import type { ReactNode } from 'react';

interface MockContextValue {
  enabled: boolean;
  setEnabled: (enabled: boolean) => void;
}

const MockContext = createContext<MockContextValue | undefined>(undefined);

export function MockProvider({ children }: { children: ReactNode }) {
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
    localStorage.setItem('mockMode', String(enabled));

    // Expose to window for debugging
    (window as any).mockMode = {
      enabled,
      enable: () => setEnabled(true),
      disable: () => setEnabled(false),
    };

    if (enabled) {
      console.log('[Mock Mode] Enabled - All API calls will use mock data');
    }
  }, [enabled]);

  return (
    <MockContext.Provider value={{ enabled, setEnabled }}>
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
