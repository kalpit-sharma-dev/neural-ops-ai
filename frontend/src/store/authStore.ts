import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { AuthConfig } from '../api/auth';

export interface AuthUser {
  id: string;
  email: string;
  name: string;
  role: string;
}

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  user: AuthUser | null;
  tenantId: string | null;
  tenantName: string;
  authEnabled: boolean;
  authConfig: AuthConfig | null;
  setTokens: (accessToken: string, refreshToken?: string | null) => void;
  setUser: (user: AuthUser, tenantId: string, tenantName: string) => void;
  setAuthConfig: (config: AuthConfig) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      user: null,
      tenantId: null,
      tenantName: '',
      authEnabled: false,
      authConfig: null,
      setTokens: (accessToken, refreshToken = null) => set({ accessToken, refreshToken }),
      setUser: (user, tenantId, tenantName) => set({ user, tenantId, tenantName }),
      setAuthConfig: (authConfig) =>
        set({
          authConfig,
          authEnabled: authConfig.authEnabled,
        }),
      logout: () =>
        set({
          accessToken: null,
          refreshToken: null,
          user: null,
          tenantId: null,
          tenantName: '',
        }),
    }),
    { name: 'neuralops-auth' },
  ),
);
