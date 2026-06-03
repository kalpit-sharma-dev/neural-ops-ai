import { useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate, useParams } from '@tanstack/react-router';
import { formatDistanceToNow, intervalToDuration } from 'date-fns';
import toast from 'react-hot-toast';
import { Share2, TrendingUp } from 'lucide-react';
import {
  acknowledgeIncident,
  fetchIncident,
  fetchIncidentTimeline,
  fetchRecommendations,
  resolveIncident,
} from '../api/incidents';
import { fetchIncidentRCA, triggerWorkflows } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { UnifiedContextPanel } from '../components/incident/UnifiedContextPanel';
import { IncidentLogsTab } from '../components/incident/IncidentLogsTab';
import { IncidentMetricsTab } from '../components/incident/IncidentMetricsTab';
import { IncidentTracesTab } from '../components/incident/IncidentTracesTab';
import { TimelineView } from '../components/incident/TimelineView';
import { TypingText } from '../components/incident/TypingText';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { CodeBlock } from '../components/ui/CodeBlock';
import { Input } from '../components/ui/Input';
import { Textarea } from '../components/ui/Textarea';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { StatusDot } from '../components/ui/StatusDot';
import { useRealtimeStore } from '../store/realtimeStore';

const TABS = ['Overview', 'Timeline', 'Logs', 'Traces', 'Metrics', 'AI Analysis', 'Recommendations'] as const;

function severityBadge(sev: string) {
  const map: Record<string, 'p1' | 'p2' | 'p3' | 'p4'> = {
    P1: 'p1',
    P2: 'p2',
    P3: 'p3',
    P4: 'p4',
  };
  return map[sev] ?? 'p4';
}

function formatMttr(start: string, resolved?: string) {
  const end = resolved ? new Date(resolved) : new Date();
  const dur = intervalToDuration({ start: new Date(start), end });
  const parts = [dur.hours, dur.minutes, dur.seconds].map((n) => String(n ?? 0).padStart(2, '0'));
  return `${parts[0]}:${parts[1]}:${parts[2]}`;
}

function healthFromService(_svc: string): 'healthy' | 'degraded' | 'down' {
  return 'degraded';
}

export default function IncidentDetail() {
  const { id = '' } = useParams({ strict: false });
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [tab, setTab] = useState<(typeof TABS)[number]>('Overview');
  const [mttrClock, setMttrClock] = useState('00:00:00');
  const [appliedRecs, setAppliedRecs] = useState<Record<number, boolean>>({});
  const [assignee, setAssignee] = useState('');
  const [resolutionNotes, setResolutionNotes] = useState('');
  const realtimeEvents = useRealtimeStore((s) => s.events);

  const incidentQuery = useQuery({
    queryKey: ['incident', id],
    queryFn: () => fetchIncident(id),
    enabled: !!id,
    refetchInterval: (q) =>
      q.state.data?.status === 'OPEN' || q.state.data?.status === 'INVESTIGATING' ? 15_000 : false,
  });

  const timelineQuery = useQuery({
    queryKey: ['incident-timeline', id],
    queryFn: () => fetchIncidentTimeline(id),
    enabled: !!id && (tab === 'Timeline' || tab === 'Overview'),
  });

  const recsQuery = useQuery({
    queryKey: ['incident-recs', id],
    queryFn: () => fetchRecommendations(id),
    enabled: !!id && tab === 'Recommendations',
  });

  const liveRcaQuery = useQuery({
    queryKey: ['incident-rca-live', id],
    queryFn: () => fetchIncidentRCA(id),
    enabled: !!id && tab === 'AI Analysis',
  });

  const ackMutation = useMutation({
    mutationFn: () => acknowledgeIncident(id),
    onSuccess: () => {
      toast.success('Incident acknowledged');
      void queryClient.invalidateQueries({ queryKey: ['incident', id] });
    },
  });

  const jiraMut = useMutation({
    mutationFn: () =>
      triggerWorkflows({
        trigger: 'incident.p1',
        context: {
          title: incidentQuery.data?.title ?? 'Incident',
          description: incidentQuery.data?.summary ?? '',
        },
      }),
    onSuccess: (runs) => {
      const log = runs?.[0]?.stepsLog?.join('; ') ?? 'Workflow triggered';
      toast.success(log);
    },
    onError: () => toast.error('Failed to create Jira ticket — connect Jira in Integrations'),
  });

  const resolveMutation = useMutation({
    mutationFn: () =>
      resolveIncident(id, resolutionNotes.trim() || 'Resolved via NeuralOps UI'),
    onSuccess: () => {
      toast.success('Incident resolved');
      void queryClient.invalidateQueries({ queryKey: ['incident', id] });
    },
  });

  const incident = incidentQuery.data;

  useEffect(() => {
    if (!incident) return;
    setMttrClock(formatMttr(incident.startTime, incident.resolvedTime));
    if (incident.resolvedTime) return;
    const timer = setInterval(() => setMttrClock(formatMttr(incident.startTime)), 1000);
    return () => clearInterval(timer);
  }, [incident]);

  useEffect(() => {
    if (!incident || incident.status === 'RESOLVED') return;
    const relevant = realtimeEvents.find(
      (e) => e.type.startsWith('incident.') && e.message.includes(incident.id.slice(0, 8)),
    );
    if (relevant) void incidentQuery.refetch();
  }, [realtimeEvents, incident, incidentQuery]);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;
      if (e.key === 'Escape') navigate({ to: '/incidents' });
      if (e.key === 'a' || e.key === 'A') ackMutation.mutate();
      if (e.key === 'r' || e.key === 'R') resolveMutation.mutate();
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [navigate, ackMutation, resolveMutation]);

  const rca = incident?.rootCauseAnalysis;
  const liveRca = liveRcaQuery.data;
  const aiSections = useMemo(() => {
    const text = rca?.rootCauseDescription ?? incident?.summary ?? '';
    return {
      executive: text.split('.')[0] + '.',
      technical: text,
      contributing: (rca?.evidence ?? []).map((e) => e.description).slice(0, 3),
      propagation: incident?.blastRadius?.length
        ? `Failure propagated to ${incident.blastRadius.join(', ')} via synchronous dependency calls.`
        : 'Cascade followed standard service dependency chain.',
      detectionGap: rca?.deploymentCorrelation
        ? `Deployment ${rca.deploymentCorrelation.version} correlated at ${(rca.deploymentCorrelation.correlationScore * 100).toFixed(0)}% — alerting lagged metric thresholds.`
        : 'Anomaly detection did not fire until error rate exceeded static thresholds.',
    };
  }, [rca, incident]);

  if (incidentQuery.isLoading) return <LoadingState label="Loading incident…" />;
  if (incidentQuery.error) {
    return <ErrorState message={getApiErrorMessage(incidentQuery.error)} onRetry={() => incidentQuery.refetch()} />;
  }
  if (!incident) return null;

  const escalate = () => jiraMut.mutate();
  const share = () => {
    void navigator.clipboard.writeText(window.location.href);
    toast.success('Incident link copied');
  };

  return (
    <div className="incident-detail-page">
      <header className="incident-detail-header">
        <p className="muted incident-breadcrumb">
          <Link to="/incidents">Incidents</Link> › {incident.title}
        </p>
        <div className="incident-detail-header__row">
          <div>
            <h1>{incident.title}</h1>
            <div className="incident-detail-badges">
              <Badge variant={severityBadge(incident.severity)}>{incident.severity}</Badge>
              <Badge variant="info">{incident.status}</Badge>
              <span className="muted">
                Opened {formatDistanceToNow(new Date(incident.startTime), { addSuffix: true })}
                {!incident.acknowledgedAt && ' · Unacknowledged'}
              </span>
            </div>
          </div>
          <div className="incident-detail-actions">
            <Button variant="ghost" onClick={share}>
              <Share2 size={14} /> Share
            </Button>
            <Button variant="secondary" onClick={escalate} disabled={jiraMut.isPending}>
              <TrendingUp size={14} /> Jira
            </Button>
          </div>
        </div>
      </header>

      <nav className="incident-tabs">
        {TABS.map((t) => (
          <button
            key={t}
            type="button"
            className={`incident-tab ${tab === t ? 'incident-tab--active' : ''}`}
            onClick={() => setTab(t)}
          >
            {t}
          </button>
        ))}
      </nav>

      <div className="incident-detail-body">
        {tab === 'Overview' && (
          <div className="incident-overview-grid">
            <UnifiedContextPanel
              incidentId={incident.id}
              primaryService={incident.affectedServices?.[0]}
              affectedServices={incident.affectedServices}
            />

            <Card title="AI Root Cause Analysis">
              {rca ? (
                <>
                  <TypingText text={rca.rootCauseDescription} />
                  <p style={{ marginTop: 12 }}>
                    First failing service:{' '}
                    <strong className="error-text">{rca.firstFailingService}</strong>
                  </p>
                  {(incident.blastRadius ?? incident.affectedServices ?? []).length > 0 && (
                    <div style={{ marginTop: 12 }}>
                      <span className="muted">Blast radius</span>
                      <ul className="insight-list">
                        {(incident.blastRadius ?? incident.affectedServices).map((svc) => (
                          <li key={svc}>{svc}</li>
                        ))}
                      </ul>
                    </div>
                  )}
                  <div style={{ marginTop: 12 }}>
                    <span className="muted">Confidence {(rca.confidence * 100).toFixed(0)}%</span>
                    <div className="error-bar" style={{ marginTop: 4 }}>
                      <div className="error-bar__fill" style={{ width: `${rca.confidence * 100}%` }} />
                    </div>
                  </div>
                  {rca.deploymentCorrelation && (
                    <div className="deployment-correlation">
                      Deployment {rca.deploymentCorrelation.version} pushed on{' '}
                      {rca.deploymentCorrelation.service} —{' '}
                      {(rca.deploymentCorrelation.correlationScore * 100).toFixed(0)}% correlation with this incident
                    </div>
                  )}
                  <ul className="insight-list">
                    {(rca.evidence ?? []).map((e) => (
                      <li key={e.id}>{e.description}</li>
                    ))}
                  </ul>
                </>
              ) : (
                <p className="muted">{incident.summary}</p>
              )}
            </Card>

            <div className="incident-overview-sidebar">
              <Card title="Incident Stats">
                <div className="mttr-clock">
                  <span className="muted">MTTR</span>
                  <strong>{mttrClock}</strong>
                </div>
                <p>
                  Affected services: <strong>{incident.affectedServices?.length ?? 0}</strong>
                </p>
                <p>
                  Error signals: <strong>{(rca?.evidence ?? []).filter((e) => e.type === 'LOG').length || '—'}</strong>
                </p>
              </Card>

              <Card title="Service health">
                {(incident.affectedServices ?? []).map((svc) => (
                  <div key={svc} className="service-list-item">
                    <StatusDot status={healthFromService(svc)} label={svc} />
                  </div>
                ))}
              </Card>

              <Card title="On-call">
                <div className="oncall-panel">
                  <div>
                    <strong>Primary</strong>
                    <p className="muted">Alex Chen · SRE</p>
                    <Button variant="secondary" size="sm">
                      Page
                    </Button>
                  </div>
                  <div>
                    <strong>Secondary</strong>
                    <p className="muted">Priya Sharma · Platform</p>
                    <Button variant="ghost" size="sm">
                      Slack
                    </Button>
                  </div>
                </div>
              </Card>
            </div>
          </div>
        )}

        {tab === 'Timeline' && (
          <Card title="Event Timeline">
            {timelineQuery.isLoading && <LoadingState />}
            <TimelineView
              events={timelineQuery.data?.timeline ?? incident.timeline ?? []}
              narrative={timelineQuery.data?.narrative}
              onMarkRootCause={() => toast.success('Marked as root cause')}
            />
          </Card>
        )}

        {tab === 'Logs' && <IncidentLogsTab incident={incident} />}
        {tab === 'Traces' && <IncidentTracesTab incident={incident} />}
        {tab === 'Metrics' && <IncidentMetricsTab incident={incident} />}

        {tab === 'AI Analysis' && (
          <Card title="Full RCA">
            {liveRcaQuery.isLoading && <LoadingState label="Loading AI RCA…" />}
            {liveRcaQuery.isError && (
              <p className="muted">Live RCA unavailable — showing incident summary fallback.</p>
            )}
            {liveRca ? (
              <>
                <p className="muted">
                  Explanation {liveRca.explanationId} · confidence {(liveRca.confidence * 100).toFixed(0)}%
                </p>
                <section className="rca-section">
                  <h4>Summary</h4>
                  <p>{liveRca.summary}</p>
                </section>
                <section className="rca-section">
                  <h4>Root hypothesis</h4>
                  <p>
                    <strong>{liveRca.root.label}</strong> ({(liveRca.root.confidence * 100).toFixed(0)}%)
                  </p>
                  <ul className="insight-list">
                    {(liveRca.root.evidence ?? []).map((e, i) => (
                      <li key={i}>
                        {e.signal}: {e.detail}
                      </li>
                    ))}
                  </ul>
                </section>
              </>
            ) : (
              <>
                <section className="rca-section">
                  <h4>1. What happened</h4>
                  <p>{aiSections.executive}</p>
                </section>
                <section className="rca-section">
                  <h4>2. Technical root cause</h4>
                  <p>{aiSections.technical}</p>
                </section>
                <section className="rca-section">
                  <h4>3. Contributing factors</h4>
                  <ul className="insight-list">
                    {aiSections.contributing.map((c, i) => (
                      <li key={i}>{c}</li>
                    ))}
                  </ul>
                </section>
                <section className="rca-section">
                  <h4>4. Why it propagated</h4>
                  <p>{aiSections.propagation}</p>
                </section>
                <section className="rca-section">
                  <h4>5. Detection gap</h4>
                  <p>{aiSections.detectionGap}</p>
                </section>
              </>
            )}
          </Card>
        )}

        {tab === 'Recommendations' && (
          <div className="recommendations-grid">
            {recsQuery.isLoading && <LoadingState />}
            {(recsQuery.data ?? incident.recommendations ?? []).map((rec, idx) => (
              <Card key={idx} title={rec.description.slice(0, 60)}>
                <Badge variant={rec.priority === 'HIGH' ? 'critical' : rec.priority === 'MEDIUM' ? 'warning' : 'info'}>
                  {rec.priority}
                </Badge>
                <Badge variant="info">{rec.type}</Badge>
                <p style={{ marginTop: 12 }}>{rec.description}</p>
                {rec.codeSnippet && <CodeBlock code={rec.codeSnippet} language="config" />}
                <label className="checkbox-row" style={{ marginTop: 12 }}>
                  <input
                    type="checkbox"
                    checked={appliedRecs[idx] ?? false}
                    onChange={(e) => setAppliedRecs((s) => ({ ...s, [idx]: e.target.checked }))}
                  />
                  Mark as Applied
                </label>
              </Card>
            ))}
          </div>
        )}
      </div>

      <aside className="incident-sticky-actions" aria-label="Incident actions">
        <div className="incident-sticky-actions__fields">
          <Input
            label="Assignee"
            placeholder="owner@company.com"
            value={assignee}
            onChange={(e) => setAssignee(e.target.value)}
            hint={assignee ? `Assigned to ${assignee}` : 'Optional — stored locally until API supports assign'}
          />
          <Textarea
            label="Resolution notes"
            placeholder="Root cause, mitigation, follow-ups…"
            value={resolutionNotes}
            onChange={(e) => setResolutionNotes(e.target.value)}
            rows={2}
          />
        </div>
        <div className="incident-sticky-actions__buttons">
          <Button variant="secondary" onClick={() => ackMutation.mutate()} disabled={ackMutation.isPending}>
            Acknowledge (A)
          </Button>
          <Button
            variant="primary"
            onClick={() => {
              if (assignee) toast.success(`Assigned to ${assignee}`);
              resolveMutation.mutate();
            }}
            disabled={resolveMutation.isPending}
          >
            Resolve (R)
          </Button>
        </div>
      </aside>
    </div>
  );
}
