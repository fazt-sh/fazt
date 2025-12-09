import { createContext, useContext, useState } from 'react';
import type { ReactNode } from 'react';
import type { User } from '../types/models';
import { api } from '../lib/api';
import { useMockMode } from './MockContext';
import { mockData } from '../lib/mockData';

interface AuthContextValue {
  user: User | null;
  isAuthenticated: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const { enabled: mockMode } = useMockMode();

  const login = async (username: string, password: string) => {
    if (mockMode) {
      // Mock login
      setUser(mockData.user);
      return;
    }

    const response = await api.post<{ user: User }>('/auth/login', {
      username,
      password,
    });
    setUser(response.user);
  };

  const logout = async () => {
    if (mockMode) {
      setUser(null);
      return;
    }

    await api.post('/auth/logout');
    setUser(null);
  };

  const checkAuth = async () => {
    if (mockMode) {
      setUser(mockData.user);
      return;
    }

    try {
      const response = await api.get<User>('/user/me');
      setUser(response);
    } catch (error) {
      setUser(null);
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        login,
        logout,
        checkAuth,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
