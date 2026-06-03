import { useEffect } from 'react';
import { Link, useRouterState } from '@tanstack/react-router';
import { ChevronDown, ChevronLeft, ChevronRight, Star } from 'lucide-react';
import { create } from 'zustand';
import { ALL_NAV_ITEMS, isNavActive, NAV_SECTIONS, type NavItem } from '../../lib/navConfig';
import { filterNavByFeature } from '../../lib/featureFlags';
import { useI18n } from '../../i18n/I18nProvider';
import { useSidebarPrefsStore } from '../../store/sidebarPrefsStore';

interface SidebarState {
  collapsed: boolean;
  mobileOpen: boolean;
  toggle: () => void;
  setMobileOpen: (open: boolean) => void;
}

export const useSidebarStore = create<SidebarState>((set) => ({
  collapsed: false,
  mobileOpen: false,
  toggle: () => set((s) => ({ collapsed: !s.collapsed })),
  setMobileOpen: (mobileOpen) => set({ mobileOpen }),
}));

function NavLink({
  item,
  collapsed,
  pathname,
  showPin,
}: {
  item: NavItem;
  collapsed: boolean;
  pathname: string;
  showPin?: boolean;
}) {
  const active = isNavActive(pathname, item.to);
  const toggleFavorite = useSidebarPrefsStore((s) => s.toggleFavorite);
  const hasFavorite = useSidebarPrefsStore((s) => s.hasFavorite);
  const Icon = item.icon;
  const { t } = useI18n();
  const label = t(`nav.${item.id}`) !== `nav.${item.id}` ? t(`nav.${item.id}`) : item.label;

  return (
    <div className="sidebar__link-row">
      <Link
        to={item.to}
        className={`sidebar__link ${active ? 'sidebar__link--active' : ''}`}
        title={label}
        aria-current={active ? 'page' : undefined}
        onClick={() => useSidebarStore.getState().setMobileOpen(false)}
      >
        <Icon size={18} aria-hidden />
        {!collapsed && <span>{label}</span>}
      </Link>
      {showPin && !collapsed && (
        <button
          type="button"
          className={`sidebar__pin ${hasFavorite(item.id) ? 'sidebar__pin--active' : ''}`}
          aria-label={hasFavorite(item.id) ? 'Unpin from favorites' : 'Pin to favorites'}
          onClick={(e) => {
            e.preventDefault();
            toggleFavorite(item.id);
          }}
        >
          <Star size={14} fill={hasFavorite(item.id) ? 'currentColor' : 'none'} />
        </button>
      )}
    </div>
  );
}

export function Sidebar() {
  const { t } = useI18n();
  const collapsed = useSidebarStore((s) => s.collapsed);
  const mobileOpen = useSidebarStore((s) => s.mobileOpen);
  const toggle = useSidebarStore((s) => s.toggle);
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const favorites = useSidebarPrefsStore((s) => s.favorites);
  const toggleSection = useSidebarPrefsStore((s) => s.toggleSection);
  const isSectionCollapsed = useSidebarPrefsStore((s) => s.isSectionCollapsed);

  const favoriteItems = favorites
    .map((id) => ALL_NAV_ITEMS.find((item) => item.id === id))
    .filter((item): item is NavItem => Boolean(item));

  useEffect(() => {
    document.documentElement.style.setProperty(
      '--sidebar-current-width',
      collapsed ? 'var(--sidebar-collapsed-width)' : 'var(--sidebar-width)',
    );
  }, [collapsed]);

  return (
    <>
      {mobileOpen && (
        <button
          type="button"
          className="sidebar-backdrop"
          aria-label="Close navigation"
          onClick={() => useSidebarStore.getState().setMobileOpen(false)}
        />
      )}
      <aside
        className={`sidebar ${collapsed ? 'sidebar--collapsed' : ''} ${mobileOpen ? 'sidebar--mobile-open' : ''}`}
      >
        <div className="sidebar__header">
          {!collapsed && (
            <div className="sidebar__brand">
              <span className="sidebar__brand-mark">N</span>
              <span className="sidebar__brand-text">NeuralOps</span>
            </div>
          )}
          <button type="button" className="sidebar__toggle" onClick={toggle} aria-label="Toggle sidebar width">
            {collapsed ? <ChevronRight size={18} /> : <ChevronLeft size={18} />}
          </button>
        </div>

        <nav className="sidebar__nav" aria-label="Main navigation">
          {favoriteItems.length > 0 && (
            <div className="sidebar__section">
              {!collapsed && <p className="sidebar__section-label">{t('section.favorites')}</p>}
              {favoriteItems.map((item) => (
                <NavLink key={`fav-${item.id}`} item={item} collapsed={collapsed} pathname={pathname} />
              ))}
            </div>
          )}

          {NAV_SECTIONS.map((section) => {
            const items = filterNavByFeature(section.items);
            if (items.length === 0) return null;
            const sectionCollapsed = isSectionCollapsed(section.id);
            const sectionLabel = t(`section.${section.id}`) !== `section.${section.id}` ? t(`section.${section.id}`) : section.label;
            return (
              <div key={section.id} className="sidebar__section">
                {!collapsed && (
                  <button
                    type="button"
                    className="sidebar__section-toggle"
                    aria-expanded={!sectionCollapsed}
                    onClick={() => toggleSection(section.id)}
                  >
                    <span>{sectionLabel}</span>
                    <ChevronDown size={14} className={sectionCollapsed ? 'sidebar__chevron--collapsed' : ''} />
                  </button>
                )}
                {(!sectionCollapsed || collapsed) &&
                  items.map((item) => (
                    <NavLink key={item.id} item={item} collapsed={collapsed} pathname={pathname} showPin />
                  ))}
              </div>
            );
          })}
        </nav>
      </aside>
    </>
  );
}
