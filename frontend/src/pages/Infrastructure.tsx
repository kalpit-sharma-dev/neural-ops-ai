import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { fetchHosts } from '../api/observability';
import { Badge } from '../components/ui/Badge';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function Infrastructure() {
  const { data, isLoading } = useQuery({ queryKey: ['infra-hosts'], queryFn: fetchHosts });

  return (
    <div>
      <PageHeader title="Infrastructure" subtitle="Host monitoring" actions={<Link to="/kubernetes">Kubernetes →</Link>} />
      {isLoading && <LoadingState />}
      <Card title="Hosts">
        <table className="data-table">
          <thead><tr><th>Host</th><th>Zone</th><th>Status</th><th>CPU</th><th>Memory</th><th>Disk</th></tr></thead>
          <tbody>
            {(data ?? []).map((h) => (
              <tr key={h.id}>
                <td>{h.name}</td>
                <td>{h.zone}</td>
                <td><Badge variant={h.status === 'healthy' ? 'healthy' : 'warning'}>{h.status}</Badge></td>
                <td>{h.cpuPercent}%</td>
                <td>{h.memoryPercent}%</td>
                <td>{h.diskPercent}%</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
