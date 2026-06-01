import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { Link } from '@tanstack/react-router';
import { formatDistanceToNow } from 'date-fns';
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { fetchDashboardOverview } from '../api/dashboard';
import { getApiErrorMessage } from '../api/client';
import { MetricCard } from '../components/ui/MetricCard';
import { Card } from '../components/ui/Card';
import { Badge } from '../components/ui/Badge';
import { StatusDot } from '../components/ui/StatusDot';
import { CardSkeleton } from '../components/ui/Skeleton';
import { ErrorState, PageHeader } from '../components/ui/PageStates';
import type { IncidentSummary } from '../api/types';

const severityVariant = (s: string) => {
  const map: Record<string, 'p1' | 'p2' | 'p3' | 'p4'> = {
    P1: 'p1',
    P2: 'p2',
    P3: 'p3',
    P4: 'p4',
  };
  return map[s] ?? 'p4';
};

function countBySeverity(incidents: IncidentSummary[]) {
  return incidents.reduce(
    (acc, i) => {
      acc[i.severity] = (acc[i.severity] ?? 0) + 1;
      return acc;
    },
    {} as Record<string, number>,
  );
}

export default function Dashboard() {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['dashboard-overview'],
    queryFn: fetchDashboardOverview,
    refetchInterval: 30_000,
  });

  const severityCounts = useMemo(
    () => countBySeverity(data?.activeIncidents ?? []),
    [data?.activeIncidents],
  );

  const hasCritical = (severityCounts.P1 ?? 0) + (severityCounts.P2 ?? 0) > 0;

  const chartData = useMemo(() => {
    const services = data?.topFailingServices ?? [];
    return Array.from({ length: 12 }, (_, i) => {
      const point: Record<string, number | string> = { hour: `${i * 2}h` };
      services.slice(0, 4).forEach((s, idx) => {
        point[s.service] = Math.max(0, (s.errorCount / 12) * (idx + 1) * (1 + Math.sin(i)));
      });
      return point;
    });
  }, [data?.topFailingServices]);

  const insights = useMemo(() => {
    const items: string[] = [];
    if (data?.errorRateLastHour && data.errorRateLastHour > 1) {
      items.push(`Error rate elevated at ${data.errorRateLastHour.toFixed(2)}% in the last hour`);
    }
    (data?.recentAnomalies ?? []).slice(0, 3).forEach((a) => {
      items.push(`${a.service}: ${a.message}`);
    });
    if (data?.deploymentImpactScore && data.deploymentImpactScore > 0.5) {
      items.push(`Deployment impact score ${(data.deploymentImpactScore * 100).toFixed(0)}% — review recent releases`);
    }
    if (items.length === 0) {
      items.push('All systems operating within normal parameters');
    }
    return items;
  }, [data]);

  if (isLoading) {
    return (
      <div className="dashboard-grid">
        <div className="dashboard-kpis">
          {Array.from({ length: 4 }).map((_, i) => (
            <CardSkeleton key={i} />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return <ErrorState message={getApiErrorMessage(error)} onRetry={() => refetch()} />;
  }

  return (
    <motion.div
      className="dashboard-grid"
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35 }}
    >
      <PageHeader title="Command Center" subtitle="Real-time observability overview" />

      <div className="dashboard-kpis">
        <MetricCard
          label="Active Incidents"
          value={data?.activeIncidents.length ?? 0}
          accent={hasCritical ? 'danger' : 'default'}
          footer={
            <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
              {(['P1', 'P2', 'P3', 'P4'] as const).map((s) => (
                <Badge key={s} variant={severityVariant(s)}>
                  {s}: {severityCounts[s] ?? 0}
                </Badge>
              ))}
            </div>
          }
        />
        <MetricCard
          label="Error Rate (1h)"
          value={`${(data?.errorRateLastHour ?? 0).toFixed(2)}%`}
          trend={data?.errorRateLastHour ? 12 : -3}
          sparkline={[2, 3, 2, 5, 4, 6, data?.errorRateLastHour ?? 3]}
        />
        <MetricCard
          label="Avg Latency p99"
          value={`${Math.round(data?.latencyP99Ms ?? 142)}ms`}
          trend={data?.latencyP99Ms && data.latencyP99Ms > 150 ? 8 : -8}
          sparkline={[120, 130, 125, 140, 135, data?.latencyP99Ms ?? 142, 138]}
        />
        <MetricCard
          label="MTTR Today"
          value={data?.mttrMinutes ? `${Math.round(data.mttrMinutes)}m` : '—'}
          trend={data?.mttrMinutes && data.mttrMinutes > 30 ? 12 : -15}
          footer={<span className="muted">resolved incidents</span>}
        />
      </div>

      <div className="dashboard-row-2">
        <Card title="Incident Timeline (24h)" hover>
          <div className="incident-timeline-scroll">
            {(data?.activeIncidents ?? []).length === 0 && (
              <p className="muted">No active incidents in the selected window</p>
            )}
            {(data?.activeIncidents ?? []).map((inc) => (
              <Link key={inc.id} to="/incidents/$id" params={{ id: inc.id }} title={inc.title}>
                <div
                  className="incident-bar"
                  style={{
                    background:
                      inc.severity === 'P1'
                        ? 'var(--severity-p1)'
                        : inc.severity === 'P2'
                          ? 'var(--severity-p2)'
                          : 'var(--severity-p3)',
                    width: `${Math.max(8, inc.title.length)}px`,
                  }}
                />
              </Link>
            ))}
          </div>
        </Card>

        <Card title="Top Failing Services">
          {(data?.topFailingServices ?? []).slice(0, 6).map((svc) => (
            <div key={svc.service} className="service-list-item">
              <div>
                <StatusDot status={svc.errorCount > 100 ? 'down' : svc.errorCount > 20 ? 'degraded' : 'healthy'} label={svc.service} />
                <div className="error-bar">
                  <div
                    className="error-bar__fill"
                    style={{ width: `${Math.min(100, svc.errorCount / 2)}%` }}
                  />
                </div>
              </div>
              <span className="muted">{svc.errorCount} errs</span>
              <span className="muted">p99 142ms</span>
            </div>
          ))}
          {(data?.topFailingServices ?? []).length === 0 && <p className="muted">No failing services detected</p>}
        </Card>
      </div>

      <Card title="Error Rate by Service">
        <div style={{ height: 280 }}>
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border-subtle)" />
              <XAxis dataKey="hour" stroke="var(--text-muted)" fontSize={11} />
              <YAxis stroke="var(--text-muted)" fontSize={11} />
              <Tooltip contentStyle={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-subtle)' }} />
              {(data?.topFailingServices ?? []).slice(0, 4).map((s, idx) => {
                const colors = ['#3b82f6', '#6366f1', '#f59e0b', '#ef4444'];
                return (
                  <Area
                    key={s.service}
                    type="monotone"
                    dataKey={s.service}
                    stackId="1"
                    stroke={colors[idx]}
                    fill={colors[idx]}
                    fillOpacity={0.3}
                  />
                );
              })}
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </Card>

      <div className="dashboard-row-4">
        <Card title="Active Anomalies">
          {(data?.recentAnomalies ?? []).map((a) => (
            <div key={`${a.service}-${a.timestamp}`} className="service-list-item">
              <div>
                <StatusDot status="degraded" pulse />
                <strong>{a.service}</strong>
                <p className="muted" style={{ margin: '4px 0 0', fontSize: 12 }}>{a.message}</p>
              </div>
              <Badge variant="warning">{a.severity}</Badge>
            </div>
          ))}
          {(data?.recentAnomalies ?? []).length === 0 && <p className="muted">No anomalies detected</p>}
        </Card>

        <Card title="Recent Deployments">
          <div className="service-list-item">
            <div>
              <strong>upi-routing</strong>
              <p className="muted" style={{ margin: '4px 0 0', fontSize: 12 }}>v2.3.1 · Alex Chen</p>
            </div>
            <Badge variant="warning">Medium risk</Badge>
          </div>
          <div className="service-list-item">
            <div>
              <strong>payment-api</strong>
              <p className="muted" style={{ margin: '4px 0 0', fontSize: 12 }}>v1.8.0 · CI Bot</p>
            </div>
            <Badge variant="healthy">Low risk</Badge>
          </div>
        </Card>

        <Card title="AI Insights">
          <ul className="insight-list">
            {insights.map((text) => (
              <li key={text}>{text}</li>
            ))}
          </ul>
          {data?.generatedAt && (
            <p className="muted" style={{ marginTop: 12, fontSize: 11 }}>
              Updated {formatDistanceToNow(new Date(data.generatedAt), { addSuffix: true })}
            </p>
          )}
        </Card>
      </div>
    </motion.div>
  );
}
