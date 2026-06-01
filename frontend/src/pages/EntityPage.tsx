import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from '@tanstack/react-router';
import { fetchTopology, queryMetric } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { logsSearch } from '../utils/logsSearch';
import { EntityShell, type EntityType } from '../components/entity/EntityShell';
import { Card } from '../components/ui/Card';
import { ErrorState, LoadingState, PageHeader } from '../components/ui/PageStates';
import { MetricCard } from '../components/ui/MetricCard';

const VALID_TYPES: EntityType[] = ['service', 'host', 'pod', 'database', 'container'];

export default function EntityPage() {
  const { type = 'service', id = '' } = useParams({ strict: false });
  const entityType = (VALID_TYPES.includes(type as EntityType) ? type : 'service') as EntityType;

  const topologyQuery = useQuery({
    queryKey: ['topology'],
    queryFn: () => fetchTopology(),
    enabled: entityType === 'service',
  });

  const latencyQuery = useQuery({
    queryKey: ['entity-metric', id],
    queryFn: () => queryMetric('latency_p99', id),
    enabled: entityType === 'service' && id.length > 0,
  });

  const node = topologyQuery.data?.nodes.find((n) => n.id === id);
  const health = node?.health ?? 'healthy';
  const lastPoint = latencyQuery.data?.points[latencyQuery.data.points.length - 1];

  if (!id) {
    return (
      <div>
        <PageHeader title="Entity" subtitle="Select an entity from topology or service map" />
        <Link to="/service-map">Open service map →</Link>
      </div>
    );
  }

  return (
    <EntityShell entityType={entityType} entityId={id} displayName={node?.displayName ?? id} health={health}>
      {topologyQuery.isLoading && <LoadingState />}
      {topologyQuery.error && <ErrorState message={getApiErrorMessage(topologyQuery.error)} />}

      {entityType === 'service' && (
        <div className="dashboard-row-3">
          <MetricCard label="Error rate" value={`${(node?.errorRate ?? 0).toFixed(2)}%`} />
          <MetricCard label="Throughput" value={`${Math.round(node?.throughputRpm ?? 0)} rpm`} />
          <MetricCard
            label="P99 latency"
            value={lastPoint?.value != null ? `${lastPoint.value.toFixed(0)}ms` : '—'}
          />
        </div>
      )}

      <Card title="Related" style={{ marginTop: 24 }}>
        <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap' }}>
          <Link to="/traces" search={{ service: id }}>View traces</Link>
          <Link to="/logs" search={logsSearch({ service: id })}>View logs</Link>
          <Link to="/metrics" search={{ service: id }}>View metrics</Link>
          <Link to="/incidents">Incidents</Link>
        </div>
      </Card>
    </EntityShell>
  );
}
