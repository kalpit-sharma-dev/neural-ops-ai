import { useState } from 'react';
import { useFilterStore, type Environment, type TimeRangePreset } from '../../store/filterStore';
import { useAuthStore } from '../../store/authStore';
import { useRealtimeStore } from '../../store/realtimeStore';
import { logoutSession } from '../../api/auth';
import { SearchInput } from '../ui/SearchInput';
import { Select } from '../ui/Select';
import { Bell, LogOut, Menu, Monitor, Moon, Network, Sun } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import toast from 'react-hot-toast';
import { useThemeStore, type ThemePreference } from '../../store/themeStore';
import { useCommandPaletteStore } from '../../store/commandPaletteStore';
import { useSidebarStore } from './Sidebar';
import { MobileFilterSheet } from './MobileFilterSheet';

const THEME_LABEL: Record<ThemePreference, string> = {
  light: 'Light',
  dark: 'Dark',
  system: 'System',
};

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
  const preference = useThemeStore((s) => s.preference);
  const cyclePreference = useThemeStore((s) => s.cyclePreference);
  const setPaletteOpen = useCommandPaletteStore((s) => s.setOpen);
  const setMobileNav = useSidebarStore((s) => s.setMobileOpen);
  const [filterSheetOpen, setFilterSheetOpen] = useState(false);

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

  const openPalette = () => setPaletteOpen(true);

  return (
    <>
      <header className="topbar">
        <button
          type="button"
          className="topbar__menu icon-btn"
          aria-label="Open navigation"
          onClick={() => setMobileNav(true)}
        >
          <Menu size={20} />
        </button>

        <div className="topbar__brand">
          <Network size={22} className="topbar__logo" aria-hidden />
          <div>
            <p className="topbar__title">NeuralOps</p>
            <p className="topbar__subtitle">AI Observability</p>
          </div>
        </div>

        <div className="topbar__controls topbar__controls--desktop">
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
          <SearchInput
            placeholder="Search… (⌘K)"
            className="topbar__search"
            readOnly
            onFocus={openPalette}
            onClick={openPalette}
            aria-label="Open command palette"
          />
        </div>

        <button
          type="button"
          className="topbar__filters-mobile"
          onClick={() => setFilterSheetOpen(true)}
          aria-label="Open filters"
        >
          Filters
        </button>

        <div className="topbar__actions">
          <button
            type="button"
            className="theme-toggle"
            onClick={cyclePreference}
            aria-label={`Theme: ${THEME_LABEL[preference]}. Click to cycle.`}
            title={`Theme: ${THEME_LABEL[preference]}`}
          >
            {preference === 'light' && <Sun size={18} />}
            {preference === 'dark' && <Moon size={18} />}
            {preference === 'system' && <Monitor size={18} />}
          </button>
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
      <MobileFilterSheet open={filterSheetOpen} onOpenChange={setFilterSheetOpen} />
    </>
  );
}
