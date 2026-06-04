import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { formatDistanceToNow } from 'date-fns';
import {
  Line,
  LineChart,
  ResponsiveContainer,
  ReferenceArea,
  XAxis,
  YAxis,
  Tooltip,
} from 'recharts';
import { fetchDashboardOverview } from '../api/dashboard';
import { fetchEntityAnomalies } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { useRealtimeStore } from '../store/realtimeStore';
import { Badge } from '../components/ui/Badge';
import { Card } from '../components/ui/Card';
import { StatusDot } from '../components/ui/StatusDot';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { DataExportMenu } from '../components/ui/DataExportMenu';
import { StitchPageShell } from '../components/stitch';

const ENVIRONMENTS = ['PROD', 'STAGING', 'DEV'] as const;

function Gauge({ score, label }: { score: number; label: string }) {
  const color = score > 60 ? 'var(--error)' : score > 30 ? 'var(--warning)' : 'var(--success)';
  const status = score > 60 ? 'Critical' : score > 30 ? 'Elevated' : 'Normal';
  const circumference = 2 * Math.PI * 45;
  const offset = circumference - (score / 100) * circumference;

  return (
    <div style={{ textAlign: 'center' }}>
      <svg className="gauge" viewBox="0 0 120 120">
        <circle cx="60" cy="60" r="45" fill="none" stroke="var(--bg-highlight)" strokeWidth="10" />
        <circle
          cx="60"
          cy="60"
          r="45"
          fill="none"
          stroke={color}
          strokeWidth="10"
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          strokeLinecap="round"
          transform="rotate(-90 60 60)"
          style={{ transition: 'stroke-dashoffset 0.8s ease' }}
        />
        <text x="60" y="58" textAnchor="middle" fill="var(--text-primary)" fontSize="22" fontWeight="700">
          {score}
        </text>
        <text x="60" y="76" textAnchor="middle" fill="var(--text-muted)" fontSize="10">
          {status}
        </text>
      </svg>
      <p className="gauge__label">{label}</p>
    </div>
  );
}

export default function AnomalyDetection() {
  const events = useRealtimeStore((s) => s.events);

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['anomalies-overview'],
    queryFn: fetchDashboardOverview,
    refetchInterval: 30_000,
  });

  const entityAnomaliesQuery = useQuery({
    queryKey: ['entity-anomalies'],
    queryFn: fetchEntityAnomalies,
    refetchInterval: 30_000,
  });

  const anomalies = data?.recentAnomalies ?? [];
  const entityAnomalies = entityAnomaliesQuery.data ?? [];
  const activeCount = Math.max(anomalies.length, entityAnomalies.length);

  const chartData = useMemo(
    () =>
      Array.from({ length: 24 }, (_, i) => ({
        hour: `${i}:00`,
        baseline: 45 + Math.sin(i / 3) * 10,
        actual: 45 + Math.sin(i / 3) * 10 + (i > 18 ? 80 : 0),
      })),
    [],
  );

  const historyRows = useMemo(
    () =>
      anomalies.map((a, idx) => ({
        id: idx,
        time: a.timestamp,
        service: a.service,
        metric: 'LATENCY',
        score: 65 + idx * 5,
        deviation: '+411%',
        duration: '12m',
        outcome: idx % 2 === 0 ? 'True Positive' : 'Investigating',
      })),
    [anomalies],
  );

  return (
    <StitchPageShell
      title="Anomaly Detection"
      subtitle={`${activeCount} active anomalies · Neural score monitoring`}
      actions={
        <DataExportMenu
          getData={() => [
            ...entityAnomalies.map((a) => ({ type: 'entity', ...a })),
            ...historyRows,
          ]}
          filenamePrefix="anomalies"
          disabled={!entityAnomalies.length && !historyRows.length}
        />
      }
    >
      {isLoading && <LoadingState />}
      {error && <ErrorState message={getApiErrorMessage(error)} onRetry={() => refetch()} />}

      <div className="anomaly-layout">
        <div>
          <div className="anomaly-gauges">
            {ENVIRONMENTS.map((env, idx) => (
              <Card key={env} hover>
                <Gauge score={20 + idx * 25 + activeCount * 5} label={`${env} Neural Score`} />
              </Card>
            ))}
          </div>

          <div style={{ marginTop: 24, display: 'grid', gap: 16 }}>
            {entityAnomalies.map((a) => (
              <Card key={a.id} hover>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                  <div>
                    <StatusDot status="degraded" pulse label={a.service} />
                    <p style={{ margin: '8px 0', fontSize: 13 }}>{a.message}</p>
                    <p className="muted" style={{ fontSize: 12 }}>
                      {a.metric} · score {a.score.toFixed(1)}
                    </p>
                  </div>
                  <div style={{ textAlign: 'right' }}>
                    <Badge variant="warning">ENTITY</Badge>
                    <p className="muted" style={{ fontSize: 11, marginTop: 8 }}>
                      {formatDistanceToNow(new Date(a.detectedAt), { addSuffix: true })}
                    </p>
                  </div>
                </div>
              </Card>
            ))}
            {anomalies.map((a) => (
              <Card key={`${a.service}-${a.timestamp}`} hover>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                  <div>
                    <StatusDot status="degraded" pulse label={a.service} />
                    <p style={{ margin: '8px 0', fontSize: 13 }}>{a.message}</p>
                    <p className="muted" style={{ fontSize: 12 }}>
                      Baseline: 45ms | Actual: 230ms | +411%
                    </p>
                  </div>
                  <div style={{ textAlign: 'right' }}>
                    <Badge variant="warning">NEW</Badge>
                    <p className="muted" style={{ fontSize: 11, marginTop: 8 }}>
                      {formatDistanceToNow(new Date(a.timestamp), { addSuffix: true })}
                    </p>
                  </div>
                </div>
              </Card>
            ))}
            {anomalies.length === 0 && !isLoading && (
              <Card><p className="muted">No active anomalies — all metrics within baseline</p></Card>
            )}
          </div>

          <Card title="Anomaly Drill-Down" hover={false} style={{ marginTop: 24 }}>
            <div style={{ height: 280 }}>
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={chartData}>
                  <XAxis dataKey="hour" stroke="var(--text-muted)" fontSize={10} />
                  <YAxis stroke="var(--text-muted)" fontSize={10} />
                  <Tooltip contentStyle={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-subtle)' }} />
                  <ReferenceArea x1="18:00" x2="23:00" fill="rgba(239,68,68,0.1)" />
                  <Line type="monotone" dataKey="baseline" stroke="var(--text-muted)" strokeDasharray="4 4" dot={false} />
                  <Line type="monotone" dataKey="actual" stroke="var(--error)" strokeWidth={2} dot={false} />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </Card>

          <Card title="Anomaly History" hover={false} style={{ marginTop: 24 }}>
            <table className="incident-table">
              <thead>
                <tr>
                  <th>Time</th>
                  <th>Service</th>
                  <th>Metric</th>
                  <th>Score</th>
                  <th>Deviation</th>
                  <th>Duration</th>
                  <th>Outcome</th>
                </tr>
              </thead>
              <tbody>
                {historyRows.map((row) => (
                  <tr key={row.id}>
                    <td>{formatDistanceToNow(new Date(row.time), { addSuffix: true })}</td>
                    <td>{row.service}</td>
                    <td>{row.metric}</td>
                    <td>{row.score}</td>
                    <td>{row.deviation}</td>
                    <td>{row.duration}</td>
                    <td>{row.outcome}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
        </div>

        <aside className="anomaly-feed">
          <h3 style={{ marginTop: 0, fontFamily: 'var(--font-display)' }}>Live Feed</h3>
          {events.filter((e) => e.type !== 'heartbeat').slice(0, 20).map((ev, i) => (
            <div key={`${ev.timestamp}-${i}`} className="service-list-item">
              <Badge variant="warning">{ev.type}</Badge>
              <span style={{ fontSize: 12 }}>{ev.message}</span>
            </div>
          ))}
          {events.length === 0 && <p className="muted">Waiting for realtime events…</p>}
        </aside>
      </div>
    </StitchPageShell>
  );
}
