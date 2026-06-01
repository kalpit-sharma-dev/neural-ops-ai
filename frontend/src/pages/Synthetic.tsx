import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { fetchSyntheticMonitors, fetchSyntheticRuns } from '../api/observability';
import { Badge } from '../components/ui/Badge';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function Synthetic() {
  const [selected, setSelected] = useState('');
  const monitorsQuery = useQuery({ queryKey: ['synthetic-monitors'], queryFn: fetchSyntheticMonitors });
  const runsQuery = useQuery({
    queryKey: ['synthetic-runs', selected],
    queryFn: () => fetchSyntheticRuns(selected),
    enabled: !!selected,
  });

  return (
    <div>
      <PageHeader title="Synthetic Monitoring" subtitle="HTTP and browser checks" />
      {monitorsQuery.isLoading && <LoadingState />}
      {(monitorsQuery.data ?? []).map((m) => (
        <Card key={m.id} title={m.name}>
          <p>{m.url} · every {m.interval}</p>
          <Badge variant={m.lastStatus === 'OK' ? 'healthy' : 'critical'}>{m.lastStatus}</Badge>
          <button type="button" onClick={() => setSelected(m.id)}>View runs</button>
        </Card>
      ))}
      {runsQuery.data && (
        <Card title="Run history">
          <table className="data-table">
            <thead><tr><th>Status</th><th>Location</th><th>Latency</th><th>Time</th></tr></thead>
            <tbody>
              {runsQuery.data.map((r) => (
                <tr key={r.id}>
                  <td>{r.status}</td>
                  <td>{r.location}</td>
                  <td>{r.latencyMs}ms</td>
                  <td>{new Date(r.ranAt).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}
    </div>
  );
}
