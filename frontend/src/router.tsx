import { createRootRoute, createRoute, createRouter, Outlet } from '@tanstack/react-router';
import { Layout } from './components/Layout';
import { ProtectedRoute } from './components/ProtectedRoute';
import { lazyPage } from './routes/lazyPage';
import { useAuthStore } from './store/authStore';

// Auth entry points stay eager so unauthenticated load is fast and flicker-free.
import Login from './pages/Login';
import AuthCallback from './pages/AuthCallback';

// All authenticated route pages are code-split so the initial bundle stays small
// and heavy dependencies (recharts, xyflow, rrweb) only load when their page does.
const Dashboard = lazyPage(() => import('./pages/Dashboard'));
const LogExplorer = lazyPage(() => import('./pages/LogExplorer'));
const LogSettings = lazyPage(() => import('./pages/LogSettings'));
const Incidents = lazyPage(() => import('./pages/Incidents'));
const IncidentDetail = lazyPage(() => import('./pages/IncidentDetail'));
const ServiceMap = lazyPage(() => import('./pages/ServiceMap'));
const RUM = lazyPage(() => import('./pages/RUM'));
const Kubernetes = lazyPage(() => import('./pages/Kubernetes'));
const Infrastructure = lazyPage(() => import('./pages/Infrastructure'));
const SessionReplay = lazyPage(() => import('./pages/SessionReplay'));
const AIChat = lazyPage(() => import('./pages/AIChat'));
const TransactionJourney = lazyPage(() => import('./pages/TransactionJourney'));
const DesignSystem = lazyPage(() => import('./pages/DesignSystem'));
const AnomalyDetection = lazyPage(() => import('./pages/AnomalyDetection'));
const TraceExplorer = lazyPage(() => import('./pages/TraceExplorer'));
const TraceDetail = lazyPage(() => import('./pages/TraceDetail'));
const ServiceFlow = lazyPage(() => import('./pages/ServiceFlow'));
const Alerts = lazyPage(() => import('./pages/Alerts'));
const Settings = lazyPage(() => import('./pages/Settings'));
const SettingsUsers = lazyPage(() => import('./pages/SettingsUsers'));
const SettingsApiKeys = lazyPage(() => import('./pages/SettingsApiKeys'));
const SettingsAudit = lazyPage(() => import('./pages/SettingsAudit'));
const SettingsUsage = lazyPage(() => import('./pages/SettingsUsage'));
const SettingsSSO = lazyPage(() => import('./pages/SettingsSSO'));
const SettingsPolicies = lazyPage(() => import('./pages/SettingsPolicies'));
const SettingsOncall = lazyPage(() => import('./pages/SettingsOncall'));
const MetricsExplorer = lazyPage(() => import('./pages/MetricsExplorer'));
const Dashboards = lazyPage(() => import('./pages/Dashboards'));
const DashboardView = lazyPage(() => import('./pages/DashboardView'));
const SLOs = lazyPage(() => import('./pages/SLOs'));
const Databases = lazyPage(() => import('./pages/Databases'));
const Middleware = lazyPage(() => import('./pages/Middleware'));
const Synthetic = lazyPage(() => import('./pages/Synthetic'));
const Workflows = lazyPage(() => import('./pages/Workflows'));
const Notebooks = lazyPage(() => import('./pages/Notebooks'));
const Security = lazyPage(() => import('./pages/Security'));
const SecurityAttackDetail = lazyPage(() => import('./pages/SecurityAttackDetail'));
const Marketplace = lazyPage(() => import('./pages/Marketplace'));
const TraceSettings = lazyPage(() => import('./pages/TraceSettings'));
const CloudMonitoring = lazyPage(() => import('./pages/CloudMonitoring'));
const WorkflowEditor = lazyPage(() => import('./pages/WorkflowEditor'));
const Integrations = lazyPage(() => import('./pages/Integrations'));
const TraceCompare = lazyPage(() => import('./pages/TraceCompare'));
const EntityPage = lazyPage(() => import('./pages/EntityPage'));
const HealthPage = lazyPage(() => import('./pages/HealthPage').then((m) => ({ default: m.HealthPage })));

function AuthedShell() {
  const authEnabled = useAuthStore((s) => s.authEnabled);
  return (
    <ProtectedRoute authRequired={authEnabled}>
      <Layout>
        <Outlet />
      </Layout>
    </ProtectedRoute>
  );
}

const rootRoute = createRootRoute({ component: () => <Outlet /> });

const loginRoute = createRoute({ getParentRoute: () => rootRoute, path: '/login', component: Login });
const authCallbackRoute = createRoute({ getParentRoute: () => rootRoute, path: '/auth/callback', component: AuthCallback });

const authedRoute = createRoute({ getParentRoute: () => rootRoute, id: 'authed', component: AuthedShell });

const indexRoute = createRoute({ getParentRoute: () => authedRoute, path: '/', component: Dashboard });
const incidentsRoute = createRoute({ getParentRoute: () => authedRoute, path: '/incidents', component: Incidents });
const incidentDetailRoute = createRoute({ getParentRoute: () => authedRoute, path: '/incidents/$id', component: IncidentDetail });
const logsRoute = createRoute({
  getParentRoute: () => authedRoute,
  path: '/logs',
  component: LogExplorer,
  validateSearch: (search: Record<string, unknown>) => {
    const result: {
      service?: string;
      traceId?: string;
      from?: string;
      to?: string;
      q?: string;
      mode?: string;
    } = {};
    if (typeof search.service === 'string') result.service = search.service;
    if (typeof search.traceId === 'string') result.traceId = search.traceId;
    if (typeof search.from === 'string') result.from = search.from;
    if (typeof search.to === 'string') result.to = search.to;
    if (typeof search.q === 'string') result.q = search.q;
    if (typeof search.mode === 'string') result.mode = search.mode;
    return result;
  },
});
const logSettingsRoute = createRoute({ getParentRoute: () => authedRoute, path: '/logs/settings', component: LogSettings });
const tracesRoute = createRoute({ getParentRoute: () => authedRoute, path: '/traces', component: TraceExplorer });
const traceSettingsRoute = createRoute({ getParentRoute: () => authedRoute, path: '/traces/settings', component: TraceSettings });
const traceCompareRoute = createRoute({
  getParentRoute: () => authedRoute,
  path: '/traces/compare',
  component: TraceCompare,
  validateSearch: (search: Record<string, unknown>) => ({
    a: typeof search.a === 'string' ? search.a : undefined,
    b: typeof search.b === 'string' ? search.b : undefined,
  }),
});
const traceDetailRoute = createRoute({ getParentRoute: () => authedRoute, path: '/traces/$traceId', component: TraceDetail });
const entityRoute = createRoute({
  getParentRoute: () => authedRoute,
  path: '/entities/$type/$id',
  component: EntityPage,
});
const serviceFlowRoute = createRoute({ getParentRoute: () => authedRoute, path: '/service-flow', component: ServiceFlow });
const transactionsRoute = createRoute({ getParentRoute: () => authedRoute, path: '/transactions', component: TransactionJourney });
const anomaliesRoute = createRoute({ getParentRoute: () => authedRoute, path: '/anomalies', component: AnomalyDetection });
const serviceMapRoute = createRoute({ getParentRoute: () => authedRoute, path: '/service-map', component: ServiceMap });
const aiChatRoute = createRoute({
  getParentRoute: () => authedRoute,
  path: '/ai-chat',
  component: AIChat,
  validateSearch: (search: Record<string, unknown>) => ({
    context: typeof search.context === 'string' ? search.context : undefined,
  }),
});
const alertsRoute = createRoute({ getParentRoute: () => authedRoute, path: '/alerts', component: Alerts });
const metricsRoute = createRoute({ getParentRoute: () => authedRoute, path: '/metrics', component: MetricsExplorer });
const dashboardsRoute = createRoute({ getParentRoute: () => authedRoute, path: '/dashboards', component: Dashboards });
const dashboardViewRoute = createRoute({
  getParentRoute: () => authedRoute,
  path: '/dashboards/$id',
  component: DashboardView,
  validateSearch: (search: Record<string, unknown>) => ({
    edit: typeof search.edit === 'string' ? search.edit : undefined,
  }),
});
const slosRoute = createRoute({ getParentRoute: () => authedRoute, path: '/slos', component: SLOs });
const infraRoute = createRoute({ getParentRoute: () => authedRoute, path: '/infrastructure', component: Infrastructure });
const k8sRoute = createRoute({ getParentRoute: () => authedRoute, path: '/kubernetes', component: Kubernetes });
const databasesRoute = createRoute({ getParentRoute: () => authedRoute, path: '/databases', component: Databases });
const middlewareRoute = createRoute({ getParentRoute: () => authedRoute, path: '/middleware', component: Middleware });
const rumRoute = createRoute({ getParentRoute: () => authedRoute, path: '/rum', component: RUM });
const rumReplayRoute = createRoute({
  getParentRoute: () => authedRoute,
  path: '/rum/sessions/$sessionId/replay',
  component: SessionReplay,
});
const syntheticRoute = createRoute({ getParentRoute: () => authedRoute, path: '/synthetic', component: Synthetic });
const workflowsRoute = createRoute({ getParentRoute: () => authedRoute, path: '/workflows', component: Workflows });
const notebooksRoute = createRoute({ getParentRoute: () => authedRoute, path: '/notebooks', component: Notebooks });
const securityRoute = createRoute({ getParentRoute: () => authedRoute, path: '/security', component: Security });
const securityAttackRoute = createRoute({
  getParentRoute: () => authedRoute,
  path: '/security/attacks/$id',
  component: SecurityAttackDetail,
});
const marketplaceRoute = createRoute({ getParentRoute: () => authedRoute, path: '/marketplace', component: Marketplace });
const cloudRoute = createRoute({ getParentRoute: () => authedRoute, path: '/cloud', component: CloudMonitoring });
const workflowEditorRoute = createRoute({
  getParentRoute: () => authedRoute,
  path: '/workflows/editor',
  component: WorkflowEditor,
  validateSearch: (search: Record<string, unknown>) => ({
    id: typeof search.id === 'string' ? search.id : undefined,
  }),
});
const integrationsRoute = createRoute({ getParentRoute: () => authedRoute, path: '/integrations', component: Integrations });
const settingsRoute = createRoute({ getParentRoute: () => authedRoute, path: '/settings', component: Settings });
const settingsUsersRoute = createRoute({ getParentRoute: () => authedRoute, path: '/settings/users', component: SettingsUsers });
const settingsApiKeysRoute = createRoute({ getParentRoute: () => authedRoute, path: '/settings/api-keys', component: SettingsApiKeys });
const settingsAuditRoute = createRoute({ getParentRoute: () => authedRoute, path: '/settings/audit', component: SettingsAudit });
const settingsUsageRoute = createRoute({ getParentRoute: () => authedRoute, path: '/settings/usage', component: SettingsUsage });
const settingsSSORoute = createRoute({ getParentRoute: () => authedRoute, path: '/settings/sso', component: SettingsSSO });
const settingsPoliciesRoute = createRoute({ getParentRoute: () => authedRoute, path: '/settings/policies', component: SettingsPolicies });
const settingsOncallRoute = createRoute({ getParentRoute: () => authedRoute, path: '/settings/oncall', component: SettingsOncall });
const designSystemRoute = createRoute({ getParentRoute: () => authedRoute, path: '/design-system', component: DesignSystem });
const healthRoute = createRoute({ getParentRoute: () => authedRoute, path: '/health', component: HealthPage });

const routeTree = rootRoute.addChildren([
  loginRoute,
  authCallbackRoute,
  authedRoute.addChildren([
    indexRoute,
    incidentsRoute,
    incidentDetailRoute,
    logsRoute,
    logSettingsRoute,
    tracesRoute,
    traceDetailRoute,
    traceSettingsRoute,
    traceCompareRoute,
    entityRoute,
    serviceFlowRoute,
    transactionsRoute,
    anomaliesRoute,
    serviceMapRoute,
    aiChatRoute,
    alertsRoute,
    metricsRoute,
    dashboardsRoute,
    dashboardViewRoute,
    slosRoute,
    infraRoute,
    k8sRoute,
    databasesRoute,
    middlewareRoute,
    rumRoute,
    rumReplayRoute,
    syntheticRoute,
    workflowsRoute,
    workflowEditorRoute,
    notebooksRoute,
    securityRoute,
    securityAttackRoute,
    marketplaceRoute,
    cloudRoute,
    integrationsRoute,
    settingsRoute,
    settingsUsersRoute,
    settingsApiKeysRoute,
    settingsAuditRoute,
    settingsUsageRoute,
    settingsSSORoute,
    settingsPoliciesRoute,
    settingsOncallRoute,
    designSystemRoute,
    healthRoute,
  ]),
]);

export const router = createRouter({ routeTree });

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
