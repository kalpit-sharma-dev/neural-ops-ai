import { getData, postData, putData, deleteData } from './client';

export interface Span {
  traceId: string;
  spanId: string;
  parentId?: string;
  service: string;
  operation: string;
  startTime: string;
  durationMs: number;
  status: string;
  tags?: Record<string, string>;
}

export interface TraceDetail {
  traceId: string;
  rootSpanId: string;
  spans: Span[];
  totalMs: number;
  service: string;
  status: string;
  spanCount: number;
}

export interface TraceSummary {
  traceId: string;
  service: string;
  operation: string;
  durationMs: number;
  status: string;
  startTime: string;
  spanCount: number;
}

export interface FlowEdge {
  source: string;
  target: string;
  callCount: number;
  errorRate: number;
  p50Ms: number;
  p95Ms: number;
}

export interface MetricSeries {
  name: string;
  labels?: Record<string, string>;
  unit?: string;
  points: { timestamp: string; value: number }[];
}

export interface Dashboard {
  id: string;
  name: string;
  description?: string;
  shared: boolean;
  tiles: { id: string; type: string; title: string; metric?: string; query?: string; position: Record<string, number> }[];
}

export interface TopologyGraph {
  nodes: { id: string; displayName: string; type: string; health: string; errorRate: number; throughputRpm: number; zone?: string }[];
  edges: { source: string; target: string; callCount: number; errorRate: number; p95Ms: number }[];
  at: string;
}

export async function fetchTrace(traceId: string) {
  return getData<TraceDetail>(`/apm/traces/${encodeURIComponent(traceId)}`);
}

export async function searchTraces(body: Record<string, unknown>) {
  return postData<TraceSummary[]>('/apm/traces/search', body);
}

export async function fetchServiceFlow() {
  return getData<FlowEdge[]>('/apm/flow');
}

export async function fetchMetricCatalog() {
  return getData<{ name: string; description: string; unit: string; labels: string[] }[]>('/metrics/catalog');
}

export async function queryMetric(name: string, service?: string) {
  return getData<MetricSeries>('/metrics/query', { name, service });
}

export async function fetchDashboards() {
  return getData<Dashboard[]>('/dashboards');
}

export async function fetchDashboard(id: string) {
  return getData<Dashboard>(`/dashboards/${id}`);
}

export async function createDashboard(body: Partial<Dashboard>) {
  return postData<Dashboard>('/dashboards', body);
}

export async function updateDashboard(id: string, body: Partial<Dashboard>) {
  return putData<Dashboard>(`/dashboards/${encodeURIComponent(id)}`, body);
}

export async function queryPromQL(query: string) {
  return getData<MetricSeries>('/metrics/promql', { query });
}

export async function fetchSessionReplay(sessionId: string) {
  return getData<{ seq: number; type: string; payload: Record<string, unknown>; recordedAt: string }[]>(
    `/rum/sessions/${encodeURIComponent(sessionId)}/replay`,
  );
}

export async function ingestRUMBeacon(body: Record<string, unknown>) {
  return postData('/rum/beacon', body);
}

export async function fetchServiceOperations(service: string) {
  return getData<string[]>(`/apm/services/${encodeURIComponent(service)}/operations`);
}

export async function fetchAttack(id: string) {
  return getData<{ id: string; type: string; sourceIp: string; service: string; blocked: boolean; detectedAt: string }>(
    `/security/attacks/${encodeURIComponent(id)}`,
  );
}

export async function createWorkflow(body: { name: string; trigger: string; enabled: boolean; steps: string[] }) {
  return postData('/workflows', body);
}

export async function createNotebook(body: { name: string; cells: { id: string; type: string; content: string }[] }) {
  return postData('/notebooks', body);
}

export async function fetchTopology(zone?: string) {
  return getData<TopologyGraph>('/topology', zone ? { zone } : undefined);
}

export async function fetchZones() {
  return getData<{ id: string; name: string; services: string[] }[]>('/zones');
}

export async function fetchEntityAnomalies() {
  return getData<{ id: string; entityId: string; service: string; metric: string; score: number; message: string; detectedAt: string }[]>('/anomalies/entities');
}

export async function fetchSLOs() {
  return getData<{ id: string; name: string; service: string; target: number; errorBudget: number; burnRate: number; status: string }[]>('/slos');
}

export async function fetchHosts() {
  return getData<{ id: string; name: string; status: string; cpuPercent: number; memoryPercent: number; diskPercent: number; zone: string }[]>('/infra/hosts');
}

export async function fetchK8sClusters() {
  return getData<{ id: string; name: string; nodes: number; pods: number; health: string }[]>('/infra/k8s/clusters');
}

export async function fetchK8sPods(namespace?: string) {
  return getData<{ id: string; name: string; namespace: string; node: string; status: string; cpuPercent: number; memoryPercent: number; restarts: number }[]>('/infra/k8s/pods', namespace ? { namespace } : undefined);
}

export async function fetchDatabases() {
  return getData<{ id: string; name: string; engine: string; status: string; qps: number; slowQueries: number; connections: number }[]>('/databases');
}

export async function fetchDBStatements(id: string) {
  return getData<{ query: string; calls: number; avgMs: number; totalMs: number }[]>(`/databases/${id}/statements`);
}

export async function fetchKafkaLag() {
  return getData<{ topic: string; consumerGroup: string; lag: number; partition: number }[]>('/middleware/kafka/lag');
}

export async function fetchRUMSessions() {
  return getData<{ id: string; userId: string; page: string; device: string; country: string; durationMs: number; errors: number; lcp: number; startedAt: string }[]>('/rum/sessions');
}

export async function fetchSyntheticMonitors() {
  return getData<{ id: string; name: string; type: string; url: string; interval: string; enabled: boolean; lastStatus: string; lastRunAt: string }[]>('/synthetic/monitors');
}

export async function fetchSyntheticRuns(id: string) {
  return getData<{ id: string; status: string; latencyMs: number; location: string; ranAt: string }[]>(`/synthetic/monitors/${id}/runs`);
}

export async function fetchWorkflows() {
  return getData<{ id: string; name: string; trigger: string; enabled: boolean; steps: string[] }[]>('/workflows');
}

export async function fetchNotebooks() {
  return getData<{ id: string; name: string; cells: { id: string; type: string; content: string }[]; updatedAt: string }[]>('/notebooks');
}

export async function fetchVulnerabilities() {
  return getData<{ id: string; cve: string; severity: string; service: string; description: string; detectedAt: string }[]>('/security/vulnerabilities');
}

export async function fetchAttacks() {
  return getData<{ id: string; type: string; sourceIp: string; service: string; blocked: boolean; detectedAt: string }[]>('/security/attacks');
}

export async function fetchIntegrations() {
  return getData<
    {
      id: string;
      integrationKey?: string;
      name: string;
      type: string;
      status: string;
      connected: boolean;
      configPublic?: Record<string, string>;
    }[]
  >('/integrations');
}

export async function connectIntegration(id: string, config?: Record<string, string>) {
  return postData<{ id: string; name: string; connected: boolean; status: string }>(
    `/integrations/${encodeURIComponent(id)}/connect`,
    config ? { config } : undefined,
  );
}

export async function fetchAdminUsers() {
  return getData<{ id: string; email: string; role: string; active: boolean }[]>('/admin/users');
}

export async function fetchAPIKeys() {
  return getData<{ id: string; name: string; prefix: string; scopes: string[]; createdAt: string }[]>('/admin/api-keys');
}

export async function fetchAuditLog() {
  return getData<{ id: string; userId: string; action: string; resource: string; detail: string; timestamp: string }[]>('/admin/audit');
}

export async function createSLO(body: {
  name: string;
  service: string;
  sliQuery: string;
  target: number;
  windowDays?: number;
}) {
  return postData('/slos', body);
}

export async function createAPIKey(body: { name: string; role?: string }) {
  return postData<{ key: { id: string; name: string; prefix: string; scopes: string[] }; secret: string }>(
    '/admin/api-keys',
    body,
  );
}

export async function fetchUsage() {
  return getData<{ logsIngestedGb: number; tracesIngested: number; aiTokensUsed: number; activeUsers: number }>('/admin/usage');
}

export async function fetchLogMetricRules() {
  return getData<{ id: string; name: string; pattern: string; service?: string; enabled: boolean }[]>('/logs/metric-rules');
}

export async function createLogMetricRule(body: { name: string; pattern: string; service?: string; enabled: boolean }) {
  return postData('/logs/metric-rules', body);
}

export async function fetchLogParsingRules() {
  return getData<{ id: string; name: string; pattern: string; field: string; enabled: boolean }[]>('/logs/parsing-rules');
}

export async function createLogParsingRule(body: { name: string; pattern: string; field: string; enabled: boolean }) {
  return postData('/logs/parsing-rules', body);
}

export async function fetchTraceRetention() {
  return getData<{ tenantId: string; retentionDays: number; headSampleRate: number; tailSampleRate: number }>(
    '/apm/retention',
  );
}

export async function updateTraceRetention(body: {
  retentionDays: number;
  headSampleRate: number;
  tailSampleRate: number;
}) {
  return putData('/apm/retention', body);
}

export async function fetchProfiles(service: string) {
  return getData<
    { functionName: string; filePath?: string; lineNo: number; selfTimeMs: number; sampleCount: number; service: string }[]
  >(`/apm/services/${encodeURIComponent(service)}/profiles`);
}

export async function fetchK8sNamespaces() {
  return getData<{ id: string; name: string; status: string; podCount: number }[]>('/infra/k8s/namespaces');
}

export async function fetchK8sDeployments(namespace?: string) {
  return getData<{ id: string; name: string; namespace: string; replicas: number; readyReplicas: number }[]>(
    '/infra/k8s/deployments',
    namespace ? { namespace } : undefined,
  );
}

export async function fetchCloudDashboards(provider?: string) {
  return getData<{ id: string; provider: string; name: string; region: string; metrics: string[] }[]>(
    '/cloud/dashboards',
    provider ? { provider } : undefined,
  );
}

export async function recordRUMConsent(body: { sessionId: string; consentGiven: boolean; consentVersion: string }) {
  return postData('/rum/consent', body);
}

export async function updateSLOBurnAlert(id: string, body: { enabled: boolean; threshold: number }) {
  return putData(`/slos/${encodeURIComponent(id)}/burn-alert`, body);
}

export async function triggerWorkflows(body: { trigger: string; context: Record<string, string> }) {
  return postData<{ runId: string; status: string; stepsLog: string[] }[]>('/workflows/trigger', body);
}

export async function fetchCloudMetrics(provider: string, metric: string, region?: string) {
  return getData<{ provider: string; metric: string; region: string; unit: string; source: string; points: { timestamp: string; value: number }[] }>(
    '/cloud/metrics',
    { provider, metric, ...(region ? { region } : {}) },
  );
}

export async function executeNotebook(id: string) {
  return postData<{ cellId: string; type: string; output: string; durationMs: number }[]>(
    `/notebooks/${encodeURIComponent(id)}/execute`,
  );
}

export async function disconnectIntegration(id: string) {
  return postData(`/integrations/${encodeURIComponent(id)}/disconnect`);
}

export function integrationOAuthStartUrl(id: string) {
  return `/api/v1/integrations/${encodeURIComponent(id)}/oauth/start`;
}

export async function fetchSSOConfig() {
  return getData<{ provider: string; metadataUrl?: string; clientId?: string; issuerUrl?: string; clientSecretSet?: boolean }>('/admin/sso');
}

export async function updateSSOConfig(body: { provider: string; metadataUrl?: string; clientId?: string; issuerUrl?: string; clientSecret?: string }) {
  return putData('/admin/sso', body);
}

export async function fetchTenantPolicies() {
  return getData<{ logRetentionDays: number; ingestionRateLimit: number }>('/admin/tenant-policies');
}

export async function updateTenantPolicies(body: { logRetentionDays: number; ingestionRateLimit: number }) {
  return putData('/admin/tenant-policies', body);
}

export async function fetchOncallSchedules() {
  return getData<{ id: string; team: string; timezone: string; enabled: boolean; rotation: { name: string; email: string; after: string }[] }[]>(
    '/admin/oncall',
  );
}

export async function createOncallSchedule(body: {
  team: string;
  timezone?: string;
  enabled?: boolean;
  rotation: { name: string; email: string; after: string }[];
}) {
  return postData('/admin/oncall', body);
}

export async function updateOncallSchedule(id: string, body: {
  team: string;
  timezone?: string;
  enabled?: boolean;
  rotation: { name: string; email: string; after: string }[];
}) {
  return putData(`/admin/oncall/${encodeURIComponent(id)}`, body);
}

export async function deleteOncallSchedule(id: string) {
  return deleteData(`/admin/oncall/${encodeURIComponent(id)}`);
}

export async function syncOncallPagerDuty(id: string, scheduleId?: string) {
  return postData(`/admin/oncall/${encodeURIComponent(id)}/sync-pagerduty`, scheduleId ? { scheduleId } : {});
}
