import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { fetchRUMSessions } from '../api/observability';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function RUM() {
  const { data, isLoading } = useQuery({ queryKey: ['rum-sessions'], queryFn: fetchRUMSessions });

  return (
    <div>
      <PageHeader
        title="Real User Monitoring"
        subtitle="Browser sessions, Core Web Vitals, session replay"
        actions={
          <a href="/rum/neuralops-rum.js" download="neuralops-rum.js" className="muted">
            Download RUM SDK
          </a>
        }
      />
      {isLoading && <LoadingState />}
      <Card title="Sessions">
        <table className="data-table">
          <thead>
            <tr>
              <th>User</th><th>Page</th><th>Device</th><th>Country</th><th>LCP</th><th>Errors</th><th>Duration</th><th />
            </tr>
          </thead>
          <tbody>
            {(data ?? []).map((s) => (
              <tr key={s.id}>
                <td>{s.userId}</td>
                <td>{s.page}</td>
                <td>{s.device}</td>
                <td>{s.country}</td>
                <td>{s.lcp}s</td>
                <td>{s.errors}</td>
                <td>{Math.round(s.durationMs / 1000)}s</td>
                <td>
                  <Link to="/rum/sessions/$sessionId/replay" params={{ sessionId: s.id }}>Replay</Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
