import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { formatDistanceToNow } from 'date-fns';
import { fetchIncidents } from '../api/incidents';
import { getApiErrorMessage } from '../api/client';
import { Badge } from '../components/ui/Badge';
import { ErrorState, LoadingState, PageHeader } from '../components/ui/PageStates';

function severityBadge(sev: string) {
  const map: Record<string, 'p1' | 'p2' | 'p3' | 'p4'> = {
    P1: 'p1',
    P2: 'p2',
    P3: 'p3',
    P4: 'p4',
  };
  return map[sev] ?? 'p4';
}

export default function Incidents() {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['incidents'],
    queryFn: () => fetchIncidents({ size: 100 }),
    refetchInterval: 30_000,
  });

  return (
    <div>
      <PageHeader title="Incidents" subtitle="Active and recent operational incidents" />

      {isLoading && <LoadingState />}
      {error && <ErrorState message={getApiErrorMessage(error)} onRetry={() => refetch()} />}

      {data && (
        <div className="ui-card">
          <table className="incident-table">
            <thead>
              <tr>
                <th>Severity</th>
                <th>Title</th>
                <th>Status</th>
                <th>Services</th>
                <th>Started</th>
              </tr>
            </thead>
            <tbody>
              {data.map((inc) => (
                <tr key={inc.id}>
                  <td>
                    <Badge variant={severityBadge(inc.severity)}>{inc.severity}</Badge>
                  </td>
                  <td>
                    <Link to="/incidents/$id" params={{ id: inc.id }}>{inc.title}</Link>
                  </td>
                  <td>{inc.status}</td>
                  <td>{inc.affectedServices?.join(', ') ?? '—'}</td>
                  <td className="muted">
                    {formatDistanceToNow(new Date(inc.startTime), { addSuffix: true })}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {data.length === 0 && <p className="muted" style={{ padding: 16 }}>No incidents found</p>}
        </div>
      )}
    </div>
  );
}
