import { useFilterStore, type Environment, type TimeRangePreset } from '../../store/filterStore';
import { useAuthStore } from '../../store/authStore';
import { useRealtimeStore } from '../../store/realtimeStore';
import { logoutSession } from '../../api/auth';
import { SearchInput } from '../ui/SearchInput';
import { Select } from '../ui/Select';
import { Bell, LogOut, Network } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import toast from 'react-hot-toast';

export function TopBar() {
  const navigate = useNavigate();
  const environment = useFilterStore((s) => s.environment);
  const timeRange = useFilterStore((s) => s.timeRange);
  const setEnvironment = useFilterStore((s) => s.setEnvironment);
  const setTimeRange = useFilterStore((s) => s.setTimeRange);
  const user = useAuthStore((s) => s.user);
  const tenantName = useAuthStore((s) => s.tenantName);
  const authEnabled = useAuthStore((s) => s.authEnabled);
  const refreshToken = useAuthStore((s) => s.refreshToken);
  const logout = useAuthStore((s) => s.logout);
  const unreadCount = useRealtimeStore((s) => s.unreadCount);
  const connected = useRealtimeStore((s) => s.connected);

  const handleLogout = () => {
    void logoutSession(refreshToken)
      .catch(() => undefined)
      .finally(() => {
        logout();
        toast.success('Signed out');
        if (authEnabled) {
          navigate({ to: '/login', replace: true });
        }
      });
  };

  return (
    <header className="topbar">
      <div className="topbar__brand">
        <Network size={22} className="topbar__logo" />
        <div>
          <p className="topbar__title">NeuralOps</p>
          <p className="topbar__subtitle">AI Observability</p>
        </div>
      </div>

      <div className="topbar__controls">
        <Select label="Env" value={environment} onChange={(e) => setEnvironment(e.target.value as Environment)}>
          <option value="ALL">ALL</option>
          <option value="PROD">PROD</option>
          <option value="STAGING">STAGING</option>
          <option value="DEV">DEV</option>
        </Select>
        <Select label="Range" value={timeRange} onChange={(e) => setTimeRange(e.target.value as TimeRangePreset)}>
          <option value="1h">Last 1h</option>
          <option value="6h">Last 6h</option>
          <option value="24h">Last 24h</option>
          <option value="7d">Last 7d</option>
        </Select>
        <SearchInput placeholder="Search… (⌘K)" className="topbar__search" onFocus={() => {}} />
      </div>

      <div className="topbar__actions">
        <button type="button" className="topbar__bell" onClick={() => navigate({ to: '/alerts' })} aria-label="Alerts">
          <Bell size={18} />
          {unreadCount > 0 && <span className="topbar__badge">{unreadCount}</span>}
        </button>
        <span className={`topbar__live ${connected ? 'connected' : ''}`} title={connected ? 'Live' : 'Offline'} />
        <div className="topbar__user">
          <span className="topbar__avatar">{user?.name?.charAt(0) ?? 'U'}</span>
          <div>
            <p className="topbar__username">{user?.name ?? 'User'}</p>
            <p className="topbar__tenant">{tenantName || 'Tenant'}</p>
          </div>
        </div>
        {authEnabled && (
          <button type="button" className="topbar__logout" onClick={handleLogout} aria-label="Sign out">
            <LogOut size={18} />
          </button>
        )}
      </div>
    </header>
  );
}
