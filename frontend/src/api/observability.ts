import { getData, postData, putData, patchData, deleteData } from './client';

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

export interface TimeRangeParams {
  start?: string;
  end?: string;
}

export interface UnifiedQueryRequest {
  query: string;
  from?: 'all' | 'logs' | 'metrics' | 'traces' | 'events';
  service?: string;
  traceId?: string;
  txnId?: string;
  limit?: number;
}

export interface QueryPlannerStep {
  signal: string;
  action: string;
  estimatedCost: number;
}

export interface QueryExplainResponse {
  valid: boolean;
  message?: string;
  steps: QueryPlannerStep[];
  stores: string[];
  joinKeys?: string[];
  cardinality: {
    maxHits: number;
    maxCardinality: number;
    estimatedCardinality: number;
    truncated: boolean;
    cardinalityCapped: boolean;
  };
  estimatedMs: number;
}

export interface UnifiedQueryResponse {
  hits: UnifiedQueryHit[];
  count: number;
  planner?: QueryPlannerStep[];
  cardinality?: QueryExplainResponse['cardinality'];
}

export interface SavedQuery {
  id: string;
  name: string;
  query: string;
  from?: string;
  service?: string;
  traceId?: string;
  txnId?: string;
  updatedAt: string;
}

export interface AutoInstrumentationRuntime {
  language: string;
  runtime: string;
  method: string;
  status: string;
  notes?: string;
  otelPackage?: string;
  features?: string[];
}

export interface CollectorFleetAgent {
  id: string;
  name: string;
  environment: string;
  version: string;
  status: string;
  lastHeartbeatAt: string;
  policyId?: string;
}

export interface CollectorPipelineStage {
  id: string;
  type: string;
  config?: Record<string, string>;
  enabled: boolean;
}

export interface CollectorPipeline {
  id: string;
  name: string;
  description?: string;
  enabled: boolean;
  stages: CollectorPipelineStage[];
  updatedAt: string;
}

export async function queryMetric(name: string, service?: string, range?: TimeRangeParams) {
  return getData<MetricSeries>('/metrics/query', {
    name,
    service,
    ...(range?.start ? { start: range.start } : {}),
    ...(range?.end ? { end: range.end } : {}),
  });
}

export async function unifiedQuery(body: UnifiedQueryRequest) {
  return postData<UnifiedQueryResponse>('/query/unified', body);
}

export interface UnifiedQueryHit {
  id: string;
  signal: string;
  service?: string;
  title: string;
  summary: string;
  severity?: string;
  timestamp: string;
  link?: string;
  fields?: Record<string, unknown>;
}

export async function explainUnifiedQuery(body: UnifiedQueryRequest) {
  return postData<QueryExplainResponse>('/query/explain', body);
}

export async function validateUnifiedQuery(body: UnifiedQueryRequest) {
  return postData<{ valid: boolean; message: string; explain?: QueryExplainResponse }>('/query/validate', body);
}

export async function fetchSavedQueries() {
  return getData<SavedQuery[]>('/query/saved');
}

export async function createSavedQuery(body: Omit<SavedQuery, 'updatedAt'>) {
  return postData<SavedQuery>('/query/saved', body);
}

export async function deleteSavedQuery(id: string) {
  return deleteData<{ deleted: boolean }>(`/query/saved/${encodeURIComponent(id)}`);
}

export async function fetchAutoInstrumentationMatrix() {
  return getData<AutoInstrumentationRuntime[]>('/collectors/autoinstrumentation');
}

export async function fetchUnifiedQueryFunctions() {
  return getData<string[]>('/query/functions');
}

export async function fetchCollectorFleet() {
  return getData<CollectorFleetAgent[]>('/collectors/fleet');
}

export async function createCollectorAgent(body: Omit<CollectorFleetAgent, 'id'>) {
  return postData<CollectorFleetAgent>('/collectors/fleet', body);
}

export async function updateCollectorAgent(id: string, body: Omit<CollectorFleetAgent, 'id'>) {
  return putData<CollectorFleetAgent>(`/collectors/fleet/${encodeURIComponent(id)}`, body);
}

export async function fetchCollectorPipelines() {
  return getData<CollectorPipeline[]>('/collectors/pipelines');
}

export async function createCollectorPipeline(body: Omit<CollectorPipeline, 'id' | 'updatedAt'>) {
  return postData<CollectorPipeline>('/collectors/pipelines', body);
}

export async function updateCollectorPipeline(id: string, body: Omit<CollectorPipeline, 'id' | 'updatedAt'>) {
  return putData<CollectorPipeline>(`/collectors/pipelines/${encodeURIComponent(id)}`, body);
}

export async function validateCollectorPipeline(id: string, body: Omit<CollectorPipeline, 'id' | 'updatedAt'>) {
  return postData<{ valid: boolean; message: string }>(`/collectors/pipelines/${encodeURIComponent(id)}/validate`, body);
}

export interface AlertPolicyRoute {
  channel: string;
  target: string;
  after: string;
  priority: number;
}

export interface AlertPolicyContext {
  runbookUrl?: string;
  owner?: string;
  topologyLink?: string;
  tracePivotLink?: string;
  logPivotLink?: string;
}

export interface AlertPolicy {
  id: string;
  name: string;
  servicePattern: string;
  severity: string;
  enabled: boolean;
  routes: AlertPolicyRoute[];
  context?: AlertPolicyContext;
  expression?: string;
  dedupeKey?: string;
}

export interface AlertEvaluationResult {
  policyId: string;
  matched: boolean;
  suppressed: boolean;
  routes?: AlertPolicyRoute[];
  context: AlertPolicyContext;
  dedupeKey?: string;
  explanation?: string;
  triggeredAt: string;
}

export async function fetchAlertPolicies() {
  return getData<AlertPolicy[]>('/alerts/policies');
}

export async function createAlertPolicy(body: Omit<AlertPolicy, 'id'>) {
  return postData<AlertPolicy>('/alerts/policies', body);
}

export async function deleteAlertPolicy(id: string) {
  return deleteData<{ deleted: boolean }>(`/alerts/policies/${encodeURIComponent(id)}`);
}

export async function triggerAlertPolicy(id: string, body: { service: string; severity?: string }) {
  return postData<AlertEvaluationResult>(`/alerts/policies/${encodeURIComponent(id)}/trigger`, body);
}

export async function fetchCollectorAgent(id: string) {
  return getData<CollectorFleetAgent>(`/collectors/fleet/${encodeURIComponent(id)}`);
}

export async function upgradeCollectorAgent(id: string, targetVersion?: string) {
  return postData<CollectorFleetAgent>(`/collectors/fleet/${encodeURIComponent(id)}/upgrade`, {
    targetVersion: targetVersion ?? '',
  });
}

export interface AlertSuppression {
  id: string;
  servicePattern: string;
  reason: string;
  startsAt: string;
  endsAt: string;
  createdBy: string;
}

export async function fetchAlertSuppressions() {
  return getData<AlertSuppression[]>('/alerts/suppressions');
}

export async function createAlertSuppression(body: { servicePattern: string; reason: string; duration: string }) {
  return postData<AlertSuppression>('/alerts/suppressions', body);
}

export async function deleteAlertSuppression(id: string) {
  return deleteData<{ deleted: boolean }>(`/alerts/suppressions/${encodeURIComponent(id)}`);
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

export async function queryPromQL(query: string, range?: TimeRangeParams) {
  return getData<MetricSeries>('/metrics/promql', {
    query,
    ...(range?.start ? { start: range.start } : {}),
    ...(range?.end ? { end: range.end } : {}),
  });
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

export interface WorkflowGraphNode {
  id: string;
  type: string;
  label: string;
  x: number;
  y: number;
}

export type WorkflowEdgeCondition = 'success' | 'failure' | 'always';

export interface WorkflowGraphEdge {
  id: string;
  source: string;
  target: string;
  /** Gates runtime traversal. Omitted/"success" follows on predecessor success. */
  condition?: WorkflowEdgeCondition | string;
}

export interface WorkflowGraph {
  nodes: WorkflowGraphNode[];
  edges: WorkflowGraphEdge[];
}

export interface Workflow {
  id: string;
  name: string;
  trigger: string;
  enabled: boolean;
  steps: string[];
  graph?: WorkflowGraph;
}

export interface WorkflowInput {
  name: string;
  trigger: string;
  enabled: boolean;
  steps: string[];
  graph?: WorkflowGraph;
}

export async function createWorkflow(body: WorkflowInput) {
  return postData<Workflow>('/workflows', body);
}

export async function updateWorkflow(id: string, body: WorkflowInput) {
  return putData<Workflow>(`/workflows/${encodeURIComponent(id)}`, body);
}

export async function deleteWorkflow(id: string) {
  return deleteData<{ deleted: boolean }>(`/workflows/${encodeURIComponent(id)}`);
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
  return getData<Workflow[]>('/workflows');
}

export async function fetchNotebooks() {
  return getData<{ id: string; name: string; cells: { id: string; type: string; content: string }[]; updatedAt: string }[]>('/notebooks');
}

export async function fetchVulnerabilities() {
  return getData<{ id: string; cve: string; severity: string; service: string; description: string; detectedAt: string }[]>('/security/vulnerabilities');
}

export interface SecurityFinding {
  id: string;
  title: string;
  category: string;
  severity: string;
  service?: string;
  asset?: string;
  status: string;
  exploitability?: string;
  incidentId?: string;
  detectedAt: string;
}

export async function fetchSecurityFindings() {
  return getData<SecurityFinding[]>('/security/findings');
}

export async function fetchSecurityFinding(id: string) {
  return getData<SecurityFinding>(`/security/findings/${encodeURIComponent(id)}`);
}

export async function correlateSecurityFinding(id: string) {
  return postData<SecurityFinding>(`/security/findings/${encodeURIComponent(id)}/correlate`, {});
}

export async function importSecurityFindings(source: string, findings: SecurityFinding[]) {
  return postData<{ accepted: boolean; source: string; count: number }>('/security/sca/import', { source, findings });
}

export interface SecurityPostureCheck {
  id: string;
  name: string;
  provider: string;
  resource: string;
  status: string;
  severity: string;
}

export async function fetchSecurityPosture() {
  return getData<SecurityPostureCheck[]>('/security/cspm/posture');
}

export async function exportSecurityToSIEM(provider: string, target: string) {
  return postData<{ queued: boolean; provider: string; target: string }>('/security/siem/exports', { provider, target });
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

export interface MarketplaceConfigField {
  key: string;
  label: string;
  secret?: boolean;
}

export interface MarketplaceExtension {
  key: string;
  name: string;
  category: string;
  description: string;
  publisher: string;
  version: string;
  installed: boolean;
  configFields?: MarketplaceConfigField[];
  configPublic?: Record<string, string>;
}

export async function fetchMarketplace() {
  return getData<MarketplaceExtension[]>('/marketplace');
}

export async function installExtension(key: string, config?: Record<string, string>) {
  return postData<MarketplaceExtension>(
    `/marketplace/${encodeURIComponent(key)}/install`,
    config ? { config } : { config: {} },
  );
}

export async function uninstallExtension(key: string) {
  return postData<{ uninstalled: boolean; key: string }>(
    `/marketplace/${encodeURIComponent(key)}/uninstall`,
  );
}

export async function connectIntegration(id: string, config?: Record<string, string>) {
  return postData<{ id: string; name: string; connected: boolean; status: string }>(
    `/integrations/${encodeURIComponent(id)}/connect`,
    config ? { config } : undefined,
  );
}

export interface AdminUser {
  id: string;
  email: string;
  role: string;
  active: boolean;
  tenantId?: string;
}

export async function fetchAdminUsers() {
  return getData<AdminUser[]>('/admin/users');
}

export async function createAdminUser(body: { email: string; role: string }) {
  return postData<AdminUser>('/admin/users', body);
}

export async function updateAdminUser(id: string, body: { role?: string; active?: boolean }) {
  return patchData<{ updated: boolean }>(`/admin/users/${encodeURIComponent(id)}`, body);
}

export const USER_ROLES = ['ADMIN', 'SRE', 'DEVELOPER', 'ALERT_MANAGER', 'READONLY'] as const;

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

export interface AIExplanationEvidence {
  signal: string;
  ref: string;
  detail: string;
}

export interface AIExplanationNode {
  id: string;
  label: string;
  confidence: number;
  evidence?: AIExplanationEvidence[];
  children?: AIExplanationNode[];
}

export interface RCAResponse {
  explanationId: string;
  incidentId: string;
  summary: string;
  confidence: number;
  root: AIExplanationNode;
  generatedAt: string;
}

export interface AIForecast {
  id: string;
  metric: string;
  service?: string;
  horizon: string;
  prediction: number;
  lowerBound: number;
  upperBound: number;
  unit: string;
  recommendation?: string;
  generatedAt: string;
}

export interface AutoFixStep {
  id: string;
  action: string;
  description: string;
  blastRadius: string;
}

export interface AutoFixPlan {
  id: string;
  incidentId: string;
  summary: string;
  requiresApproval: boolean;
  policyPass: boolean;
  policyReason?: string;
  steps: AutoFixStep[];
  createdAt: string;
}

export interface AutoFixActionRecord {
  id: string;
  planId: string;
  incidentId: string;
  status: string;
  startedAt: string;
  finishedAt?: string;
  logs: string[];
}

export async function fetchIncidentRCA(incidentId: string) {
  return getData<RCAResponse>(`/ai/rca/${encodeURIComponent(incidentId)}`);
}

export interface IncidentDeepLinks {
  logs: string;
  traces: string;
  metrics: string;
  security: string;
  serviceMap: string;
  workflows: string;
  aiChat: string;
}

export interface SuggestedQuery {
  label: string;
  kind: string;
  href: string;
}

export interface IncidentUnifiedContext {
  incidentId: string;
  primaryService: string;
  affectedServices: string[];
  deepLinks: IncidentDeepLinks;
  securityFindings: SecurityFinding[];
  suggestedQueries: SuggestedQuery[];
  generatedAt: string;
}

export async function fetchIncidentUnifiedContext(
  incidentId: string,
  params?: { service?: string; services?: string },
) {
  const qs = new URLSearchParams();
  if (params?.service) qs.set('service', params.service);
  if (params?.services) qs.set('services', params.services);
  const q = qs.toString();
  return getData<IncidentUnifiedContext>(
    `/unified/incidents/${encodeURIComponent(incidentId)}/context${q ? `?${q}` : ''}`,
  );
}

export async function fetchAIExplanation(explanationId: string) {
  return getData<RCAResponse>(`/ai/explanations/${encodeURIComponent(explanationId)}`);
}

export async function postAIForecast(body: { metric: string; service?: string; horizon?: string }) {
  return postData<AIForecast>('/ai/forecast', body);
}

export async function createAutoFixPlan(incidentId: string) {
  return postData<AutoFixPlan>('/ai/autofix/plan', { incidentId });
}

export async function executeAutoFix(planId: string, approved: boolean) {
  return postData<AutoFixActionRecord>('/ai/autofix/execute', { planId, approved });
}

export async function rollbackAutoFix(actionId: string) {
  return postData<AutoFixActionRecord>('/ai/autofix/rollback', { actionId });
}

export interface CloudAsset {
  id: string;
  provider: string;
  type: string;
  name: string;
  region: string;
  accountId: string;
  status: string;
  tags?: Record<string, string>;
  monthlyUsd?: number;
  updatedAt: string;
}

export interface CloudAssetTopology {
  nodes: { id: string; label: string; provider: string; type: string; health: string }[];
  edges: { source: string; target: string; relation: string }[];
  at: string;
}

export interface FinOpsCostSeries {
  scope: string;
  unit: string;
  total: number;
  budget: number;
  budgetId?: string;
  points: { timestamp: string; amount: number }[];
  forecast?: FinOpsForecast;
  tagCoveragePct?: number;
  costView?: string;
}

export interface FinOpsForecast {
  horizon: string;
  p50: number;
  p95: number;
  lower: number;
  upper: number;
  method: string;
}

export interface FinOpsCostAnomaly {
  id: string;
  scope: string;
  service: string;
  provider: string;
  deltaPct: number;
  amountUsd: number;
  severity: string;
  status?: string;
  description: string;
  detectedAt: string;
  feedback?: string;
  probableCause?: string;
}

export interface FinOpsBreakdownNode {
  dimension: string;
  key: string;
  amountUsd: number;
  pct: number;
  children?: FinOpsBreakdownNode[];
}

export interface FinOpsAllocationRule {
  id: string;
  name: string;
  dimension: string;
  tagKey: string;
  tagValue?: string;
  priority: number;
  enabled: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface FinOpsBudget {
  id: string;
  name: string;
  scopeType: string;
  scopeValue: string;
  period: string;
  amountUsd: number;
  thresholds: number[];
  notifyPolicyId?: string;
  spendUsd?: number;
  burnPct?: number;
}

export interface FinOpsRecommendation {
  id: string;
  type: string;
  resourceId: string;
  scope: string;
  title: string;
  description: string;
  projectedSavingsUsd: number;
  realizedSavingsUsd?: number;
  riskScore: number;
  status: string;
  ticketId?: string;
  ticketUrl?: string;
  createdAt: string;
}

export interface FinOpsK8sCostRow {
  cluster: string;
  namespace: string;
  workload: string;
  cpuCostUsd: number;
  memCostUsd: number;
  idleCostUsd: number;
  totalUsd: number;
}

export interface FinOpsCarbonFootprint {
  scope: string;
  period: string;
  co2eKg: number;
  renewablePct: number;
  recommendation?: string;
  methodology?: string;
  factorVersion?: string;
  byDimension?: { dimension: string; key: string; co2eKg: number; pct: number }[];
  sci?: { co2ePerRequest: number; co2ePerTransaction: number; requestsPerMonth: number };
}

export interface FinOpsCommitment {
  id: string;
  provider: string;
  commitmentType: string;
  region: string;
  coveragePct: number;
  utilizationPct: number;
  monthlyCommitUsd: number;
  expiresAt: string;
  status: string;
}

export interface FinOpsCommitmentRecommendation {
  id: string;
  provider: string;
  commitmentType: string;
  termMonths: number;
  breakEvenMonths: number;
  monthlySavingsUsd: number;
  riskScore: number;
  description: string;
}

export interface FinOpsUnitEconomics {
  metric: string;
  scope: string;
  totalCostUsd: number;
  totalUnits: number;
  costPerUnit: number;
  period: string;
}

export interface FinOpsTagSuggestion {
  id: string;
  resourceId: string;
  suggestedKey: string;
  suggestedValue: string;
  confidence: number;
  spendUsd: number;
}

export interface FinOpsCarbonRecommendation {
  id: string;
  title: string;
  description: string;
  co2eReductionKg: number;
  costDeltaUsd: number;
  region: string;
}

export interface FinOpsReport {
  id: string;
  name: string;
  scope: string;
  format: string;
  url: string;
  generatedAt: string;
}

export interface FinOpsReportSchedule {
  id: string;
  name: string;
  scope: string;
  format: string;
  cadence: string;
  deliveryChannel: string;
  deliveryTarget: string;
  enabled: boolean;
}

export interface FinOpsAuditEntry {
  id: string;
  actor: string;
  action: string;
  entityType: string;
  entityId: string;
  createdAt: string;
}

export interface NetworkFlow {
  id: string;
  source: string;
  destination: string;
  protocol: string;
  port: number;
  bytes: number;
  packets: number;
  latencyMs: number;
  lossPct: number;
  jitterMs: number;
  timestamp: string;
}

export interface NetworkDevice {
  id: string;
  name: string;
  type: string;
  site: string;
  status: string;
  cpuUtil: number;
  memUtil: number;
  uptimePct: number;
}

export interface NetworkTopologyGraph {
  nodes: { id: string; label: string; type: string; health: string }[];
  edges: { source: string; target: string; latencyMs: number; lossPct: number; utilPct: number }[];
  at: string;
}

export interface NetworkAnomaly {
  id: string;
  link: string;
  metric: string;
  value: number;
  baseline: number;
  severity: string;
  description: string;
  detectedAt: string;
}

export async function fetchCloudAssets(provider?: string) {
  return getData<CloudAsset[]>('/cloud/assets', provider ? { provider } : undefined);
}

export async function fetchCloudAssetTopology() {
  return getData<CloudAssetTopology>('/cloud/topology');
}

export async function fetchFinOpsCosts(params?: { scope?: string; provider?: string; from?: string; to?: string }) {
  return getData<FinOpsCostSeries>('/finops/costs', params);
}

export async function fetchFinOpsAnomalies(params?: { scope?: string; severity?: string; status?: string }) {
  return getData<FinOpsCostAnomaly[]>('/finops/anomalies', params);
}

export async function fetchFinOpsCarbon(params?: { scope?: string; dimension?: string }) {
  return getData<FinOpsCarbonFootprint>('/finops/carbon', params);
}

export async function fetchFinOpsCostBreakdown(params?: { dimension?: string; from?: string; to?: string }) {
  return getData<FinOpsBreakdownNode>('/finops/costs/breakdown', params);
}

export async function fetchFinOpsAllocationRules() {
  return getData<FinOpsAllocationRule[]>('/finops/allocation/rules');
}

export async function createFinOpsAllocationRule(body: Omit<FinOpsAllocationRule, 'id' | 'createdAt' | 'updatedAt'>) {
  return postData<FinOpsAllocationRule>('/finops/allocation/rules', body);
}

export async function deleteFinOpsAllocationRule(id: string) {
  return deleteData<{ deleted: boolean }>(`/finops/allocation/rules/${encodeURIComponent(id)}`);
}

export async function fetchFinOpsKubernetesCost(params?: { cluster?: string; namespace?: string }) {
  return getData<FinOpsK8sCostRow[]>('/finops/kubernetes/cost', params);
}

export async function submitFinOpsAnomalyFeedback(id: string, feedback: 'confirm' | 'false_positive' | 'expected') {
  return postData<FinOpsCostAnomaly>(`/finops/anomalies/${encodeURIComponent(id)}/feedback`, { feedback });
}

export async function fetchFinOpsRecommendations(params?: { type?: string; scope?: string; status?: string }) {
  return getData<FinOpsRecommendation[]>('/finops/recommendations', params);
}

export async function finOpsRecommendationAction(id: string, action: string) {
  return postData<FinOpsRecommendation>(`/finops/recommendations/${encodeURIComponent(id)}/actions`, { action });
}

export async function fetchFinOpsBudgets() {
  return getData<FinOpsBudget[]>('/finops/budgets');
}

export interface FinOpsBudgetAlert {
  id: string;
  budgetId: string;
  budgetName: string;
  scope: string;
  threshold: number;
  burnPct: number;
  forecastPct?: number;
  severity: string;
  message: string;
}

export async function fetchFinOpsBudgetAlerts() {
  return getData<FinOpsBudgetAlert[]>('/finops/budgets/alerts');
}

export async function fetchFinOpsAnomaly(id: string) {
  return getData<FinOpsCostAnomaly>(`/finops/anomalies/${encodeURIComponent(id)}`);
}

export interface FinOpsIngestStatus {
  lastIngestAt?: string;
  lagSeconds: number;
  stale: boolean;
  providers: string[];
}

export async function fetchFinOpsIngestStatus() {
  return getData<FinOpsIngestStatus>('/finops/ingest/status');
}

export async function createFinOpsBudget(body: Omit<FinOpsBudget, 'id' | 'spendUsd' | 'burnPct'>) {
  return postData<FinOpsBudget>('/finops/budgets', body);
}

export async function deleteFinOpsBudget(id: string) {
  return deleteData<{ deleted: boolean }>(`/finops/budgets/${encodeURIComponent(id)}`);
}

export async function fetchFinOpsForecast(params?: { scope?: string; horizon?: string }) {
  return getData<FinOpsForecast>('/finops/forecast', params);
}

export async function runFinOpsIngest() {
  return postData<{ count: number }>('/finops/ingest/run', {});
}

export async function importFinOpsCost(body: {
  source: string;
  format?: string;
  dedupeKey?: string;
  items: { resourceId: string; service: string; provider?: string; region?: string; effectiveCost: number; usageType?: string; tags?: Record<string, string> }[];
}) {
  return postData<{ id: string; lineCount: number; totalUsd: number; version: number }>('/finops/imports', body);
}

export async function fetchFinOpsTagSuggestions() {
  return getData<FinOpsTagSuggestion[]>('/finops/allocation/tag-suggestions');
}

export async function fetchFinOpsCommitments(params?: { provider?: string; status?: string }) {
  return getData<FinOpsCommitment[]>('/finops/commitments', params);
}

export async function fetchFinOpsCommitmentRecommendations() {
  return getData<FinOpsCommitmentRecommendation[]>('/finops/commitments/recommendations');
}

export async function fetchFinOpsUnitEconomics(params?: { metric?: string; scope?: string }) {
  return getData<FinOpsUnitEconomics>('/finops/unit-economics', params);
}

export async function fetchFinOpsCarbonRecommendations() {
  return getData<FinOpsCarbonRecommendation[]>('/finops/carbon/recommendations');
}

export async function fetchFinOpsReports(scope?: string) {
  return getData<FinOpsReport[]>('/finops/reports', scope ? { scope } : undefined);
}

export async function createFinOpsReportSchedule(body: Omit<FinOpsReportSchedule, 'id'>) {
  return postData<FinOpsReportSchedule>('/finops/reports/schedule', body);
}

export async function fetchFinOpsAuditLog() {
  return getData<FinOpsAuditEntry[]>('/finops/audit');
}

export interface FinOpsChargebackStatement {
  id: string;
  costCenter: string;
  mode: string;
  period: string;
  totalUsd: number;
  lines: { team: string; service: string; amountUsd: number; allocatedPct: number }[];
  exportUrl?: string;
  generatedAt: string;
}

export interface FinOpsScenarioResult {
  id: string;
  name: string;
  scenarioType: string;
  baselineCostUsd: number;
  projectedCostUsd: number;
  costDeltaUsd: number;
  baselineCo2eKg: number;
  projectedCo2eKg: number;
  co2eDeltaKg: number;
  summary: string;
  createdAt: string;
}

export interface FinOpsCommitmentAlert {
  id: string;
  commitmentId: string;
  provider: string;
  alertType: string;
  severity: string;
  message: string;
  utilizationPct?: number;
  expiresAt?: string;
}

export interface FinOpsGovernancePolicy {
  id: string;
  name: string;
  residencyRegion: string;
  allowedScopes: string[];
  chargebackMode: string;
  enabled: boolean;
}

export async function fetchFinOpsChargebackStatements(costCenter?: string) {
  return getData<FinOpsChargebackStatement[]>('/finops/chargeback/statements', costCenter ? { costCenter } : undefined);
}

export async function generateFinOpsChargebackStatement(body: { costCenter: string; mode?: string; period?: string }) {
  return postData<FinOpsChargebackStatement>('/finops/chargeback/statements', body);
}

export async function runFinOpsScenario(body: {
  name: string;
  scenarioType: string;
  scope?: string;
  params?: Record<string, string>;
}) {
  return postData<FinOpsScenarioResult>('/finops/scenarios', body);
}

export async function fetchFinOpsScenarios() {
  return getData<FinOpsScenarioResult[]>('/finops/scenarios');
}

export async function fetchFinOpsCommitmentAlerts(provider?: string) {
  return getData<FinOpsCommitmentAlert[]>('/finops/commitments/alerts', provider ? { provider } : undefined);
}

export async function finOpsCarbonAction(id: string, action: 'simulate' | 'apply' | 'dismiss') {
  return postData<{ id: string; status: string; co2eReductionKg: number; costDeltaUsd: number }>(
    `/finops/carbon/recommendations/${encodeURIComponent(id)}/actions`,
    { action },
  );
}

export async function fetchFinOpsGovernancePolicies() {
  return getData<FinOpsGovernancePolicy[]>('/finops/governance/policies');
}

export async function fetchNetworkFlows(limit?: number) {
  return getData<NetworkFlow[]>('/network/flows', limit ? { limit: String(limit) } : undefined);
}

export async function fetchNetworkDevices() {
  return getData<NetworkDevice[]>('/network/devices');
}

export async function fetchNetworkTopology() {
  return getData<NetworkTopologyGraph>('/network/topology');
}

export async function fetchNetworkAnomalies() {
  return getData<NetworkAnomaly[]>('/network/anomalies');
}

export interface NetFlowRecord {
  id: string;
  exporter: string;
  srcIp: string;
  dstIp: string;
  protocol: string;
  bytes: number;
  packets: number;
  application: string;
  timestamp: string;
}

export interface SDWANTunnel {
  id: string;
  site: string;
  provider: string;
  latencyMs: number;
  jitterMs: number;
  lossPct: number;
  status: string;
  timestamp: string;
}

export interface WirelessLink {
  id: string;
  apName: string;
  clientMac: string;
  ssid: string;
  rssiDbm: number;
  throughputMbps: number;
  packetLossPct: number;
  timestamp: string;
}

export async function fetchNetFlowRecords() {
  return getData<NetFlowRecord[]>('/network/flows/netflow');
}

export async function fetchSDWANTunnels() {
  return getData<SDWANTunnel[]>('/network/sdwan/tunnels');
}

export async function fetchWirelessLinks() {
  return getData<WirelessLink[]>('/network/wireless/links');
}

export interface AlertPolicyFeedback {
  id: string;
  policyId: string;
  service: string;
  helpful: boolean;
  comment?: string;
  createdAt: string;
}

export async function fetchAlertPolicyFeedback(policyId: string) {
  return getData<AlertPolicyFeedback[]>(`/alerts/policies/${encodeURIComponent(policyId)}/feedback`);
}

export interface RUMFunnelStep {
  name: string;
  event: string;
  count: number;
  conversionPct: number;
}

export interface RUMFunnel {
  id: string;
  name: string;
  steps: RUMFunnelStep[];
  overallConversionPct: number;
  createdAt: string;
  updatedAt: string;
}

export interface SyntheticBrowserTest {
  id: string;
  name: string;
  url: string;
  script: string;
  locations: string[];
  enabled: boolean;
  lastStatus: string;
  createdAt: string;
}

export interface SyntheticMobileTest {
  id: string;
  name: string;
  platform: string;
  bundleId: string;
  script: string;
  enabled: boolean;
  lastStatus: string;
  createdAt: string;
}

export interface SyntheticPrivateLocation {
  id: string;
  name: string;
  region: string;
  agentVersion: string;
  status: string;
  lastHeartbeatAt: string;
}

export interface BusinessKPIDefinition {
  key: string;
  label: string;
  unit: string;
  target: number;
}

export interface BusinessKPIPack {
  id: string;
  name: string;
  category: string;
  description: string;
  kpis: BusinessKPIDefinition[];
  connectors: string[];
  enabled: boolean;
}

export async function fetchRUMFunnels() {
  return getData<RUMFunnel[]>('/rum/funnels');
}

export async function createRUMFunnel(body: { name: string; steps: { name: string; event: string }[] }) {
  return postData<RUMFunnel>('/rum/funnels', body);
}

export async function fetchSyntheticBrowserTests() {
  return getData<SyntheticBrowserTest[]>('/synthetic/browser-tests');
}

export async function createSyntheticBrowserTest(body: {
  name: string;
  url: string;
  script?: string;
  locations?: string[];
  enabled?: boolean;
}) {
  return postData<SyntheticBrowserTest>('/synthetic/browser-tests', body);
}

export async function fetchSyntheticMobileTests() {
  return getData<SyntheticMobileTest[]>('/synthetic/mobile-tests');
}

export async function createSyntheticMobileTest(body: {
  name: string;
  platform?: string;
  bundleId?: string;
  script?: string;
  enabled?: boolean;
}) {
  return postData<SyntheticMobileTest>('/synthetic/mobile-tests', body);
}

export async function fetchSyntheticPrivateLocations() {
  return getData<SyntheticPrivateLocation[]>('/synthetic/private-locations');
}

export async function registerSyntheticPrivateLocation(body: { name: string; region?: string; agentVersion?: string }) {
  return postData<SyntheticPrivateLocation>('/synthetic/private-locations', body);
}

export async function fetchBusinessKPIPacks() {
  return getData<BusinessKPIPack[]>('/business/kpi-packs');
}

export async function enableBusinessKPIPack(id: string) {
  return postData<BusinessKPIPack>(`/business/kpi-packs/${encodeURIComponent(id)}/enable`);
}

export interface ABACPolicyRule {
  id: string;
  effect: string;
  action: string;
  resource: string;
  condition: string;
}

export interface ABACPolicy {
  enabled: boolean;
  rules: ABACPolicyRule[];
  updatedAt: string;
}

export interface DataResidencyPolicy {
  primaryRegion: string;
  allowedRegions: string[];
  piiStorageRegion: string;
  crossBorderDenied: boolean;
  updatedAt: string;
}

export interface BrandingTheme {
  productName: string;
  logoUrl: string;
  primaryColor: string;
  accentColor: string;
  supportEmail: string;
  customDomain: string;
  updatedAt: string;
}

export interface MSPTenant {
  id: string;
  name: string;
  slug: string;
  plan: string;
  status: string;
  userCount: number;
  region: string;
  createdAt: string;
}

export interface ExportJob {
  id: string;
  type: string;
  destination: string;
  status: string;
  rowsExported: number;
  startedAt: string;
  finishedAt?: string;
  message?: string;
}

export async function fetchABACPolicies() {
  return getData<ABACPolicy>('/admin/abac-policies');
}

export async function updateABACPolicies(body: ABACPolicy) {
  return putData<ABACPolicy>('/admin/abac-policies', body);
}

export async function fetchDataResidency() {
  return getData<DataResidencyPolicy>('/admin/data-residency');
}

export async function updateDataResidency(body: DataResidencyPolicy) {
  return putData<DataResidencyPolicy>('/admin/data-residency', body);
}

export async function fetchBranding() {
  return getData<BrandingTheme>('/admin/branding');
}

export async function updateBranding(body: BrandingTheme) {
  return putData<BrandingTheme>('/admin/branding', body);
}

export async function fetchMSPTenants() {
  return getData<MSPTenant[]>('/admin/msp/tenants');
}

export async function createMSPTenant(body: Pick<MSPTenant, 'name' | 'slug' | 'plan' | 'region'>) {
  return postData<MSPTenant>('/admin/msp/tenants', body);
}

export interface LogTierPolicy {
  tenantId: string;
  hotRetentionDays: number;
  warmRetentionDays: number;
  coldRetentionDays: number;
  restoreSlaHours: number;
  updatedAt: string;
}

export async function fetchLogTierPolicy() {
  return getData<LogTierPolicy>('/logs/tiering');
}

export async function updateLogTierPolicy(body: Partial<LogTierPolicy>) {
  return putData<LogTierPolicy>('/logs/tiering', body);
}

export interface ServerlessFunction {
  id: string;
  name: string;
  provider: string;
  runtime: string;
  region: string;
  invocations24h: number;
  errorRatePct: number;
  p95DurationMs: number;
  coldStartPct: number;
  memoryMb: number;
  status: string;
  updatedAt: string;
}

export async function fetchServerlessFunctions() {
  return getData<ServerlessFunction[]>('/infra/serverless/functions');
}

export async function exportToWarehouse(destination: string) {
  return postData<ExportJob>('/exports/warehouse', { destination });
}

export async function exportToBI(destination: string) {
  return postData<ExportJob>('/exports/bi', { destination });
}

export async function exportEvents(destination: string) {
  return postData<ExportJob>('/exports/events', { destination });
}

export interface NFRBenchmark {
  id: string;
  name: string;
  category: string;
  target: string;
  actual: string;
  unit: string;
  pass: boolean;
  measuredAt: string;
}

export interface NFRReliabilityDrill {
  id: string;
  name: string;
  type: string;
  region: string;
  status: string;
  rtoSeconds: number;
  rtoSloSeconds: number;
  pass: boolean;
  executedAt: string;
}

export interface NFRA11yReport {
  standard: string;
  level: string;
  pagesAudited: number;
  violationsCritical: number;
  violationsSerious: number;
  pass: boolean;
  lastAuditAt: string;
}

export interface NFRLocale {
  code: string;
  name: string;
  coveragePct: number;
  enabled: boolean;
}

export interface NFRCertificationReport {
  version: string;
  signedAt: string;
  overallPass: boolean;
  benchmarksPass: boolean;
  reliabilityPass: boolean;
  accessibilityPass: boolean;
  i18nReady: boolean;
  evidenceUris: string[];
}

export async function fetchNFRBenchmarks() {
  return getData<NFRBenchmark[]>('/nfr/benchmarks');
}

export async function fetchNFRReliability() {
  return getData<NFRReliabilityDrill[]>('/nfr/reliability');
}

export async function fetchNFRA11y() {
  return getData<NFRA11yReport>('/nfr/accessibility');
}

export async function fetchNFRLocales() {
  return getData<NFRLocale[]>('/nfr/i18n/locales');
}

export async function fetchNFRCertification() {
  return getData<NFRCertificationReport>('/nfr/certification');
}

export interface DerivedMetricSample {
  metricId: string;
  service: string;
  bucketTs: string;
  value: number;
  sampleCount: number;
}

export interface AlertScoreBucket {
  policyId: string;
  service: string;
  bucketTs: string;
  signalCount: number;
  fatigueScore: number;
}

export async function fetchDerivedMetricSamples(metricId: string) {
  return getData<DerivedMetricSample[]>(`/metrics/derived/${encodeURIComponent(metricId)}/samples`);
}

export async function fetchAlertPolicyScores(policyId: string) {
  return getData<AlertScoreBucket[]>(`/alerts/policies/${encodeURIComponent(policyId)}/scores`);
}

export async function submitAlertPolicyFeedback(policyId: string, body: { service: string; helpful: boolean; comment?: string }) {
  return postData(`/alerts/policies/${encodeURIComponent(policyId)}/feedback`, body);
}

export async function runNFRSlaCertification() {
  return postData<{ runId: string; passed: boolean }>('/nfr/sla/run', {});
}

export async function runNFRBenchmark() {
  return postData<{ runId: string; passed: boolean; p95Ms: number; throughputRps: number }>('/nfr/benchmark/run', {});
}

export async function fetchMultiRegionStatus() {
  return getData<{
    enabled: boolean;
    localRegion: string;
    crossBorderOk: boolean;
    peers: { region: string; gatewayUrl: string; status: string; isPrimary?: boolean }[];
  }>('/admin/regions');
}
