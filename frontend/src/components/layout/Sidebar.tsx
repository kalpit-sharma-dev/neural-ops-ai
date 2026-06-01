import { useEffect } from 'react';
import { Link, useRouterState } from '@tanstack/react-router';
import {
  Activity,
  BarChart3,
  Bell,
  BookOpen,
  Cpu,
  Database,
  Flame,
  GitBranch,
  LayoutGrid,
  Layers,
  Route,
  Server,
  Settings,
  Share2,
  Shield,
  Terminal,
  Workflow,
  ChevronLeft,
  ChevronRight,
  Globe,
  Gauge,
  Store,
} from 'lucide-react';
import { create } from 'zustand';

interface SidebarState {
  collapsed: boolean;
  toggle: () => void;
}

export const useSidebarStore = create<SidebarState>((set) => ({
  collapsed: false,
  toggle: () => set((s) => ({ collapsed: !s.collapsed })),
}));

const navItems = [
  { to: '/', icon: LayoutGrid, label: 'Dashboard' },
  { to: '/incidents', icon: Flame, label: 'Incidents' },
  { to: '/logs', icon: Terminal, label: 'Log Explorer' },
  { to: '/traces', icon: GitBranch, label: 'Trace Explorer' },
  { to: '/service-flow', icon: Route, label: 'Service Flow' },
  { to: '/metrics', icon: BarChart3, label: 'Metrics' },
  { to: '/dashboards', icon: Layers, label: 'Dashboards' },
  { to: '/transactions', icon: Route, label: 'Transactions' },
  { to: '/anomalies', icon: Activity, label: 'Anomalies' },
  { to: '/slos', icon: Gauge, label: 'SLOs' },
  { to: '/service-map', icon: Share2, label: 'Service Map' },
  { to: '/infrastructure', icon: Server, label: 'Infrastructure' },
  { to: '/kubernetes', icon: Server, label: 'Kubernetes' },
  { to: '/databases', icon: Database, label: 'Databases' },
  { to: '/middleware', icon: Layers, label: 'Middleware' },
  { to: '/rum', icon: Globe, label: 'RUM' },
  { to: '/synthetic', icon: Activity, label: 'Synthetic' },
  { to: '/ai-chat', icon: Cpu, label: 'AI Assistant' },
  { to: '/cloud', icon: Server, label: 'Cloud' },
  { to: '/workflows', icon: Workflow, label: 'Workflows' },
  { to: '/notebooks', icon: BookOpen, label: 'Notebooks' },
  { to: '/security', icon: Shield, label: 'Security' },
  { to: '/marketplace', icon: Store, label: 'Marketplace' },
  { to: '/integrations', icon: Share2, label: 'Integrations' },
  { to: '/alerts', icon: Bell, label: 'Alerts' },
  { to: '/settings', icon: Settings, label: 'Settings' },
];

export function Sidebar() {
  const collapsed = useSidebarStore((s) => s.collapsed);
  const toggle = useSidebarStore((s) => s.toggle);
  const pathname = useRouterState({ select: (state) => state.location.pathname });

  useEffect(() => {
    document.documentElement.style.setProperty(
      '--sidebar-current-width',
      collapsed ? 'var(--sidebar-collapsed-width)' : 'var(--sidebar-width)',
    );
  }, [collapsed]);

  return (
    <aside className={`sidebar ${collapsed ? 'sidebar--collapsed' : ''}`}>
      <button type="button" className="sidebar__toggle" onClick={toggle} aria-label="Toggle sidebar">
        {collapsed ? <ChevronRight size={18} /> : <ChevronLeft size={18} />}
      </button>
      <nav className="sidebar__nav">
        {navItems.map(({ to, icon: Icon, label }) => {
          const active = to === '/' ? pathname === '/' : pathname === to || pathname.startsWith(`${to}/`);
          return (
            <Link
              key={to}
              to={to}
              className={`sidebar__link ${active ? 'sidebar__link--active' : ''}`}
              title={label}
            >
              <Icon size={18} />
              {!collapsed && <span>{label}</span>}
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}
