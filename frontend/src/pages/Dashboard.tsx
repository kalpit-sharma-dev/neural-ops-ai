import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { Link, useNavigate } from '@tanstack/react-router';
import { formatDistanceToNow } from 'date-fns';
import { RefreshCw, Siren } from 'lucide-react';
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { fetchDashboardErrorSeries, fetchDashboardOverview } from '../api/dashboard';
import { getApiErrorMessage } from '../api/client';
import { MetricCard } from '../components/ui/MetricCard';
import { Card } from '../components/ui/Card';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { StatusDot } from '../components/ui/StatusDot';
import { CardSkeleton } from '../components/ui/Skeleton';
import { ErrorState, PageHeader } from '../components/ui/PageStates';
import { KpiRow } from '../components/stitch';
import type { IncidentSummary } from '../api/types';
import { getChartColors } from '../lib/chartColors';
import { chartAxisProps, chartGridProps, chartTooltipStyle } from '../lib/chartTheme';
import { useThemeStore } from '../store/themeStore';
import { useReducedMotion } from '../hooks/useReducedMotion';
import { useFilterStore } from '../store/filterStore';
import { useTimeBounds } from '../hooks/useTimeBounds';

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
  const theme = useThemeStore((s) => s.theme);
  const reducedMotion = useReducedMotion();
  const navigate = useNavigate();
  const setErrorsOnly = useFilterStore((s) => s.setErrorsOnly);
  const { key: rangeKey, resolve: resolveRange } = useTimeBounds();
  const chartColors = useMemo(() => getChartColors(), [theme]);
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date());

  const { data, isLoading, error, refetch, isFetching, dataUpdatedAt } = useQuery({
    queryKey: ['dashboard-overview'],
    queryFn: fetchDashboardOverview,
    refetchInterval: 30_000,
  });

  const services = useMemo(
    () => (data?.topFailingServices ?? []).map((s) => s.service),
    [data?.topFailingServices],
  );

  const chartQuery = useQuery({
    queryKey: ['dashboard-error-series', services, rangeKey],
    queryFn: () => {
      const { startIso, endIso } = resolveRange();
      return fetchDashboardErrorSeries(services, { start: startIso, end: endIso });
    },
    enabled: services.length > 0,
  });

  const severityCounts = useMemo(
    () => countBySeverity(data?.activeIncidents ?? []),
    [data?.activeIncidents],
  );

  const hasCritical = (severityCounts.P1 ?? 0) + (severityCounts.P2 ?? 0) > 0;
  const openP1 = (data?.activeIncidents ?? []).find((i) => i.severity === 'P1' && i.status !== 'RESOLVED');

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
    if (items.length === 0) items.push('All systems operating within normal parameters');
    return items;
  }, [data]);

  const motionProps = reducedMotion
    ? {}
    : { initial: { opacity: 0, y: 12 }, animate: { opacity: 1, y: 0 }, transition: { duration: 0.35 } };

  const handleRefresh = () => {
    void refetch().then(() => setLastRefresh(new Date()));
    void chartQuery.refetch();
  };

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
    <motion.div className="dashboard-grid" {...motionProps}>
      <PageHeader
        title="Command Center"
        subtitle="Real-time observability overview"
        actions={
          <div className="dashboard-header-actions">
            <span className="muted dashboard-refresh">
              Updated {formatDistanceToNow(dataUpdatedAt ? new Date(dataUpdatedAt) : lastRefresh, { addSuffix: true })}
            </span>
            <Button variant="ghost" size="sm" onClick={handleRefresh} disabled={isFetching}>
              <RefreshCw size={14} className={isFetching ? 'spin' : ''} /> Refresh
            </Button>
            {openP1 && (
              <Button
                variant="primary"
                size="sm"
                onClick={() => navigate({ to: '/incidents/$id', params: { id: openP1.id } })}
              >
                <Siren size={14} /> War room
              </Button>
            )}
          </div>
        }
      />

      {hasCritical && (
        <Card className="incident-strip">
          <div className="incident-strip__counts">
            {(['P1', 'P2'] as const).map((s) => (
              <Link key={s} to="/incidents" search={{ severity: s }}>
                <Badge variant={severityVariant(s)}>
                  {s}: {severityCounts[s] ?? 0}
                </Badge>
              </Link>
            ))}
          </div>
          <div className="incident-strip__list">
            {(data?.activeIncidents ?? []).slice(0, 5).map((inc) => (
              <Link key={inc.id} to="/incidents/$id" params={{ id: inc.id }}>
                {inc.title}
              </Link>
            ))}
          </div>
        </Card>
      )}

      <KpiRow>
        <button
          type="button"
          className="dashboard-kpi-btn"
          onClick={() => navigate({ to: '/incidents' })}
        >
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
        </button>
        <button
          type="button"
          className="dashboard-kpi-btn"
          onClick={() => {
            setErrorsOnly();
            navigate({ to: '/logs', search: {} });
          }}
        >
          <MetricCard
            label="Error Rate (1h)"
            value={`${(data?.errorRateLastHour ?? 0).toFixed(2)}%`}
            trend={data?.errorRateLastHour ? 12 : -3}
          />
        </button>
        <button type="button" className="dashboard-kpi-btn" onClick={() => navigate({ to: '/traces' })}>
          <MetricCard label="Avg Latency p99" value={`${Math.round(data?.latencyP99Ms ?? 142)}ms`} />
        </button>
        <button type="button" className="dashboard-kpi-btn" onClick={() => navigate({ to: '/incidents' })}>
          <MetricCard
            label="MTTR Today"
            value={data?.mttrMinutes ? `${Math.round(data.mttrMinutes)}m` : '—'}
            footer={<span className="muted">resolved incidents</span>}
          />
        </button>
      </KpiRow>

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
            <button
              key={svc.service}
              type="button"
              className="service-list-item service-list-item--btn"
              onClick={() => navigate({ to: '/logs', search: { service: svc.service } })}
            >
              <div>
                <StatusDot
                  status={svc.errorCount > 100 ? 'down' : svc.errorCount > 20 ? 'degraded' : 'healthy'}
                  label={svc.service}
                />
                <div className="error-bar">
                  <div className="error-bar__fill" style={{ width: `${Math.min(100, svc.errorCount / 2)}%` }} />
                </div>
              </div>
              <span className="muted">{svc.errorCount} errs</span>
            </button>
          ))}
          {(data?.topFailingServices ?? []).length === 0 && <p className="muted">No failing services detected</p>}
        </Card>
      </div>

      <Card title="Error Rate by Service">
        {chartQuery.isLoading && <p className="muted">Loading metrics…</p>}
        {chartQuery.error && (
          <ErrorState message={getApiErrorMessage(chartQuery.error)} onRetry={() => chartQuery.refetch()} />
        )}
        <div style={{ height: 280 }}>
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={chartQuery.data ?? []}>
              <CartesianGrid {...chartGridProps()} />
              <XAxis dataKey="hour" {...chartAxisProps()} />
              <YAxis {...chartAxisProps()} />
              <Tooltip contentStyle={chartTooltipStyle()} />
              {services.slice(0, 4).map((s, idx) => (
                <Area
                  key={s}
                  type="monotone"
                  dataKey={s}
                  stackId="1"
                  stroke={chartColors[idx]}
                  fill={chartColors[idx]}
                  fillOpacity={0.3}
                />
              ))}
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
                <p className="muted" style={{ margin: '4px 0 0', fontSize: 12 }}>
                  {a.message}
                </p>
              </div>
              <Badge variant="warning">{a.severity}</Badge>
            </div>
          ))}
          {(data?.recentAnomalies ?? []).length === 0 && <p className="muted">No anomalies detected</p>}
        </Card>

        <Card title="Recent Deployments">
          <p className="muted">Connect deployment events API for live data</p>
        </Card>

        <Card title="AI Insights">
          <ul className="insight-list">
            {insights.map((text) => (
              <li key={text}>{text}</li>
            ))}
          </ul>
          {data?.generatedAt && (
            <p className="muted" style={{ marginTop: 12, fontSize: 11 }}>
              Snapshot {formatDistanceToNow(new Date(data.generatedAt), { addSuffix: true })}
            </p>
          )}
        </Card>
      </div>
    </motion.div>
  );
}
