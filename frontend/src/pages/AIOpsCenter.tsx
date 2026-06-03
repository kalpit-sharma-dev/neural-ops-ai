import { useState } from 'react';
import { Link } from '@tanstack/react-router';
import { useMutation, useQuery } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import {
  createAutoFixPlan,
  executeAutoFix,
  fetchIncidentRCA,
  postAIForecast,
  rollbackAutoFix,
  type AIExplanationNode,
  type AutoFixActionRecord,
  type AutoFixPlan,
} from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { StitchPageShell } from '../components/stitch';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Badge } from '../components/ui/Badge';
import { LoadingState } from '../components/ui/PageStates';

function ConfidenceBar({ value }: { value: number }) {
  const pct = Math.round(value * 100);
  return (
    <div className="confidence-bar" style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
      <div style={{ flex: 1, height: 6, background: 'var(--border-subtle)', borderRadius: 4 }}>
        <div style={{ width: `${pct}%`, height: '100%', background: 'var(--accent-primary)', borderRadius: 4 }} />
      </div>
      <span style={{ fontSize: 12, color: 'var(--text-muted)' }}>{pct}%</span>
    </div>
  );
}

function ExplanationTree({ node, depth = 0 }: { node: AIExplanationNode; depth?: number }) {
  return (
    <div style={{ marginLeft: depth * 16, marginTop: depth ? 12 : 0 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
        <strong>{node.label}</strong>
        <Badge variant="info">{node.id}</Badge>
      </div>
      <ConfidenceBar value={node.confidence} />
      {node.evidence?.length ? (
        <ul className="insight-list" style={{ marginTop: 8 }}>
          {node.evidence.map((ev, i) => (
            <li key={i}>
              <Badge variant="info">{ev.signal}</Badge> {ev.ref}: {ev.detail}
            </li>
          ))}
        </ul>
      ) : null}
      {node.children?.map((child) => (
        <ExplanationTree key={child.id} node={child} depth={depth + 1} />
      ))}
    </div>
  );
}

export default function AIOpsCenter() {
  const [incidentId, setIncidentId] = useState('inc-1');
  const [forecastMetric, setForecastMetric] = useState('cpu.utilization');
  const [forecastService, setForecastService] = useState('payment-service');
  const [forecastHorizon, setForecastHorizon] = useState('24h');
  const [plan, setPlan] = useState<AutoFixPlan | null>(null);
  const [action, setAction] = useState<AutoFixActionRecord | null>(null);

  const rcaQuery = useQuery({
    queryKey: ['ai-rca', incidentId],
    queryFn: () => fetchIncidentRCA(incidentId),
    enabled: false,
  });

  const forecastMut = useMutation({
    mutationFn: () => postAIForecast({ metric: forecastMetric, service: forecastService, horizon: forecastHorizon }),
    onSuccess: () => toast.success('Forecast generated'),
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const planMut = useMutation({
    mutationFn: () => createAutoFixPlan(incidentId),
    onSuccess: (p) => {
      setPlan(p);
      setAction(null);
      toast.success('AutoFix plan created');
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const executeMut = useMutation({
    mutationFn: () => {
      if (!plan) throw new Error('Create a plan first');
      return executeAutoFix(plan.id, true);
    },
    onSuccess: (a) => {
      setAction(a);
      toast.success('AutoFix executed');
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const rollbackMut = useMutation({
    mutationFn: () => {
      if (!action) throw new Error('No action to roll back');
      return rollbackAutoFix(action.id);
    },
    onSuccess: (a) => {
      setAction(a);
      toast.success('Rollback completed');
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  return (
    <StitchPageShell
      title="AI Ops Center"
      subtitle="Explainable RCA, capacity forecasting, and guarded AutoFix remediation."
    >
      <div style={{ display: 'grid', gap: 24 }}>
        <Card title="Root cause analysis" data-testid="aiops-rca-card">
          <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', marginBottom: 16 }}>
            <Input
              label="Incident ID"
              value={incidentId}
              onChange={(e) => setIncidentId(e.target.value)}
              data-testid="aiops-incident-id"
            />
            <Button
              data-testid="aiops-load-rca"
              onClick={() => rcaQuery.refetch()}
              disabled={!incidentId || rcaQuery.isFetching}
            >
              Load RCA
            </Button>
            <Link to="/incidents/$id" params={{ id: incidentId }} className="btn-link">
              Open incident
            </Link>
          </div>
          {rcaQuery.isFetching && <LoadingState />}
          {rcaQuery.data && (
            <>
              <p>{rcaQuery.data.summary}</p>
              <ConfidenceBar value={rcaQuery.data.confidence} />
              <p style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 8 }}>
                Explanation: {rcaQuery.data.explanationId}
              </p>
              <ExplanationTree node={rcaQuery.data.root} />
            </>
          )}
        </Card>

        <Card title="Forecast" data-testid="aiops-forecast-card">
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: 12 }}>
            <Input label="Metric" value={forecastMetric} onChange={(e) => setForecastMetric(e.target.value)} />
            <Input label="Service" value={forecastService} onChange={(e) => setForecastService(e.target.value)} />
            <Input label="Horizon" value={forecastHorizon} onChange={(e) => setForecastHorizon(e.target.value)} />
          </div>
          <Button
            data-testid="aiops-run-forecast"
            style={{ marginTop: 12 }}
            onClick={() => forecastMut.mutate()}
            disabled={forecastMut.isPending}
          >
            Run forecast
          </Button>
          {forecastMut.data && (
            <div style={{ marginTop: 16 }}>
              <p>
                Prediction: <strong>{forecastMut.data.prediction}</strong> {forecastMut.data.unit} (
                {forecastMut.data.lowerBound}–{forecastMut.data.upperBound})
              </p>
              {forecastMut.data.recommendation && <p>{forecastMut.data.recommendation}</p>}
            </div>
          )}
        </Card>

        <Card title="AutoFix" data-testid="aiops-autofix-card">
          <p style={{ marginBottom: 12 }}>
            Plans with high blast-radius steps require explicit approval before execution.
          </p>
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
            <Button data-testid="aiops-create-plan" onClick={() => planMut.mutate()} disabled={planMut.isPending}>
              Create plan
            </Button>
            <Button
              data-testid="aiops-execute-plan"
              variant="primary"
              onClick={() => executeMut.mutate()}
              disabled={!plan || executeMut.isPending}
            >
              Execute (approved)
            </Button>
            <Button
              data-testid="aiops-rollback"
              variant="secondary"
              onClick={() => rollbackMut.mutate()}
              disabled={!action || rollbackMut.isPending}
            >
              Rollback
            </Button>
          </div>
          {plan && (
            <div style={{ marginTop: 16 }}>
              <p>
                <strong>{plan.summary}</strong>
              </p>
              <Badge variant={plan.policyPass ? 'success' : 'warning'}>
                {plan.policyPass ? 'Policy pass' : 'Approval required'}
              </Badge>
              {plan.policyReason && <p style={{ marginTop: 8 }}>{plan.policyReason}</p>}
              <ul className="insight-list" style={{ marginTop: 12 }}>
                {plan.steps.map((s) => (
                  <li key={s.id}>
                    {s.action}: {s.description}{' '}
                    <Badge variant={s.blastRadius === 'high' ? 'critical' : 'info'}>{s.blastRadius}</Badge>
                  </li>
                ))}
              </ul>
            </div>
          )}
          {action && (
            <div style={{ marginTop: 16 }}>
              <p>
                Action <code>{action.id}</code> — {action.status}
              </p>
              <ul className="insight-list">
                {action.logs.map((line, i) => (
                  <li key={i}>{line}</li>
                ))}
              </ul>
            </div>
          )}
        </Card>
      </div>
    </StitchPageShell>
  );
}
