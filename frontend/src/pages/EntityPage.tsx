import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useParams, useSearch } from '@tanstack/react-router';
import { Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { fetchTopology, queryMetric, searchTraces } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { logsSearch } from '../utils/logsSearch';
import { EntityShell, type EntityTab, type EntityType } from '../components/entity/EntityShell';
import { Card } from '../components/ui/Card';
import { DomainEmptyState } from '../components/ui/DomainEmptyState';
import { ErrorState, LoadingState, PageHeader } from '../components/ui/PageStates';
import { MetricCard } from '../components/ui/MetricCard';
import { chartCartesianDefaults } from '../lib/chartTheme';

const VALID_TYPES: EntityType[] = ['service', 'host', 'pod', 'database', 'container'];

export default function EntityPage() {
  const { type = 'service', id = '' } = useParams({ strict: false });
  const search = useSearch({ strict: false }) as { tab?: EntityTab };
  const [tab, setTab] = useState<EntityTab>(search.tab ?? 'overview');
  const entityType = (VALID_TYPES.includes(type as EntityType) ? type : 'service') as EntityType;

  const topologyQuery = useQuery({
    queryKey: ['topology'],
    queryFn: () => fetchTopology(),
    enabled: entityType === 'service',
  });

  const latencyQuery = useQuery({
    queryKey: ['entity-metric', id],
    queryFn: () => queryMetric('latency_p99', id),
    enabled: entityType === 'service' && id.length > 0 && (tab === 'metrics' || tab === 'overview'),
  });

  const tracesQuery = useQuery({
    queryKey: ['entity-traces', id],
    queryFn: () => searchTraces({ service: id, limit: 25 }),
    enabled: entityType === 'service' && id.length > 0 && tab === 'traces',
  });

  const node = topologyQuery.data?.nodes.find((n) => n.id === id);
  const health = node?.health ?? 'healthy';
  const lastPoint = latencyQuery.data?.points[latencyQuery.data.points.length - 1];
  const chartDefaults = chartCartesianDefaults();
  const chartData = (latencyQuery.data?.points ?? []).map((p) => ({
    time: new Date(p.timestamp).toLocaleTimeString(),
    value: p.value,
  }));

  if (!id) {
    return (
      <div>
        <PageHeader title="Entity" subtitle="Select an entity from topology or service map" />
        <Link to="/service-map">Open service map →</Link>
      </div>
    );
  }

  return (
    <EntityShell
      entityType={entityType}
      entityId={id}
      displayName={node?.displayName ?? id}
      health={health}
      activeTab={tab}
      onTabChange={setTab}
    >
      {topologyQuery.isLoading && tab === 'overview' && <LoadingState />}
      {topologyQuery.error && <ErrorState message={getApiErrorMessage(topologyQuery.error)} onRetry={() => topologyQuery.refetch()} />}

      {tab === 'overview' && entityType === 'service' && (
        <>
          <div className="dashboard-row-3">
            <MetricCard label="Error rate" value={`${(node?.errorRate ?? 0).toFixed(2)}%`} />
            <MetricCard label="Throughput" value={`${Math.round(node?.throughputRpm ?? 0)} rpm`} />
            <MetricCard
              label="P99 latency"
              value={lastPoint?.value != null ? `${lastPoint.value.toFixed(0)}ms` : '—'}
            />
          </div>
          <Card title="Quick links" style={{ marginTop: 24 }}>
            <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap' }}>
              <button type="button" className="pill" onClick={() => setTab('traces')}>Traces</button>
              <button type="button" className="pill" onClick={() => setTab('logs')}>Logs</button>
              <button type="button" className="pill" onClick={() => setTab('metrics')}>Metrics</button>
              <Link to="/service-map">Service map</Link>
              <Link to="/incidents">Incidents</Link>
            </div>
          </Card>
        </>
      )}

      {tab === 'metrics' && (
        <>
          {latencyQuery.isLoading && <LoadingState />}
          {latencyQuery.error && <ErrorState message={getApiErrorMessage(latencyQuery.error)} onRetry={() => latencyQuery.refetch()} />}
          {latencyQuery.data && (
            <Card title="P99 latency">
              <ResponsiveContainer width="100%" height={260}>
                <LineChart data={chartData}>
                  <XAxis dataKey="time" {...chartDefaults.axis} />
                  <YAxis {...chartDefaults.axis} />
                  <Tooltip contentStyle={chartDefaults.tooltipStyle} />
                  <Line type="monotone" dataKey="value" stroke="var(--accent-primary)" dot={false} />
                </LineChart>
              </ResponsiveContainer>
            </Card>
          )}
        </>
      )}

      {tab === 'logs' && (
        <Card title="Logs">
          <p className="muted">Open Log Explorer filtered to this entity.</p>
          <Link to="/logs" search={logsSearch({ service: id })} className="pill">
            View logs for {id} →
          </Link>
        </Card>
      )}

      {tab === 'traces' && (
        <>
          {tracesQuery.isLoading && <LoadingState />}
          {tracesQuery.error && <ErrorState message={getApiErrorMessage(tracesQuery.error)} onRetry={() => tracesQuery.refetch()} />}
          {(tracesQuery.data?.length ?? 0) === 0 && tracesQuery.isFetched && <DomainEmptyState domain="traces" />}
          {tracesQuery.data && tracesQuery.data.length > 0 && (
            <Card title="Recent traces">
              {tracesQuery.data.map((t) => (
                <Link key={t.traceId} to="/traces/$traceId" params={{ traceId: t.traceId }} className="log-row">
                  <span>{t.operation}</span>
                  <span className="muted">{t.durationMs}ms · {t.status}</span>
                </Link>
              ))}
            </Card>
          )}
        </>
      )}
    </EntityShell>
  );
}
