import { useEffect } from 'react';
import { useNavigate, useSearch } from '@tanstack/react-router';
import toast from 'react-hot-toast';
import { exchangeOIDCCode } from '../api/auth';
import { LoadingState } from '../components/ui/PageStates';
import { useAuthStore } from '../store/authStore';

export default function AuthCallback() {
  const search = useSearch({ strict: false }) as { code?: string };
  const navigate = useNavigate();
  const setTokens = useAuthStore((s) => s.setTokens);
  const setUser = useAuthStore((s) => s.setUser);

  useEffect(() => {
    const code = search.code;
    if (!code) {
      toast.error('Missing authorization code');
      navigate({ to: '/login', replace: true });
      return;
    }

    void exchangeOIDCCode(code)
      .then((session) => {
        setTokens(session.accessToken, session.refreshToken);
        if (session.user && session.tenantId) {
          setUser(session.user, session.tenantId, session.tenantName ?? session.tenantId);
        }
        toast.success('Signed in with SSO');
        navigate({ to: '/', replace: true });
      })
      .catch(() => {
        toast.error('Failed to complete sign-in');
        navigate({ to: '/login', replace: true });
      });
  }, [navigate, search.code, setTokens, setUser]);

  return <LoadingState label="Completing sign-in…" />;
}
