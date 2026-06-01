import { useEffect } from 'react';
import { RouterProvider } from '@tanstack/react-router';
import { fetchAuthConfig, fetchMe } from './api/auth';
import { useAuthStore } from './store/authStore';
import { router } from './router';

function AuthBootstrap({ children }: { children: React.ReactNode }) {
  const setAuthConfig = useAuthStore((s) => s.setAuthConfig);
  const setUser = useAuthStore((s) => s.setUser);
  const accessToken = useAuthStore((s) => s.accessToken);

  useEffect(() => {
    void fetchAuthConfig().then(setAuthConfig).catch(() => undefined);
  }, [setAuthConfig]);

  useEffect(() => {
    if (!accessToken) return;
    void fetchMe()
      .then((profile) => setUser(profile.user, profile.tenantId, profile.tenantName))
      .catch(() => useAuthStore.getState().logout());
  }, [accessToken, setUser]);

  return children;
}

export default function App() {
  return (
    <AuthBootstrap>
      <RouterProvider router={router} />
    </AuthBootstrap>
  );
}
