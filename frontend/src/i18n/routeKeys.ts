/** Maps app pathname to i18n page id (longest prefix wins). */
const ROUTE_TABLE: Array<{ prefix: string; id: string }> = [
  { prefix: '/observability/materialization', id: 'materialization' },
  { prefix: '/settings/platform', id: 'platform-about' },
  { prefix: '/settings/alert-suppressions', id: 'alert-suppressions' },
  { prefix: '/settings/alert-policies', id: 'alert-policies' },
  { prefix: '/settings/api-keys', id: 'settings-api-keys' },
  { prefix: '/settings/oncall', id: 'settings-oncall' },
  { prefix: '/settings/policies', id: 'settings-policies' },
  { prefix: '/settings/audit', id: 'settings-audit' },
  { prefix: '/settings/usage', id: 'settings-usage' },
  { prefix: '/settings/users', id: 'settings-users' },
  { prefix: '/settings/sso', id: 'settings-sso' },
  { prefix: '/logs/settings', id: 'log-settings' },
  { prefix: '/traces/settings', id: 'trace-settings' },
  { prefix: '/traces/compare', id: 'trace-compare' },
  { prefix: '/workflows/editor', id: 'workflow-editor' },
  { prefix: '/rum/sessions/', id: 'session-replay' },
  { prefix: '/security/findings/', id: 'security-finding' },
  { prefix: '/security/attacks/', id: 'security-attack' },
  { prefix: '/cloud-finops-network', id: 'cloud-finops-network' },
  { prefix: '/business-observability', id: 'business-observability' },
  { prefix: '/enterprise-governance', id: 'enterprise-governance' },
  { prefix: '/nfr-certification', id: 'nfr-certification' },
  { prefix: '/query-workbench', id: 'query-workbench' },
  { prefix: '/collectors/fleet', id: 'collectors-fleet' },
  { prefix: '/ai-ops', id: 'ai-ops' },
  { prefix: '/ai-chat', id: 'ai-chat' },
  { prefix: '/service-flow', id: 'service-flow' },
  { prefix: '/service-map', id: 'service-map' },
  { prefix: '/dashboards/', id: 'dashboard-view' },
  { prefix: '/dashboards', id: 'dashboards' },
  { prefix: '/incidents/', id: 'incident-detail' },
  { prefix: '/traces/', id: 'trace-detail' },
  { prefix: '/entities/', id: 'entity' },
  { prefix: '/transactions', id: 'transactions' },
  { prefix: '/anomalies', id: 'anomalies' },
  { prefix: '/infrastructure', id: 'infrastructure' },
  { prefix: '/kubernetes', id: 'kubernetes' },
  { prefix: '/middleware', id: 'middleware' },
  { prefix: '/databases', id: 'databases' },
  { prefix: '/synthetic', id: 'synthetic' },
  { prefix: '/marketplace', id: 'marketplace' },
  { prefix: '/integrations', id: 'integrations' },
  { prefix: '/workflows', id: 'workflows' },
  { prefix: '/notebooks', id: 'notebooks' },
  { prefix: '/metrics', id: 'metrics' },
  { prefix: '/incidents', id: 'incidents' },
  { prefix: '/security', id: 'security' },
  { prefix: '/settings', id: 'settings' },
  { prefix: '/traces', id: 'traces' },
  { prefix: '/alerts', id: 'alerts' },
  { prefix: '/logs', id: 'logs' },
  { prefix: '/cloud', id: 'cloud' },
  { prefix: '/rum', id: 'rum' },
  { prefix: '/slos', id: 'slos' },
  { prefix: '/design-system', id: 'design-system' },
  { prefix: '/health', id: 'health' },
  { prefix: '/login', id: 'login' },
  { prefix: '/', id: 'dashboard' },
];

export function resolvePageId(pathname: string): string | undefined {
  const path = pathname.split('?')[0] || '/';
  for (const { prefix, id } of ROUTE_TABLE) {
    if (prefix === '/') {
      if (path === '/') return id;
      continue;
    }
    if (path === prefix || path.startsWith(`${prefix}/`)) return id;
  }
  return undefined;
}

/** Nav id for sidebar (may differ from page id for settings sub-routes). */
export function pageTitleKey(pageId: string): string {
  const navAlias: Record<string, string> = {
    'log-settings': 'logs',
    'trace-settings': 'traces',
    'trace-compare': 'traces',
    'trace-detail': 'traces',
    'incident-detail': 'incidents',
    'dashboard-view': 'dashboards',
    'security-attack': 'security',
    'security-finding': 'security',
    'session-replay': 'rum',
    'workflow-editor': 'workflows',
    entity: 'service-map',
    'settings-users': 'settings',
    'settings-api-keys': 'settings',
    'settings-audit': 'settings',
    'settings-usage': 'settings',
    'settings-sso': 'settings',
    'settings-policies': 'settings',
    'settings-oncall': 'settings',
  };
  return navAlias[pageId] ?? pageId;
}
