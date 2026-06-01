import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import toast from 'react-hot-toast';
import { devLogin, fetchAuthConfig, startOIDCLogin } from '../api/auth';
import { getApiErrorMessage } from '../api/client';
import { Button } from '../components/ui/Button';
import { useAuthStore } from '../store/authStore';

export default function Login() {
  const navigate = useNavigate();
  const setTokens = useAuthStore((s) => s.setTokens);
  const setUser = useAuthStore((s) => s.setUser);
  const authConfig = useAuthStore((s) => s.authConfig);
  const [email, setEmail] = useState('demo@neuralops.ai');
  const [loading, setLoading] = useState(false);

  const startSSO = async () => {
    setLoading(true);
    try {
      const { authorizationUrl } = await startOIDCLogin();
      window.location.href = authorizationUrl;
    } catch (error) {
      toast.error(getApiErrorMessage(error));
      setLoading(false);
    }
  };

  const loginDev = async () => {
    setLoading(true);
    try {
      const response = await devLogin(email, authConfig?.defaultTenant);
      setTokens(response.accessToken, response.refreshToken);
      if (response.user && response.tenantId) {
        setUser(response.user, response.tenantId, response.tenantName ?? 'Tenant');
      } else {
        const me = await fetchAuthConfig();
        setUser(
          { id: 'dev', email, name: email.split('@')[0], role: 'ADMIN' },
          me.defaultTenant,
          'NeuralOps Demo',
        );
      }
      toast.success('Signed in');
      navigate({ to: '/' });
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-page">
      <div className="login-card">
        <h1>NeuralOps</h1>
        <p className="muted">Sign in to your observability workspace</p>

        {authConfig?.oidcEnabled && (
          <Button variant="primary" onClick={() => void startSSO()} disabled={loading} style={{ width: '100%' }}>
            Continue with SSO
          </Button>
        )}

        {authConfig?.devLoginEnabled && (
          <div className="login-dev">
            <label htmlFor="email">Dev login email</label>
            <input
              id="email"
              className="filter-input"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="demo@neuralops.ai"
            />
            <Button variant="secondary" onClick={() => void loginDev()} disabled={loading} style={{ width: '100%' }}>
              Dev sign in
            </Button>
          </div>
        )}

        {!authConfig?.oidcEnabled && !authConfig?.devLoginEnabled && (
          <p className="muted">Authentication is disabled in this environment.</p>
        )}
      </div>
    </div>
  );
}
