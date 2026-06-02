import type { ReactNode } from 'react';
import { Navigate, useLocation } from '@tanstack/react-router';
import { useAuthStore } from '../store/authStore';

interface ProtectedRouteProps {
  children: ReactNode;
  authRequired?: boolean;
}

export function ProtectedRoute({ children, authRequired = false }: ProtectedRouteProps) {
  const accessToken = useAuthStore((s) => s.accessToken);
  const authEnabled = useAuthStore((s) => s.authEnabled);
  const location = useLocation();
  const onAuthRoute = location.pathname === '/login' || location.pathname === '/auth/callback';

  // Prevent self-redirect loops: the pathless authed shell can render while
  // auth routes are active, so never redirect when already on an auth route.
  if (authRequired && authEnabled && !accessToken && !onAuthRoute) {
    return <Navigate to="/login" search={{ redirect: location.pathname }} replace />;
  }

  return children;
}
