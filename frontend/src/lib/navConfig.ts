import type { LucideIcon } from 'lucide-react';
import {
  Activity,
  BarChart3,
  Bell,
  BookOpen,
  Cpu,
  Database,
  Flame,
  GitBranch,
  Globe,
  Gauge,
  LayoutGrid,
  Layers,
  Route,
  Server,
  Settings,
  Share2,
  Shield,
  Store,
  Terminal,
  Workflow,
} from 'lucide-react';

export interface NavItem {
  id: string;
  to: string;
  label: string;
  icon: LucideIcon;
  keywords?: string[];
  featureFlag?: string;
}

export interface NavSection {
  id: string;
  label: string;
  items: NavItem[];
}

export const NAV_SECTIONS: NavSection[] = [
  {
    id: 'observe',
    label: 'Observe',
    items: [
      { id: 'dashboard', to: '/', label: 'Command Center', icon: LayoutGrid, keywords: ['home', 'overview'] },
      { id: 'logs', to: '/logs', label: 'Log Explorer', icon: Terminal, keywords: ['search'] },
      { id: 'traces', to: '/traces', label: 'Trace Explorer', icon: GitBranch, keywords: ['apm'] },
      { id: 'service-flow', to: '/service-flow', label: 'Service Flow', icon: Route },
      { id: 'metrics', to: '/metrics', label: 'Metrics', icon: BarChart3 },
      { id: 'dashboards', to: '/dashboards', label: 'Dashboards', icon: Layers },
      { id: 'service-map', to: '/service-map', label: 'Service Map', icon: Share2, keywords: ['topology'] },
      { id: 'transactions', to: '/transactions', label: 'Transactions', icon: Route, keywords: ['upi'] },
      { id: 'anomalies', to: '/anomalies', label: 'Anomalies', icon: Activity },
    ],
  },
  {
    id: 'respond',
    label: 'Respond',
    items: [
      { id: 'incidents', to: '/incidents', label: 'Incidents', icon: Flame, keywords: ['p1', 'outage'] },
      { id: 'alerts', to: '/alerts', label: 'Alerts', icon: Bell },
      { id: 'slos', to: '/slos', label: 'SLOs', icon: Gauge },
    ],
  },
  {
    id: 'platform',
    label: 'Platform',
    items: [
      { id: 'infrastructure', to: '/infrastructure', label: 'Infrastructure', icon: Server },
      { id: 'kubernetes', to: '/kubernetes', label: 'Kubernetes', icon: Server, featureFlag: 'KUBERNETES' },
      { id: 'databases', to: '/databases', label: 'Databases', icon: Database },
      { id: 'middleware', to: '/middleware', label: 'Middleware', icon: Layers },
      { id: 'rum', to: '/rum', label: 'RUM', icon: Globe, featureFlag: 'RUM' },
      { id: 'synthetic', to: '/synthetic', label: 'Synthetic', icon: Activity, featureFlag: 'SYNTHETIC' },
      { id: 'cloud', to: '/cloud', label: 'Cloud', icon: Server },
      { id: 'security', to: '/security', label: 'Security', icon: Shield },
    ],
  },
  {
    id: 'automate',
    label: 'Automate',
    items: [
      { id: 'ai-chat', to: '/ai-chat', label: 'AI Assistant', icon: Cpu, keywords: ['copilot', 'chat'] },
      { id: 'workflows', to: '/workflows', label: 'Workflows', icon: Workflow },
      { id: 'notebooks', to: '/notebooks', label: 'Notebooks', icon: BookOpen },
    ],
  },
  {
    id: 'admin',
    label: 'Admin',
    items: [
      { id: 'marketplace', to: '/marketplace', label: 'Marketplace', icon: Store },
      { id: 'integrations', to: '/integrations', label: 'Integrations', icon: Share2 },
      { id: 'settings', to: '/settings', label: 'Settings', icon: Settings },
    ],
  },
];

export const ALL_NAV_ITEMS: NavItem[] = NAV_SECTIONS.flatMap((s) => s.items);

export function isNavActive(pathname: string, to: string): boolean {
  if (to === '/') return pathname === '/';
  return pathname === to || pathname.startsWith(`${to}/`);
}
