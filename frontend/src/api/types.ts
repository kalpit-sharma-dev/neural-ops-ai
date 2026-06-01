export interface ApiResponse<T> {
  status: string;
  data: T;
  timestamp: string;
}

export interface ApiError {
  status: string;
  errorCode: string;
  message: string;
  timestamp: string;
}

export interface PlatformInfo {
  service: string;
  environment: string;
  version: string;
}

export interface HealthStatus {
  healthy: boolean;
}

export interface IncidentSummary {
  id: string;
  title: string;
  severity: string;
  status: string;
  service?: string;
}

export interface DashboardOverview {
  activeIncidents: IncidentSummary[];
  errorRateLastHour: number;
  serviceHealth: ServiceHealthScore[];
  topFailingServices: FailingService[];
  recentAnomalies: AnomalySummary[];
  deploymentImpactScore: number;
  latencyP99Ms?: number;
  mttrMinutes?: number;
  generatedAt: string;
}

export interface ServiceHealthScore {
  service: string;
  score: number;
}

export interface FailingService {
  service: string;
  errorCount: number;
}

export interface AnomalySummary {
  service: string;
  message: string;
  severity: string;
  timestamp: string;
}

export interface Incident {
  id: string;
  title: string;
  summary: string;
  severity: string;
  status: string;
  affectedServices: string[];
  rootCauseAnalysis?: RootCause;
  blastRadius?: string[];
  startTime: string;
  resolvedTime?: string;
  acknowledgedAt?: string;
  mttr?: number;
  recommendations?: Recommendation[];
  timeline?: IncidentEvent[];
  createdAt?: string;
  updatedAt?: string;
}

export interface RootCause {
  firstFailingService: string;
  rootCauseDescription: string;
  evidence?: Evidence[];
  deploymentCorrelation?: DeploymentCorrelation;
  confidence: number;
}

export interface Evidence {
  id: string;
  type: string;
  description: string;
  snippet?: string;
  timestamp: string;
}

export interface DeploymentCorrelation {
  deploymentId: string;
  service: string;
  version: string;
  deployedAt: string;
  correlationScore: number;
}

export interface Recommendation {
  type: string;
  description: string;
  codeSnippet?: string;
  priority: string;
}

export interface IncidentEvent {
  id?: string;
  eventType: string;
  description: string;
  timestamp: string;
  service?: string;
}

export interface IncidentTimelineResponse {
  timeline: IncidentEvent[];
  narrative: string;
}

export interface LogSearchRequest {
  query?: string;
  messageRegex?: string;
  service?: string;
  severity?: string;
  traceId?: string;
  txnId?: string;
  host?: string;
  pod?: string;
  startTime?: string;
  endTime?: string;
  size?: number;
  searchAfter?: unknown[];
}

export interface LogHit {
  id: string;
  score: number;
  timestamp: string;
  service: string;
  severity: string;
  message: string;
  traceId?: string;
  txnId?: string;
  host?: string;
  pod?: string;
  plainEnglish?: string;
  classification?: string;
  labels?: Record<string, string>;
  source?: string;
}

export interface SearchResponse {
  hits: LogHit[];
  total: number;
  searchAfter?: unknown[];
  tookMs: number;
  aggregations?: {
    byService: Bucket[];
    bySeverity: Bucket[];
    byHour: Bucket[];
  };
}

export interface Bucket {
  key: string;
  count: number;
}

export interface AISearchResponse {
  parsed: {
    query?: string;
    service?: string;
    severity?: string;
    explanation?: string;
  };
  results: SearchResponse;
}

export interface DependencyEdge {
  target: string;
  callCount: number;
  p99LatencyMs: number;
}

export type DependencyMap = Record<string, DependencyEdge[]>;

export interface Transaction {
  txnId: string;
  txnType: string;
  status: string;
  hops: TxnHop[];
  totalLatencyMs: number;
  failedAt?: string;
  retryCount: number;
  startedAt?: string;
  completedAt?: string;
}

export interface TxnHop {
  serviceName: string;
  spanId: string;
  traceId: string;
  latencyMs: number;
  status: string;
  errorMessage?: string;
  startedAt?: string;
  completedAt?: string;
}

export interface RealtimeEvent {
  type: string;
  service?: string;
  severity?: string;
  message: string;
  timestamp: string;
}
