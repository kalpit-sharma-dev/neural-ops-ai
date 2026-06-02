import { useState } from 'react';
import { useAuthStore } from '../../store/authStore';
import { useRealtimeStore } from '../../store/realtimeStore';
import { logoutSession } from '../../api/auth';
import { SearchInput } from '../ui/SearchInput';
import { Bell, LogOut, Menu, Monitor, Moon, Network, Sun } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import toast from 'react-hot-toast';
import { useThemeStore, type ThemePreference } from '../../store/themeStore';
import { useCommandPaletteStore } from '../../store/commandPaletteStore';
import { useSidebarStore } from './Sidebar';
import { MobileFilterSheet } from './MobileFilterSheet';
import { EnvPillBar } from './EnvPillBar';
import { TimeRangePicker } from './TimeRangePicker';

const THEME_LABEL: Record<ThemePreference, string> = {
  light: 'Light',
  dark: 'Dark',
  system: 'System',
};

export function TopBar() {
  const navigate = useNavigate();
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
          <EnvPillBar />
          <TimeRangePicker />
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
