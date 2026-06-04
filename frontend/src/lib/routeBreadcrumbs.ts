import { ALL_NAV_ITEMS } from './navConfig';

export interface RouteCrumb {
  /** Human-readable label */
  label: string;
  /** Full pathname to navigate to (e.g. /logs/settings) */
  path: string;
  /** Single URL segment (for path trail) */
  segment: string;
  isCurrent?: boolean;
}

/** Exact paths that are not top-level nav items. */
const EXACT_PATH_LABELS: Record<string, string> = {
  '/logs/settings': 'Log settings',
  '/traces/settings': 'Sampling & retention',
  '/traces/compare': 'Trace compare',
  '/workflows/editor': 'Workflow editor',
  '/settings/platform': 'Platform',
  '/settings/users': 'Users',
  '/settings/api-keys': 'API keys',
  '/settings/audit': 'Audit log',
  '/settings/usage': 'Usage',
  '/settings/sso': 'SSO / IdP',
  '/settings/policies': 'Tenant policies',
  '/settings/oncall': 'On-call schedules',
  '/settings/alert-policies': 'Alert policies',
  '/settings/alert-suppressions': 'Suppressions',
  '/observability/materialization': 'Streaming materialization',
  '/security/attacks': 'Attacks',
  '/security/findings': 'Findings',
  '/rum/sessions': 'Sessions',
  '/entities': 'Entities',
  '/auth/callback': 'Auth callback',
};

const STATIC_SEGMENTS = new Set([
  'logs',
  'traces',
  'incidents',
  'alerts',
  'metrics',
  'dashboards',
  'settings',
  'security',
  'entities',
  'workflows',
  'kubernetes',
  'infrastructure',
  'middleware',
  'databases',
  'synthetic',
  'rum',
  'cloud',
  'slos',
  'reports',
  'integrations',
  'marketplace',
  'notebooks',
  'anomalies',
  'transactions',
  'service-map',
  'service-flow',
  'query-workbench',
  'collectors',
  'fleet',
  'observability',
  'materialization',
  'enterprise-governance',
  'business-observability',
  'cloud-finops-network',
  'nfr-certification',
  'ai-ops',
  'ai-chat',
  'design-system',
  'health',
  'login',
  'compare',
  'editor',
  'platform',
  'users',
  'api-keys',
  'audit',
  'usage',
  'sso',
  'policies',
  'oncall',
  'alert-policies',
  'alert-suppressions',
  'attacks',
  'findings',
  'sessions',
  'replay',
]);

function navLabelForExactPath(path: string): string | undefined {
  if (path === '/') return 'Command Center';
  const match = ALL_NAV_ITEMS.filter((item) => item.to === path).sort((a, b) => b.to.length - a.to.length)[0];
  return match?.label;
}

function longestNavPrefix(path: string): string | undefined {
  const matches = ALL_NAV_ITEMS.filter(
    (item) => item.to !== '/' && (path === item.to || path.startsWith(`${item.to}/`)),
  );
  if (!matches.length) return undefined;
  matches.sort((a, b) => b.to.length - a.to.length);
  return matches[0]?.label;
}

function truncateId(value: string, max = 20): string {
  if (value.length <= max) return value;
  return `${value.slice(0, 8)}…${value.slice(-4)}`;
}

function formatSegment(seg: string): string {
  return seg
    .split('-')
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ');
}

function dynamicLabel(
  segment: string,
  segments: string[],
  index: number,
  params: Record<string, string | undefined>,
): string {
  const parent = segments[index - 1];
  const grandparent = segments[index - 2];

  if (parent === 'incidents') return `Incident ${truncateId(segment)}`;
  if (parent === 'traces' && segment !== 'settings' && segment !== 'compare') {
    return `Trace ${truncateId(segment)}`;
  }
  if (parent === 'dashboards') return `Dashboard ${truncateId(segment)}`;
  if (parent === 'attacks') return `Attack ${truncateId(segment)}`;
  if (parent === 'findings') return `Finding ${truncateId(segment)}`;
  if (parent === 'entities') return 'Entities';
  if (grandparent === 'entities' && parent) {
    const entityId = params.id ?? segment;
    return truncateId(entityId);
  }
  if (segments[index - 2] === 'entities') return formatSegment(segment);
  if (parent === 'sessions' && grandparent === 'rum') return `Session ${truncateId(segment)}`;
  if (segment === 'replay') return 'Replay';
  if (parent === 'workflows' && segment === 'editor') return 'Editor';
  if (params.id === segment) return truncateId(segment);
  if (params.traceId === segment) return `Trace ${truncateId(segment)}`;
  if (params.sessionId === segment) return `Session ${truncateId(segment)}`;
  if (params.type === segment) return formatSegment(segment);

  return truncateId(segment);
}

function isDynamicSegment(
  segment: string,
  segments: string[],
  index: number,
  params: Record<string, string | undefined>,
): boolean {
  if (STATIC_SEGMENTS.has(segment)) return false;
  if (segment === 'editor' || segment === 'replay' || segment === 'compare') return false;
  if (Object.values(params).some((v) => v === segment)) return true;
  const parent = segments[index - 1];
  if (['incidents', 'dashboards'].includes(parent)) return true;
  if (parent === 'traces' && segment !== 'settings') return true;
  if (parent === 'attacks' || parent === 'findings') return true;
  if (parent === 'sessions') return true;
  if (segments[index - 2] === 'entities') return true;
  return !STATIC_SEGMENTS.has(segment) && segment.length > 0;
}

function labelForPath(
  path: string,
  segments: string[],
  segmentIndex: number,
  params: Record<string, string | undefined>,
): string {
  if (EXACT_PATH_LABELS[path]) return EXACT_PATH_LABELS[path];
  const exactNav = navLabelForExactPath(path);
  if (exactNav) return exactNav;

  const segment = segments[segmentIndex];
  if (isDynamicSegment(segment, segments, segmentIndex, params)) {
    return dynamicLabel(segment, segments, segmentIndex, params);
  }

  const prefixLabel = longestNavPrefix(path);
  if (prefixLabel && path.endsWith(segment) && segmentIndex === segments.length - 1) {
    const navItem = ALL_NAV_ITEMS.find((item) => path === item.to || path.startsWith(`${item.to}/`));
    if (navItem && path === navItem.to) return navItem.label;
  }

  if (segment === 'settings' && segments[segmentIndex - 1] === 'logs') return 'Log settings';
  if (segment === 'settings' && segments[segmentIndex - 1] === 'traces') return 'Sampling & retention';

  return formatSegment(segment);
}

/**
 * Builds a full clickable trail from root (/) through the current pathname.
 */
export function buildRouteCrumbs(
  pathname: string,
  params: Record<string, string | undefined> = {},
): RouteCrumb[] {
  const path = pathname.split('?')[0] || '/';
  if (path === '/') {
    return [{ label: 'Command Center', path: '/', segment: '', isCurrent: true }];
  }

  const segments = path.split('/').filter(Boolean);
  const crumbs: RouteCrumb[] = [{ label: 'Command Center', path: '/', segment: '' }];

  let acc = '';
  for (let i = 0; i < segments.length; i++) {
    const segment = segments[i];
    acc += `/${segment}`;
    const isCurrent = i === segments.length - 1;
    crumbs.push({
      label: labelForPath(acc, segments, i, params),
      path: acc,
      segment,
      isCurrent,
    });
  }

  crumbs[crumbs.length - 1].isCurrent = true;
  return crumbs;
}

/** Pages that should not show the global navigation chrome. */
export function shouldHidePageNavigation(pathname: string): boolean {
  return pathname === '/login' || pathname.startsWith('/auth/');
}

/** Primary nav destinations (sidebar roots) — no back bar on these. */
export function isTopLevelScreen(pathname: string): boolean {
  const path = pathname.split('?')[0] || '/';
  if (path === '/') return true;
  return ALL_NAV_ITEMS.some((item) => item.to === path);
}

/** Nested or detail routes where back + labeled breadcrumbs apply. */
export function isSubScreen(pathname: string): boolean {
  if (shouldHidePageNavigation(pathname)) return false;
  return !isTopLevelScreen(pathname);
}
