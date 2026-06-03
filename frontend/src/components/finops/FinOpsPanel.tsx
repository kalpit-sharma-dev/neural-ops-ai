import { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  createFinOpsBudget,
  createFinOpsReportSchedule,
  deleteFinOpsAllocationRule,
  fetchFinOpsAllocationRules,
  fetchFinOpsAnomalies,
  fetchFinOpsAnomaly,
  fetchFinOpsAuditLog,
  fetchFinOpsBudgetAlerts,
  fetchFinOpsBudgets,
  fetchFinOpsCarbon,
  fetchFinOpsCarbonRecommendations,
  fetchFinOpsChargebackStatements,
  fetchFinOpsCommitmentAlerts,
  fetchFinOpsCommitmentRecommendations,
  fetchFinOpsCommitments,
  fetchFinOpsCostBreakdown,
  fetchFinOpsCosts,
  fetchFinOpsForecast,
  fetchFinOpsGovernancePolicies,
  fetchFinOpsIngestStatus,
  fetchFinOpsKubernetesCost,
  fetchFinOpsRecommendations,
  fetchFinOpsReports,
  fetchFinOpsScenarios,
  fetchFinOpsTagSuggestions,
  fetchFinOpsUnitEconomics,
  finOpsCarbonAction,
  finOpsRecommendationAction,
  generateFinOpsChargebackStatement,
  runFinOpsIngest,
  runFinOpsScenario,
  submitFinOpsAnomalyFeedback,
} from '../../api/observability';
import { getApiErrorMessage } from '../../api/client';
import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Select } from '../ui/Select';
import { ErrorState, LoadingState } from '../ui/PageStates';

const FINOPS_SUBTABS = ['Overview', 'Allocation', 'Chargeback', 'What-If', 'Anomalies', 'Optimization', 'Budgets', 'Commitments', 'Sustainability', 'Reports'] as const;

function severityVariant(s: string): 'critical' | 'warning' | 'info' {
  if (s === 'high' || s === 'critical') return 'critical';
  if (s === 'medium') return 'warning';
  return 'info';
}

function BreakdownTree({ node, depth = 0 }: { node: { key: string; amountUsd: number; pct: number; children?: typeof node[] }; depth?: number }) {
  return (
    <div style={{ marginLeft: depth * 16, marginBottom: 6 }}>
      <strong>{node.key || 'total'}</strong> — ${node.amountUsd.toFixed(0)} ({node.pct.toFixed(1)}%)
      {(node.children ?? []).map((c, i) => (
        <BreakdownTree key={`${c.key}-${i}`} node={c} depth={depth + 1} />
      ))}
    </div>
  );
}

interface FinOpsPanelProps {
  costScope: string;
  provider: string;
  onScopeChange: (scope: string) => void;
}

export function FinOpsPanel({ costScope, provider, onScopeChange }: FinOpsPanelProps) {
  const [subTab, setSubTab] = useState<(typeof FINOPS_SUBTABS)[number]>('Overview');
  const [breakdownDim, setBreakdownDim] = useState('team');
  const [selectedAnomalyId, setSelectedAnomalyId] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const costsQuery = useQuery({
    queryKey: ['finops-costs', costScope, provider],
    queryFn: () => fetchFinOpsCosts({ scope: costScope, provider: provider || undefined }),
    enabled: subTab === 'Overview' || subTab === 'Budgets',
  });

  const forecastQuery = useQuery({
    queryKey: ['finops-forecast', costScope],
    queryFn: () => fetchFinOpsForecast({ scope: costScope, horizon: 'month' }),
    enabled: subTab === 'Overview',
  });

  const breakdownQuery = useQuery({
    queryKey: ['finops-breakdown', breakdownDim],
    queryFn: () => fetchFinOpsCostBreakdown({ dimension: breakdownDim }),
    enabled: subTab === 'Allocation',
  });

  const rulesQuery = useQuery({
    queryKey: ['finops-allocation-rules'],
    queryFn: fetchFinOpsAllocationRules,
    enabled: subTab === 'Allocation',
  });

  const k8sQuery = useQuery({
    queryKey: ['finops-k8s-cost'],
    queryFn: () => fetchFinOpsKubernetesCost({ cluster: 'prod-cluster' }),
    enabled: subTab === 'Allocation',
  });

  const anomaliesQuery = useQuery({
    queryKey: ['finops-anomalies', costScope],
    queryFn: () => fetchFinOpsAnomalies({ scope: costScope !== 'all' ? costScope : undefined }),
    enabled: subTab === 'Anomalies' || subTab === 'Overview',
  });

  const recsQuery = useQuery({
    queryKey: ['finops-recommendations', costScope],
    queryFn: () => fetchFinOpsRecommendations({ scope: costScope !== 'all' ? costScope : undefined }),
    enabled: subTab === 'Optimization',
  });

  const budgetsQuery = useQuery({
    queryKey: ['finops-budgets'],
    queryFn: fetchFinOpsBudgets,
    enabled: subTab === 'Budgets',
  });

  const budgetAlertsQuery = useQuery({
    queryKey: ['finops-budget-alerts'],
    queryFn: fetchFinOpsBudgetAlerts,
    enabled: subTab === 'Budgets' || subTab === 'Overview',
  });

  const ingestStatusQuery = useQuery({
    queryKey: ['finops-ingest-status'],
    queryFn: fetchFinOpsIngestStatus,
    enabled: subTab === 'Overview',
    refetchInterval: 60_000,
  });

  const anomalyDetailQuery = useQuery({
    queryKey: ['finops-anomaly', selectedAnomalyId],
    queryFn: () => fetchFinOpsAnomaly(selectedAnomalyId!),
    enabled: !!selectedAnomalyId,
  });

  const tagSuggestQuery = useQuery({
    queryKey: ['finops-tag-suggestions'],
    queryFn: fetchFinOpsTagSuggestions,
    enabled: subTab === 'Allocation',
  });

  const commitmentsQuery = useQuery({
    queryKey: ['finops-commitments'],
    queryFn: () => fetchFinOpsCommitments(),
    enabled: subTab === 'Commitments',
  });

  const commitRecsQuery = useQuery({
    queryKey: ['finops-commitment-recs'],
    queryFn: fetchFinOpsCommitmentRecommendations,
    enabled: subTab === 'Commitments',
  });

  const carbonQuery = useQuery({
    queryKey: ['finops-carbon', costScope],
    queryFn: () => fetchFinOpsCarbon({ scope: costScope, dimension: 'team' }),
    enabled: subTab === 'Sustainability',
  });

  const carbonRecsQuery = useQuery({
    queryKey: ['finops-carbon-recs'],
    queryFn: fetchFinOpsCarbonRecommendations,
    enabled: subTab === 'Sustainability',
  });

  const unitEconQuery = useQuery({
    queryKey: ['finops-unit-econ', costScope],
    queryFn: () => fetchFinOpsUnitEconomics({ metric: 'request', scope: costScope }),
    enabled: subTab === 'Overview' || subTab === 'Sustainability',
  });

  const reportsQuery = useQuery({
    queryKey: ['finops-reports'],
    queryFn: () => fetchFinOpsReports(costScope),
    enabled: subTab === 'Reports',
  });

  const auditQuery = useQuery({
    queryKey: ['finops-audit'],
    queryFn: fetchFinOpsAuditLog,
    enabled: subTab === 'Reports',
  });

  const chargebackQuery = useQuery({
    queryKey: ['finops-chargeback', costScope],
    queryFn: () => fetchFinOpsChargebackStatements(costScope === 'all' ? undefined : costScope),
    enabled: subTab === 'Chargeback',
  });

  const scenariosQuery = useQuery({
    queryKey: ['finops-scenarios'],
    queryFn: fetchFinOpsScenarios,
    enabled: subTab === 'What-If',
  });

  const commitmentAlertsQuery = useQuery({
    queryKey: ['finops-commitment-alerts'],
    queryFn: () => fetchFinOpsCommitmentAlerts(),
    enabled: subTab === 'Commitments',
  });

  const governanceQuery = useQuery({
    queryKey: ['finops-governance'],
    queryFn: fetchFinOpsGovernancePolicies,
    enabled: subTab === 'Chargeback',
  });

  const ingestMutation = useMutation({
    mutationFn: runFinOpsIngest,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['finops-costs'] });
      queryClient.invalidateQueries({ queryKey: ['finops-anomalies'] });
      queryClient.invalidateQueries({ queryKey: ['finops-recommendations'] });
      queryClient.invalidateQueries({ queryKey: ['finops-ingest-status'] });
      queryClient.invalidateQueries({ queryKey: ['finops-budget-alerts'] });
    },
  });

  const feedbackMutation = useMutation({
    mutationFn: ({ id, feedback }: { id: string; feedback: 'confirm' | 'false_positive' | 'expected' }) =>
      submitFinOpsAnomalyFeedback(id, feedback),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['finops-anomalies'] }),
  });

  const recActionMutation = useMutation({
    mutationFn: ({ id, action }: { id: string; action: string }) => finOpsRecommendationAction(id, action),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['finops-recommendations'] }),
  });

  const deleteRuleMutation = useMutation({
    mutationFn: deleteFinOpsAllocationRule,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['finops-allocation-rules'] }),
  });

  const createBudgetMutation = useMutation({
    mutationFn: createFinOpsBudget,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['finops-budgets'] }),
  });

  const scheduleReportMutation = useMutation({
    mutationFn: createFinOpsReportSchedule,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['finops-reports'] }),
  });

  const chargebackMutation = useMutation({
    mutationFn: generateFinOpsChargebackStatement,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['finops-chargeback'] }),
  });

  const scenarioMutation = useMutation({
    mutationFn: runFinOpsScenario,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['finops-scenarios'] }),
  });

  const carbonActionMutation = useMutation({
    mutationFn: ({ id, action }: { id: string; action: 'simulate' | 'apply' | 'dismiss' }) =>
      finOpsCarbonAction(id, action),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['finops-carbon-recs'] }),
  });

  const maxCost = useMemo(() => {
    const pts = costsQuery.data?.points ?? [];
    return Math.max(...pts.map((p) => p.amount), 1);
  }, [costsQuery.data]);

  const burnPct = useMemo(() => {
    const d = costsQuery.data;
    if (!d || d.budget <= 0) return 0;
    return Math.min(100, (d.total / d.budget) * 100);
  }, [costsQuery.data]);

  return (
    <>
      <div className="tab-bar" style={{ marginBottom: 16 }}>
        {FINOPS_SUBTABS.map((st) => (
          <button
            key={st}
            type="button"
            className={`tab-bar__item ${subTab === st ? 'tab-bar__item--active' : ''}`}
            onClick={() => setSubTab(st)}
            data-testid={`finops-subtab-${st.toLowerCase()}`}
          >
            {st}
          </button>
        ))}
        <button
          type="button"
          className="tab-bar__item"
          onClick={() => ingestMutation.mutate()}
          disabled={ingestMutation.isPending}
          data-testid="finops-ingest-run"
          style={{ marginLeft: 'auto' }}
        >
          {ingestMutation.isPending ? 'Ingesting…' : 'Run ingest'}
        </button>
      </div>

      <div style={{ marginBottom: 16, maxWidth: 220 }}>
        <Select label="Cost scope" value={costScope} onChange={(e) => onScopeChange(e.target.value)}>
          <option value="all">All</option>
          <option value="payments">Payments</option>
          <option value="platform">Platform</option>
        </Select>
      </div>

      {subTab === 'Overview' && (
        <>
          {costsQuery.isLoading && <LoadingState />}
          {costsQuery.error && (
            <ErrorState message={getApiErrorMessage(costsQuery.error)} onRetry={() => costsQuery.refetch()} />
          )}
          {costsQuery.data && (
            <Card title="Cost trend & forecast" data-testid="finops-costs-card">
              <p>
                Total: <strong>${costsQuery.data.total.toFixed(0)}</strong> {costsQuery.data.unit} · Budget: $
                {costsQuery.data.budget.toFixed(0)}
                {costsQuery.data.tagCoveragePct != null && (
                  <> · Tag coverage: {costsQuery.data.tagCoveragePct.toFixed(0)}%</>
                )}
              </p>
              <div
                className="metric-sparkline"
                style={{ display: 'flex', alignItems: 'flex-end', gap: 2, height: 80, marginTop: 12 }}
              >
                {costsQuery.data.points.map((p) => (
                  <div
                    key={p.timestamp}
                    title={`${p.timestamp}: $${p.amount}`}
                    style={{
                      flex: 1,
                      background: 'var(--accent)',
                      height: `${Math.min(100, (p.amount / maxCost) * 100)}%`,
                      minHeight: 2,
                    }}
                  />
                ))}
              </div>
              {(forecastQuery.data ?? costsQuery.data.forecast) && (
                <p className="muted" style={{ marginTop: 12 }}>
                  Forecast ({(forecastQuery.data ?? costsQuery.data.forecast)!.horizon}): P50 $
                  {(forecastQuery.data ?? costsQuery.data.forecast)!.p50.toFixed(0)} · band $
                  {(forecastQuery.data ?? costsQuery.data.forecast)!.lower.toFixed(0)}–$
                  {(forecastQuery.data ?? costsQuery.data.forecast)!.upper.toFixed(0)}
                </p>
              )}
              {unitEconQuery.data && (
                <p className="muted" style={{ marginTop: 8 }}>
                  Unit economics: ${unitEconQuery.data.costPerUnit.toFixed(6)}/request ·{' '}
                  {unitEconQuery.data.totalUnits.toLocaleString()} requests/30d
                </p>
              )}
              <div style={{ marginTop: 12, background: 'var(--surface-muted)', borderRadius: 4, overflow: 'hidden' }}>
                <div
                  style={{
                    width: `${burnPct}%`,
                    height: 8,
                    background: burnPct >= 90 ? 'var(--danger)' : burnPct >= 70 ? 'var(--warning)' : 'var(--accent)',
                  }}
                  data-testid="finops-budget-burn"
                />
              </div>
              <p className="muted" style={{ marginTop: 4 }}>
                Budget burn: {burnPct.toFixed(0)}%
              </p>
              {ingestStatusQuery.data && (
                <p className="muted" style={{ marginTop: 8 }} data-testid="finops-ingest-status">
                  Billing ingest:{' '}
                  {ingestStatusQuery.data.stale ? (
                    <Badge variant="warning">stale ({Math.round(ingestStatusQuery.data.lagSeconds / 3600)}h lag)</Badge>
                  ) : (
                    <Badge variant="healthy">fresh</Badge>
                  )}
                  {ingestStatusQuery.data.providers.length > 0 && (
                    <> · providers: {ingestStatusQuery.data.providers.join(', ')}</>
                  )}
                </p>
              )}
              {(budgetAlertsQuery.data ?? []).length > 0 && (
                <div style={{ marginTop: 12 }} data-testid="finops-budget-alerts-preview">
                  {(budgetAlertsQuery.data ?? []).slice(0, 2).map((a) => (
                    <p key={a.id} className="muted">
                      <Badge variant={severityVariant(a.severity)}>{a.threshold}%</Badge> {a.message}
                    </p>
                  ))}
                </div>
              )}
            </Card>
          )}
          {(anomaliesQuery.data ?? []).length > 0 && (
            <Card title="Recent anomalies" style={{ marginTop: 24 }} data-testid="finops-anomalies-preview">
              {(anomaliesQuery.data ?? []).slice(0, 3).map((a) => (
                <div key={a.id} style={{ marginBottom: 8 }}>
                  <button type="button" className="link-button" onClick={() => setSelectedAnomalyId(a.id)}>
                    <Badge variant={severityVariant(a.severity)}>{a.severity}</Badge> {a.service} (+{a.deltaPct}%)
                  </button>
                </div>
              ))}
            </Card>
          )}
        </>
      )}

      {subTab === 'Allocation' && (
        <>
          <div style={{ marginBottom: 16, maxWidth: 220 }}>
            <Select label="Breakdown dimension" value={breakdownDim} onChange={(e) => setBreakdownDim(e.target.value)}>
              <option value="team">Team</option>
              <option value="environment">Environment</option>
              <option value="provider">Provider</option>
              <option value="service">Service</option>
            </Select>
          </div>
          {breakdownQuery.isLoading && <LoadingState />}
          {breakdownQuery.data && (
            <Card title="Cost breakdown" data-testid="finops-breakdown-card">
              <BreakdownTree node={breakdownQuery.data} />
            </Card>
          )}
          <Card title="Allocation rules" style={{ marginTop: 24 }} data-testid="finops-rules-card">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Dimension</th>
                  <th>Tag</th>
                  <th>Priority</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {(rulesQuery.data ?? []).map((r) => (
                  <tr key={r.id}>
                    <td>{r.name}</td>
                    <td>{r.dimension}</td>
                    <td>
                      {r.tagKey}
                      {r.tagValue ? `=${r.tagValue}` : ''}
                    </td>
                    <td>{r.priority}</td>
                    <td>
                      <button type="button" onClick={() => deleteRuleMutation.mutate(r.id)}>
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
          <Card title="Tag suggestions (ML-assisted)" style={{ marginTop: 24 }} data-testid="finops-tag-suggestions-card">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Resource</th>
                  <th>Suggested tag</th>
                  <th>Confidence</th>
                  <th>Spend</th>
                </tr>
              </thead>
              <tbody>
                {(tagSuggestQuery.data ?? []).map((t) => (
                  <tr key={t.id}>
                    <td>{t.resourceId}</td>
                    <td>
                      {t.suggestedKey}={t.suggestedValue}
                    </td>
                    <td>{(t.confidence * 100).toFixed(0)}%</td>
                    <td>${t.spendUsd.toFixed(0)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
          <Card title="Kubernetes cost" style={{ marginTop: 24 }} data-testid="finops-k8s-card">
            {k8sQuery.isLoading && <LoadingState />}
            <table className="data-table">
              <thead>
                <tr>
                  <th>Namespace</th>
                  <th>Workload</th>
                  <th>CPU</th>
                  <th>Mem</th>
                  <th>Idle</th>
                  <th>Total</th>
                </tr>
              </thead>
              <tbody>
                {(k8sQuery.data ?? []).map((r, i) => (
                  <tr key={`${r.namespace}-${r.workload}-${i}`}>
                    <td>{r.namespace}</td>
                    <td>{r.workload}</td>
                    <td>${r.cpuCostUsd.toFixed(0)}</td>
                    <td>${r.memCostUsd.toFixed(0)}</td>
                    <td>${r.idleCostUsd.toFixed(0)}</td>
                    <td>${r.totalUsd.toFixed(0)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
        </>
      )}

      {subTab === 'Chargeback' && (
        <>
          <Card title="Showback / Chargeback statements" data-testid="finops-chargeback-card">
            {chargebackQuery.isLoading && <LoadingState />}
            {(governanceQuery.data ?? [])[0] && (
              <p className="muted" style={{ marginBottom: 12 }}>
                Governance mode: {(governanceQuery.data ?? [])[0].chargebackMode} · residency{' '}
                {(governanceQuery.data ?? [])[0].residencyRegion}
              </p>
            )}
            <button
              type="button"
              data-testid="finops-chargeback-generate"
              onClick={() =>
                chargebackMutation.mutate({
                  costCenter: costScope === 'all' ? 'payments' : costScope,
                  mode: (governanceQuery.data ?? [])[0]?.chargebackMode ?? 'showback',
                  period: 'monthly',
                })
              }
            >
              Generate statement
            </button>
            <table className="data-table" style={{ marginTop: 16 }}>
              <thead>
                <tr>
                  <th>Cost center</th>
                  <th>Mode</th>
                  <th>Total</th>
                  <th>Generated</th>
                </tr>
              </thead>
              <tbody>
                {(chargebackQuery.data ?? []).map((st) => (
                  <tr key={st.id}>
                    <td>{st.costCenter}</td>
                    <td>
                      <Badge variant="info">{st.mode}</Badge>
                    </td>
                    <td>${st.totalUsd.toFixed(0)}</td>
                    <td>{new Date(st.generatedAt).toLocaleString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            {(chargebackQuery.data ?? [])[0]?.lines?.map((l) => (
              <div key={l.team} style={{ marginTop: 8 }}>
                {l.team}: ${l.amountUsd.toFixed(0)} ({l.allocatedPct.toFixed(1)}%)
              </div>
            ))}
          </Card>
        </>
      )}

      {subTab === 'What-If' && (
        <>
          <Card title="Scenario modeling" data-testid="finops-scenarios-card">
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginBottom: 16 }}>
              <button
                type="button"
                onClick={() =>
                  scenarioMutation.mutate({
                    name: 'Region migration',
                    scenarioType: 'region_migration',
                    scope: costScope,
                    params: { from_region: 'us-east-1', to_region: 'us-west-2' },
                  })
                }
              >
                Region migration
              </button>
              <button
                type="button"
                onClick={() =>
                  scenarioMutation.mutate({
                    name: 'Instance family change',
                    scenarioType: 'instance_family',
                    scope: costScope,
                    params: { target_family: 'm7g' },
                  })
                }
              >
                Instance family
              </button>
              <button
                type="button"
                onClick={() =>
                  scenarioMutation.mutate({
                    name: 'Commitment purchase',
                    scenarioType: 'commitment_purchase',
                    scope: costScope,
                    params: { term_months: '12' },
                  })
                }
              >
                Commitment purchase
              </button>
            </div>
            {scenariosQuery.isLoading && <LoadingState />}
            <table className="data-table">
              <thead>
                <tr>
                  <th>Scenario</th>
                  <th>Cost Δ</th>
                  <th>CO₂e Δ</th>
                  <th>Summary</th>
                </tr>
              </thead>
              <tbody>
                {(scenariosQuery.data ?? []).map((s) => (
                  <tr key={s.id}>
                    <td>{s.name}</td>
                    <td>${s.costDeltaUsd.toFixed(0)}</td>
                    <td>{s.co2eDeltaKg.toFixed(0)} kg</td>
                    <td>{s.summary}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
        </>
      )}

      {subTab === 'Anomalies' && (
        <>
          <Card title="Cost anomalies" data-testid="finops-anomalies-card">
            {anomaliesQuery.isLoading && <LoadingState />}
            {(anomaliesQuery.data ?? []).map((a) => (
              <div key={a.id} style={{ marginBottom: 16, paddingBottom: 12, borderBottom: '1px solid var(--border)' }}>
                <button type="button" className="link-button" onClick={() => setSelectedAnomalyId(a.id)}>
                  <Badge variant={severityVariant(a.severity)}>{a.severity}</Badge>{' '}
                  <strong>{a.service}</strong> (+{a.deltaPct}% · ${a.amountUsd})
                </button>
              {a.status && (
                <>
                  {' '}
                  <Badge variant="info">{a.status}</Badge>
                </>
              )}
              <p className="muted">{a.description}</p>
              {a.probableCause && <p className="muted">Probable cause: {a.probableCause}</p>}
              <div style={{ display: 'flex', gap: 8, marginTop: 8, flexWrap: 'wrap' }}>
                {(['confirm', 'false_positive', 'expected'] as const).map((fb) => (
                  <button
                    key={fb}
                    type="button"
                    disabled={feedbackMutation.isPending}
                    onClick={() => feedbackMutation.mutate({ id: a.id, feedback: fb })}
                    data-testid={`finops-anomaly-feedback-${fb}`}
                  >
                    {fb.replace('_', ' ')}
                  </button>
                ))}
              </div>
            </div>
          ))}
          </Card>
          {selectedAnomalyId && (
            <Card title="Anomaly detail" style={{ marginTop: 24 }} data-testid="finops-anomaly-drawer">
              {anomalyDetailQuery.isLoading && <LoadingState />}
              {anomalyDetailQuery.data && (
                <>
                  <p>
                    <Badge variant={severityVariant(anomalyDetailQuery.data.severity)}>
                      {anomalyDetailQuery.data.severity}
                    </Badge>{' '}
                    <strong>{anomalyDetailQuery.data.service}</strong> · {anomalyDetailQuery.data.provider}
                  </p>
                  <p>{anomalyDetailQuery.data.description}</p>
                  <p className="muted">
                    Detected {new Date(anomalyDetailQuery.data.detectedAt).toLocaleString()} · +$
                    {anomalyDetailQuery.data.amountUsd} ({anomalyDetailQuery.data.deltaPct}%)
                  </p>
                  {anomalyDetailQuery.data.probableCause && (
                    <p>
                      <strong>Probable cause:</strong> {anomalyDetailQuery.data.probableCause}
                    </p>
                  )}
                  {anomalyDetailQuery.data.feedback && (
                    <p className="muted">Feedback: {anomalyDetailQuery.data.feedback}</p>
                  )}
                  <div style={{ display: 'flex', gap: 8, marginTop: 8, flexWrap: 'wrap' }}>
                    {(['confirm', 'false_positive', 'expected'] as const).map((fb) => (
                      <button
                        key={fb}
                        type="button"
                        disabled={feedbackMutation.isPending}
                        onClick={() => feedbackMutation.mutate({ id: selectedAnomalyId, feedback: fb })}
                      >
                        {fb.replace('_', ' ')}
                      </button>
                    ))}
                    <button type="button" onClick={() => setSelectedAnomalyId(null)}>
                      Close
                    </button>
                  </div>
                </>
              )}
            </Card>
          )}
        </>
      )}

      {subTab === 'Optimization' && (
        <Card title="Recommendations" data-testid="finops-recommendations-card">
          {recsQuery.isLoading && <LoadingState />}
          <table className="data-table">
            <thead>
              <tr>
                <th>Type</th>
                <th>Title</th>
                <th>Projected</th>
                <th>Realized</th>
                <th>Status</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {(recsQuery.data ?? []).map((r) => (
                <tr key={r.id}>
                  <td>
                    <Badge variant="info">{r.type}</Badge>
                  </td>
                  <td>{r.title}</td>
                  <td>${r.projectedSavingsUsd.toFixed(0)}</td>
                  <td>${(r.realizedSavingsUsd ?? 0).toFixed(0)}</td>
                  <td>
                    {r.status}
                    {r.ticketId && (
                      <span className="muted" style={{ display: 'block', fontSize: 12 }}>
                        {r.ticketId}
                      </span>
                    )}
                  </td>
                  <td style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
                    <button type="button" onClick={() => recActionMutation.mutate({ id: r.id, action: 'acknowledge' })}>
                      Ack
                    </button>
                    <button type="button" onClick={() => recActionMutation.mutate({ id: r.id, action: 'ticket' })}>
                      Ticket
                    </button>
                    <button type="button" onClick={() => recActionMutation.mutate({ id: r.id, action: 'autofix' })}>
                      AutoFix
                    </button>
                    <button type="button" onClick={() => recActionMutation.mutate({ id: r.id, action: 'realized' })}>
                      Realized
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}

      {subTab === 'Budgets' && (
        <>
          <Card title="Budgets" data-testid="finops-budgets-card">
            {budgetsQuery.isLoading && <LoadingState />}
            <table className="data-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Scope</th>
                  <th>Amount</th>
                  <th>Spend</th>
                  <th>Burn</th>
                </tr>
              </thead>
              <tbody>
                {(budgetsQuery.data ?? []).map((b) => (
                  <tr key={b.id}>
                    <td>{b.name}</td>
                    <td>
                      {b.scopeType}/{b.scopeValue}
                    </td>
                    <td>${b.amountUsd.toFixed(0)}</td>
                    <td>${(b.spendUsd ?? 0).toFixed(0)}</td>
                    <td>{(b.burnPct ?? 0).toFixed(0)}%</td>
                  </tr>
                ))}
              </tbody>
            </table>
            <button
              type="button"
              style={{ marginTop: 12 }}
              data-testid="finops-budget-create"
              onClick={() =>
                createBudgetMutation.mutate({
                  name: `Budget ${Date.now()}`,
                  scopeType: 'team',
                  scopeValue: costScope === 'all' ? 'platform' : costScope,
                  period: 'monthly',
                  amountUsd: costsQuery.data?.forecast?.p50 ?? 10000,
                  thresholds: [50, 80, 100],
                })
              }
            >
              Add budget from forecast
            </button>
          </Card>
          {(budgetAlertsQuery.data ?? []).length > 0 && (
            <Card title="Budget alerts" style={{ marginTop: 24 }} data-testid="finops-budget-alerts-card">
              {(budgetAlertsQuery.data ?? []).map((a) => (
                <div key={a.id} style={{ marginBottom: 12 }}>
                  <Badge variant={severityVariant(a.severity)}>{a.threshold}%</Badge> {a.message}
                </div>
              ))}
            </Card>
          )}
          {costsQuery.data && (
            <Card title="Forecast vs budget" style={{ marginTop: 24 }}>
              <p>
                Current spend: ${costsQuery.data.total.toFixed(0)} · Budget: ${costsQuery.data.budget.toFixed(0)} ·
                Forecast P50: ${(costsQuery.data.forecast?.p50 ?? 0).toFixed(0)}
              </p>
            </Card>
          )}
        </>
      )}

      {subTab === 'Commitments' && (
        <>
          <Card title="Commitment coverage" data-testid="finops-commitments-card">
            {commitmentsQuery.isLoading && <LoadingState />}
            <table className="data-table">
              <thead>
                <tr>
                  <th>Provider</th>
                  <th>Type</th>
                  <th>Region</th>
                  <th>Coverage</th>
                  <th>Utilization</th>
                  <th>Monthly</th>
                  <th>Expires</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {(commitmentsQuery.data ?? []).map((c) => (
                  <tr key={c.id}>
                    <td>{c.provider}</td>
                    <td>{c.commitmentType}</td>
                    <td>{c.region}</td>
                    <td>{c.coveragePct}%</td>
                    <td>{c.utilizationPct}%</td>
                    <td>${c.monthlyCommitUsd.toFixed(0)}</td>
                    <td>{new Date(c.expiresAt).toLocaleDateString()}</td>
                    <td>
                      <Badge variant={c.status === 'expiring' ? 'warning' : 'healthy'}>{c.status}</Badge>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
          <Card title="Purchase recommendations" style={{ marginTop: 24 }} data-testid="finops-commitment-recs-card">
            <ul className="insight-list">
              {(commitRecsQuery.data ?? []).map((r) => (
                <li key={r.id}>
                  <strong>{r.provider}</strong> {r.commitmentType} ({r.termMonths}mo) — save ${r.monthlySavingsUsd}/mo ·
                  break-even {r.breakEvenMonths}mo — {r.description}
                </li>
              ))}
            </ul>
          </Card>
          <Card title="Commitment alerts" style={{ marginTop: 24 }} data-testid="finops-commitment-alerts-card">
            {(commitmentAlertsQuery.data ?? []).map((a) => (
              <div key={a.id} style={{ marginBottom: 12 }}>
                <Badge variant={severityVariant(a.severity)}>{a.alertType}</Badge> {a.message}
              </div>
            ))}
          </Card>
        </>
      )}

      {subTab === 'Sustainability' && (
        <>
          {carbonQuery.isLoading && <LoadingState />}
          {carbonQuery.data && (
            <Card title="Carbon footprint" data-testid="finops-carbon-card">
              <p>
                {carbonQuery.data.co2eKg} kg CO₂e ({carbonQuery.data.period}) · {carbonQuery.data.renewablePct}%
                renewable
              </p>
              {carbonQuery.data.methodology && (
                <p className="muted">
                  {carbonQuery.data.methodology} · factors v{carbonQuery.data.factorVersion}
                </p>
              )}
              {carbonQuery.data.sci && (
                <p className="muted">
                  SCI: {carbonQuery.data.sci.co2ePerRequest.toFixed(6)} g CO₂e/request ·{' '}
                  {carbonQuery.data.sci.co2ePerTransaction.toFixed(4)} g CO₂e/transaction
                </p>
              )}
              {(carbonQuery.data.byDimension ?? []).map((d) => (
                <div key={d.key} style={{ marginTop: 8 }}>
                  {d.key}: {d.co2eKg} kg ({d.pct}%)
                </div>
              ))}
            </Card>
          )}
          <Card title="Carbon reduction recommendations" style={{ marginTop: 24 }} data-testid="finops-carbon-recs-card">
            {(carbonRecsQuery.data ?? []).map((r) => (
              <div key={r.id} style={{ marginBottom: 12 }}>
                <strong>{r.title}</strong> — −{r.co2eReductionKg} kg CO₂e · cost Δ ${r.costDeltaUsd}/mo
                <p className="muted">{r.description}</p>
                <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
                  {(['simulate', 'apply', 'dismiss'] as const).map((action) => (
                    <button
                      key={action}
                      type="button"
                      disabled={carbonActionMutation.isPending}
                      onClick={() => carbonActionMutation.mutate({ id: r.id, action })}
                    >
                      {action}
                    </button>
                  ))}
                </div>
              </div>
            ))}
          </Card>
        </>
      )}

      {subTab === 'Reports' && (
        <>
          <Card title="Generated reports" data-testid="finops-reports-card">
            {reportsQuery.isLoading && <LoadingState />}
            <table className="data-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Format</th>
                  <th>Scope</th>
                  <th>Generated</th>
                </tr>
              </thead>
              <tbody>
                {(reportsQuery.data ?? []).map((r) => (
                  <tr key={r.id}>
                    <td>{r.name}</td>
                    <td>{r.format}</td>
                    <td>{r.scope}</td>
                    <td>{new Date(r.generatedAt).toLocaleString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            <button
              type="button"
              style={{ marginTop: 12 }}
              data-testid="finops-report-schedule"
              onClick={() =>
                scheduleReportMutation.mutate({
                  name: `Report ${Date.now()}`,
                  scope: costScope,
                  format: 'csv',
                  cadence: 'weekly',
                  deliveryChannel: 'email',
                  deliveryTarget: 'finops@neuralops.ai',
                  enabled: true,
                })
              }
            >
              Schedule weekly report
            </button>
          </Card>
          <Card title="Audit log" style={{ marginTop: 24 }} data-testid="finops-audit-card">
            {auditQuery.isLoading && <LoadingState />}
            <table className="data-table">
              <thead>
                <tr>
                  <th>Time</th>
                  <th>Actor</th>
                  <th>Action</th>
                  <th>Entity</th>
                </tr>
              </thead>
              <tbody>
                {(auditQuery.data ?? []).map((e) => (
                  <tr key={e.id}>
                    <td>{new Date(e.createdAt).toLocaleString()}</td>
                    <td>{e.actor}</td>
                    <td>{e.action}</td>
                    <td>
                      {e.entityType}/{e.entityId}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
        </>
      )}
    </>
  );
}
