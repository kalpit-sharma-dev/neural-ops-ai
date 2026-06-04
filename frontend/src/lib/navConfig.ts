import type { LucideIcon } from 'lucide-react';
import {
  Activity,
  BarChart3,
  Bell,
  BookOpen,
  FileText,
  Cpu,
  Radar,
  Database,
  Flame,
  GitBranch,
  Globe,
  Gauge,
  LayoutGrid,
  Layers,
  SearchCode,
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
      {
        id: 'business-observability',
        to: '/business-observability',
        label: 'Business observability',
        icon: BarChart3,
        keywords: ['funnel', 'kpi', 'browser test'],
        featureFlag: 'BUSINESS_OBSERVABILITY',
      },
      { id: 'cloud', to: '/cloud', label: 'Cloud', icon: Server },
      {
        id: 'cloud-finops-network',
        to: '/cloud-finops-network',
        label: 'Cloud & Network',
        icon: Globe,
        keywords: ['finops', 'npm', 'assets', 'topology'],
        featureFlag: 'CLOUD_FINOPS',
      },
      {
        id: 'collectors-fleet',
        to: '/collectors/fleet',
        label: 'Collectors Fleet',
        icon: Radar,
        keywords: ['agent', 'pipeline', 'fleet'],
        featureFlag: 'COLLECTORS_FLEET',
      },
      { id: 'security', to: '/security', label: 'Security', icon: Shield },
    ],
  },
  {
    id: 'automate',
    label: 'Automate',
    items: [
      { id: 'ai-chat', to: '/ai-chat', label: 'AI Assistant', icon: Cpu, keywords: ['copilot', 'chat'] },
      { id: 'ai-ops', to: '/ai-ops', label: 'AI Ops Center', icon: Gauge, keywords: ['rca', 'forecast', 'autofix'], featureFlag: 'AI_OPS' },
      {
        id: 'query-workbench',
        to: '/query-workbench',
        label: 'Query Workbench',
        icon: SearchCode,
        keywords: ['nexql', 'query', 'cross-signal'],
        featureFlag: 'QUERY_WORKBENCH',
      },
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
      {
        id: 'enterprise-governance',
        to: '/enterprise-governance',
        label: 'Governance',
        icon: Shield,
        keywords: ['abac', 'msp', 'residency'],
        featureFlag: 'ENTERPRISE_GOVERNANCE',
      },
      {
        id: 'reports',
        to: '/reports',
        label: 'Reports & exports',
        icon: FileText,
        keywords: ['download', 'export', 'csv', 'json'],
      },
      {
        id: 'nfr-certification',
        to: '/nfr-certification',
        label: 'NFR certification',
        icon: Gauge,
        keywords: ['wcag', 'benchmark', 'dr'],
        featureFlag: 'NFR_CERTIFICATION',
      },
      {
        id: 'alert-policies',
        to: '/settings/alert-policies',
        label: 'Alert policies',
        icon: Bell,
        keywords: ['routing', 'escalation'],
        featureFlag: 'ALERT_POLICIES',
      },
      {
        id: 'alert-suppressions',
        to: '/settings/alert-suppressions',
        label: 'Suppressions',
        icon: Bell,
        keywords: ['silence', 'maintenance'],
        featureFlag: 'ALERT_POLICIES',
      },
      {
        id: 'materialization',
        to: '/observability/materialization',
        label: 'Streaming materialization',
        icon: Activity,
        keywords: ['kafka', 'derived', 'fatigue'],
        featureFlag: 'MATERIALIZATION',
      },
      { id: 'settings', to: '/settings', label: 'Settings', icon: Settings },
    ],
  },
];

export const ALL_NAV_ITEMS: NavItem[] = NAV_SECTIONS.flatMap((s) => s.items);

export function isNavActive(pathname: string, to: string): boolean {
  if (to === '/') return pathname === '/';
  return pathname === to || pathname.startsWith(`${to}/`);
}
