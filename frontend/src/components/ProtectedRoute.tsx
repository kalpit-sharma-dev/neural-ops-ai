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

  if (authRequired && authEnabled && !accessToken) {
    return <Navigate to="/login" search={{ redirect: location.pathname }} replace />;
  }

  return children;
}
