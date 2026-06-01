import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from '@tanstack/react-router';
import { fetchAttack } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { logsSearch } from '../utils/logsSearch';
import { Badge } from '../components/ui/Badge';
import { Card } from '../components/ui/Card';
import { ErrorState, LoadingState, PageHeader } from '../components/ui/PageStates';

export default function SecurityAttackDetail() {
  const { id = '' } = useParams({ strict: false });

  const { data, isLoading, error } = useQuery({
    queryKey: ['security-attack', id],
    queryFn: () => fetchAttack(id),
    enabled: id.length > 0,
  });

  return (
    <div>
      <PageHeader
        title={data?.type ?? 'Attack detail'}
        subtitle="Runtime attack event"
        actions={<Link to="/security">← Security</Link>}
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={getApiErrorMessage(error)} />}
      {data && (
        <Card title="Event">
          <div style={{ display: 'flex', gap: 12, marginBottom: 12 }}>
            <Badge variant={data.blocked ? 'healthy' : 'critical'}>{data.blocked ? 'Blocked' : 'Open'}</Badge>
            <span>{data.type}</span>
          </div>
          <p><strong>Source IP:</strong> {data.sourceIp}</p>
          <p><strong>Target service:</strong> {data.service}</p>
          <p className="muted">Detected {new Date(data.detectedAt).toLocaleString()}</p>
          <div style={{ display: 'flex', gap: 16, marginTop: 16 }}>
            <Link to="/traces" search={{ service: data.service }}>View traces</Link>
            <Link to="/logs" search={logsSearch({ service: data.service })}>View logs</Link>
            <Link to="/entities/$type/$id" params={{ type: 'service', id: data.service }}>Entity page</Link>
          </div>
        </Card>
      )}
    </div>
  );
}
