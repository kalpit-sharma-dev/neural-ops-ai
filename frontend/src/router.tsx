import { createRootRoute, createRoute, createRouter, Outlet } from '@tanstack/react-router';
import { Layout } from './components/Layout';
import { ProtectedRoute } from './components/ProtectedRoute';
import Dashboard from './pages/Dashboard';
import LogExplorer from './pages/LogExplorer';
import LogSettings from './pages/LogSettings';
import Incidents from './pages/Incidents';
import IncidentDetail from './pages/IncidentDetail';
import { lazyPage } from './routes/lazyPage';

const ServiceMap = lazyPage(() => import('./pages/ServiceMap'));
const RUM = lazyPage(() => import('./pages/RUM'));
const Kubernetes = lazyPage(() => import('./pages/Kubernetes'));
const Infrastructure = lazyPage(() => import('./pages/Infrastructure'));
const SessionReplay = lazyPage(() => import('./pages/SessionReplay'));
import AIChat from './pages/AIChat';
import TransactionJourney from './pages/TransactionJourney';
import DesignSystem from './pages/DesignSystem';
import AnomalyDetection from './pages/AnomalyDetection';
import TraceExplorer from './pages/TraceExplorer';
import TraceDetail from './pages/TraceDetail';
import ServiceFlow from './pages/ServiceFlow';
import Alerts from './pages/Alerts';
import Settings from './pages/Settings';
import SettingsUsers from './pages/SettingsUsers';
import SettingsApiKeys from './pages/SettingsApiKeys';
import SettingsAudit from './pages/SettingsAudit';
import SettingsUsage from './pages/SettingsUsage';
import SettingsSSO from './pages/SettingsSSO';
import SettingsPolicies from './pages/SettingsPolicies';
import SettingsOncall from './pages/SettingsOncall';
import MetricsExplorer from './pages/MetricsExplorer';
import Dashboards from './pages/Dashboards';
import DashboardView from './pages/DashboardView';
import SLOs from './pages/SLOs';
import Databases from './pages/Databases';
import Middleware from './pages/Middleware';
import Synthetic from './pages/Synthetic';
import Workflows from './pages/Workflows';
import Notebooks from './pages/Notebooks';
import Security from './pages/Security';
import SecurityAttackDetail from './pages/SecurityAttackDetail';
import Marketplace from './pages/Marketplace';
import TraceSettings from './pages/TraceSettings';
import CloudMonitoring from './pages/CloudMonitoring';
import WorkflowEditor from './pages/WorkflowEditor';
import Integrations from './pages/Integrations';
import TraceCompare from './pages/TraceCompare';
import EntityPage from './pages/EntityPage';
import { HealthPage } from './pages/HealthPage';
import Login from './pages/Login';
import AuthCallback from './pages/AuthCallback';
import { useAuthStore } from './store/authStore';

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
const workflowEditorRoute = createRoute({ getParentRoute: () => authedRoute, path: '/workflows/editor', component: WorkflowEditor });
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
