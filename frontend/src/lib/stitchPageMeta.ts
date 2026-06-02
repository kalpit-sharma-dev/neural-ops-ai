/**
 * Stitch screen manifest — maps app routes to their canonical Google Stitch
 * design screens (project "Prompt-Based Design Generator").
 *
 * Source of truth for design parity work. Generated/curated from the Stitch MCP
 * inventory (docs/stitch-inventory.json) during the Phase 0 foundation pass.
 *
 * Preview URL pattern:
 *   https://stitch.withgoogle.com/preview/{STITCH_PROJECT_ID}?node-id={screenId}
 */

export const STITCH_PROJECT_ID = '37487902246048120';

export type ShellLayout = 'shell' | 'focus' | 'auth';

/** Implementation phase, aligned with the Stitch implementation plan. */
export type StitchPhase = 1 | 2 | 3 | 4 | 5 | 6 | 8;

export interface StitchScreenMeta {
  /** App route path (TanStack Router). */
  route: string;
  /** React page component name. */
  component: string;
  /** Human-readable Stitch screen title. */
  title: string;
  /** Canonical Stitch screen id (node-id), or null when no design exists yet. */
  screenId: string | null;
  /** Shell layout the route renders in. */
  layout: ShellLayout;
  /** Implementation phase. */
  phase: StitchPhase;
  /** Optional light-theme (Lumen) screen id. */
  lumenScreenId?: string;
  /** Notes about gaps or special handling. */
  note?: string;
}

export const STITCH_SCREENS: StitchScreenMeta[] = [
  // —— Auth & shell ——
  {
    route: '/login',
    component: 'Login',
    title: 'NeuralOps - Login',
    screenId: 'a864750f547e413bb5b63d4e3a52a181',
    lumenScreenId: 'e3de9f36718148108b9746e2c5207aab',
    layout: 'auth',
    phase: 1,
  },
  {
    route: '/auth/callback',
    component: 'AuthCallback',
    title: 'SSO redirect',
    screenId: null,
    layout: 'auth',
    phase: 1,
    note: 'No dedicated Stitch screen; loading/redirect state only.',
  },

  // —— Observe ——
  {
    route: '/',
    component: 'Dashboard',
    title: 'NeuralOps - Command Center',
    screenId: '9219f212f70f4257a2d51df33852d7c1',
    lumenScreenId: 'c061bc55ec31437ebc3057b922136d8d',
    layout: 'shell',
    phase: 1,
    note: 'Canonical shell reference for sidebar, top bar, env tabs.',
  },
  {
    route: '/logs',
    component: 'LogExplorer',
    title: 'NeuralOps - Log Explorer',
    screenId: '0cd0afabbace4bfebac53ffee97b147f',
    lumenScreenId: '7d3f5202f5ac47bfb16e41b7247da6ad',
    layout: 'shell',
    phase: 2,
  },
  {
    route: '/logs/settings',
    component: 'LogSettings',
    title: 'NeuralOps - Log Settings',
    screenId: 'ee61d2d640ab4455b95cba58e189704e',
    layout: 'shell',
    phase: 4,
  },
  {
    route: '/traces',
    component: 'TraceExplorer',
    title: 'NeuralOps - Trace Explorer',
    screenId: 'c7cf6a4896f440d1aca3628b88348c28',
    layout: 'shell',
    phase: 2,
    note: 'Use desktop frame, not mobile 6962e2df…',
  },
  {
    route: '/traces/$traceId',
    component: 'TraceDetail',
    title: 'NeuralOps - Trace Detail (UPI Failure)',
    screenId: '55f3699c9df64dc8b9f21777cf244bc0',
    layout: 'shell',
    phase: 4,
  },
  {
    route: '/traces/compare',
    component: 'TraceCompare',
    title: 'NeuralOps - Trace Compare Analysis',
    screenId: 'ecb792d4c3d14c2a9d24d181a3b71917',
    layout: 'shell',
    phase: 4,
  },
  {
    route: '/traces/settings',
    component: 'TraceSettings',
    title: 'NeuralOps - Trace Sampling & Retention',
    screenId: '12c195b47bc74906a27b4af32722015c',
    layout: 'shell',
    phase: 4,
  },
  {
    route: '/service-flow',
    component: 'ServiceFlow',
    title: 'NeuralOps - Service Flow',
    screenId: '7c7533d543bf411d8a1410c099990804',
    layout: 'shell',
    phase: 3,
  },
  {
    route: '/metrics',
    component: 'MetricsExplorer',
    title: 'NeuralOps - Metrics Explorer',
    screenId: '1637ab0178844d77a570950f338289e1',
    layout: 'shell',
    phase: 2,
  },
  {
    route: '/dashboards',
    component: 'Dashboards',
    title: 'NeuralOps - Dashboards List',
    screenId: 'fffd2350ccb7482f9ab644f1f63c691e',
    lumenScreenId: '8375c0e83c5b44739fa28682b4763a74',
    layout: 'shell',
    phase: 3,
  },
  {
    route: '/dashboards/$id',
    component: 'DashboardView',
    title: 'NeuralOps - Dashboard Editor',
    screenId: 'c770f81cf1284b2e8cb99d99f76bc734',
    layout: 'shell',
    phase: 3,
  },
  {
    route: '/service-map',
    component: 'ServiceMap',
    title: 'NeuralOps - Service Map',
    screenId: '030a5837ecb84808bac323d7ec92029a',
    layout: 'shell',
    phase: 2,
  },
  {
    route: '/entities/$type/$id',
    component: 'EntityPage',
    title: 'NeuralOps - Entity: payment-api',
    screenId: '9b9895c9fa2c4415a69d653f72fb3e1f',
    layout: 'shell',
    phase: 4,
  },
  {
    route: '/transactions',
    component: 'TransactionJourney',
    title: 'NeuralOps - Transaction Journey',
    screenId: 'da8a743169614b0788ee1351fb563b85',
    layout: 'shell',
    phase: 3,
  },
  {
    route: '/anomalies',
    component: 'AnomalyDetection',
    title: 'NeuralOps - Anomaly Detection Dashboard',
    screenId: '4c4064947cf54909be39bf166c24d446',
    layout: 'shell',
    phase: 3,
  },

  // —— Respond ——
  {
    route: '/incidents',
    component: 'Incidents',
    title: 'NeuralOps - Incidents List',
    screenId: 'c2d41efbe7bf4e0597afb9e7a0bbd1dd',
    lumenScreenId: 'f5f61895c0504078babb668bb6e73011',
    layout: 'shell',
    phase: 3,
  },
  {
    route: '/incidents/$id',
    component: 'IncidentDetail',
    title: 'NeuralOps - Incident Detail',
    screenId: '6f46e7ed41e14e6da842d05619cc6e4a',
    layout: 'focus',
    phase: 2,
  },
  {
    route: '/alerts',
    component: 'Alerts',
    title: 'NeuralOps - Alerts',
    screenId: '6cc6a9e5a64b41c681d650117ef1c022',
    layout: 'shell',
    phase: 2,
  },
  {
    route: '/slos',
    component: 'SLOs',
    title: 'NeuralOps - SLO Management',
    screenId: 'df42bd8ec62540f4ab653562486b0054',
    lumenScreenId: 'f9d21d0a4e4347b1b9d4699c42b568b5',
    layout: 'shell',
    phase: 3,
  },

  // —— Platform ——
  {
    route: '/infrastructure',
    component: 'Infrastructure',
    title: 'NeuralOps - Infrastructure & Kubernetes',
    screenId: 'ecc529cb841540489bc2d8dbebf640c5',
    layout: 'shell',
    phase: 5,
    note: 'Combined artboard; split infra vs k8s in code.',
  },
  {
    route: '/kubernetes',
    component: 'Kubernetes',
    title: 'NeuralOps - Kubernetes Cluster Detail',
    screenId: 'dfc0e3de9c6f43ea9c2f64a9f680bffb',
    layout: 'shell',
    phase: 5,
  },
  {
    route: '/databases',
    component: 'Databases',
    title: 'NeuralOps - Databases Platform',
    screenId: 'fb0664d7c7cd4efc8ce2c879c1e4c72a',
    layout: 'shell',
    phase: 4,
  },
  {
    route: '/middleware',
    component: 'Middleware',
    title: 'NeuralOps - Middleware Monitoring',
    screenId: 'ba24577829d741a4ae422e4d38f86314',
    layout: 'shell',
    phase: 4,
  },
  {
    route: '/rum',
    component: 'RUM',
    title: 'NeuralOps - Real User Monitoring (RUM)',
    screenId: 'd201f47f19d94ada8406c28bd972153b',
    layout: 'shell',
    phase: 6,
  },
  {
    route: '/rum/sessions/$sessionId/replay',
    component: 'SessionReplay',
    title: 'RUM - Session Replay',
    screenId: null,
    layout: 'shell',
    phase: 6,
    note: 'Gap: derive from RUM artboard or generate §23 replay frame.',
  },
  {
    route: '/synthetic',
    component: 'Synthetic',
    title: 'NeuralOps - Synthetic Monitoring',
    screenId: '3c6d7ae4ea3b4f07aa5f971befdcfca4',
    layout: 'shell',
    phase: 6,
  },
  {
    route: '/cloud',
    component: 'CloudMonitoring',
    title: 'NeuralOps - Cloud Monitoring',
    screenId: 'ca888946970142af8c1bedcf53427236',
    layout: 'shell',
    phase: 4,
  },
  {
    route: '/security',
    component: 'Security',
    title: 'NeuralOps - Security Monitoring',
    screenId: '1072b670f8004514b454252c3773831e',
    layout: 'shell',
    phase: 6,
  },
  {
    route: '/security/attacks/$id',
    component: 'SecurityAttackDetail',
    title: 'NeuralOps - Security Vulnerabilities',
    screenId: 'd663d7ec0fe64101ab2de492d0f2c591',
    layout: 'shell',
    phase: 6,
    note: 'Reference frame; attack detail drawer may need a dedicated screen.',
  },

  // —— Automate ——
  {
    route: '/ai-chat',
    component: 'AIChat',
    title: 'NeuralOps - AI Assistant',
    screenId: '4cb39d9f61af4865b830ef7e326105dd',
    layout: 'shell',
    phase: 2,
  },
  {
    route: '/workflows',
    component: 'Workflows',
    title: 'NeuralOps - Workflows (list)',
    screenId: null,
    layout: 'shell',
    phase: 3,
    note: 'Gap: only editor exists; reuse editor list pane or generate list artboard.',
  },
  {
    route: '/workflows/editor',
    component: 'WorkflowEditor',
    title: 'NeuralOps - Workflows Editor',
    screenId: 'd980e4d6b6ba4263b68aa8d3645f819e',
    layout: 'shell',
    phase: 3,
  },
  {
    route: '/notebooks',
    component: 'Notebooks',
    title: 'NeuralOps - Notebooks Analysis Workspace',
    screenId: '654c87d65f394766a4188fc74156f8d8',
    layout: 'shell',
    phase: 3,
  },

  // —— Admin ——
  {
    route: '/marketplace',
    component: 'Marketplace',
    title: 'NeuralOps - Marketplace Catalog',
    screenId: '4b2645ed11954e3abff2edcfe6f6995a',
    layout: 'shell',
    phase: 4,
  },
  {
    route: '/integrations',
    component: 'Integrations',
    title: 'NeuralOps - Integrations Marketplace',
    screenId: '680709399ebe445faffe57110bba766a',
    layout: 'shell',
    phase: 4,
  },
  {
    route: '/settings',
    component: 'Settings',
    title: 'NeuralOps - Settings Overview',
    screenId: 'acb97fbfccb14ea88336bf864d50304e',
    lumenScreenId: 'c435b6c0acc4470c95a9c7a2d53c1942',
    layout: 'shell',
    phase: 4,
  },
  {
    route: '/settings/users',
    component: 'SettingsUsers',
    title: 'NeuralOps - Settings: Users & Teams',
    screenId: '33fd57aa179c431284a344f69a98c13b',
    layout: 'shell',
    phase: 4,
  },
  {
    route: '/settings/api-keys',
    component: 'SettingsApiKeys',
    title: 'NeuralOps - Settings: API Keys',
    screenId: '48c7987f57ff448faeb8f94586a4fb89',
    layout: 'shell',
    phase: 3,
  },
  {
    route: '/settings/sso',
    component: 'SettingsSSO',
    title: 'NeuralOps - Settings: SSO',
    screenId: 'c1848c5c83ae4b2fbfb1c54914accec8',
    layout: 'shell',
    phase: 3,
  },
  {
    route: '/settings/audit',
    component: 'SettingsAudit',
    title: 'NeuralOps - Settings: Audit Log',
    screenId: '53230243da7146da81980911986a3209',
    layout: 'shell',
    phase: 4,
  },
  {
    route: '/settings/usage',
    component: 'SettingsUsage',
    title: 'NeuralOps - Settings: Usage & Billing',
    screenId: '680fa11081424956a375b81d49501e57',
    layout: 'shell',
    phase: 4,
  },
  {
    route: '/settings/policies',
    component: 'SettingsPolicies',
    title: 'NeuralOps - Settings: Tenant Policies',
    screenId: '2ed396f6018c47838d0702d1e52208c8',
    layout: 'shell',
    phase: 4,
  },
  {
    route: '/settings/oncall',
    component: 'SettingsOncall',
    title: 'NeuralOps - Settings: On-call Schedules',
    screenId: 'd70db942e4184ab9ba5f12643b44dc5d',
    layout: 'shell',
    phase: 4,
  },
  {
    route: '/design-system',
    component: 'DesignSystem',
    title: 'Design System',
    screenId: null,
    layout: 'shell',
    phase: 1,
    note: 'No Stitch screen; mirrors docs/stitch/DESIGN.md in a live page.',
  },
  {
    route: '/health',
    component: 'HealthPage',
    title: 'System Health',
    screenId: null,
    layout: 'shell',
    phase: 1,
    note: 'No Stitch screen; operational page.',
  },
];

const SCREENS_BY_ROUTE = new Map(STITCH_SCREENS.map((s) => [s.route, s]));

/** Build the Stitch preview URL for a screen id. */
export function stitchPreviewUrl(screenId: string): string {
  return `https://stitch.withgoogle.com/preview/${STITCH_PROJECT_ID}?node-id=${screenId}`;
}

/** Look up the Stitch screen meta for an exact route path. */
export function getStitchMeta(route: string): StitchScreenMeta | undefined {
  return SCREENS_BY_ROUTE.get(route);
}

/** Routes that still need a Stitch design (gaps to close before polish). */
export function getStitchGaps(): StitchScreenMeta[] {
  return STITCH_SCREENS.filter((s) => s.screenId === null);
}
