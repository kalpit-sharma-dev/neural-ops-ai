import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { Line, LineChart, ResponsiveContainer } from 'recharts';
import { fetchHosts } from '../api/observability';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

function sparkline(points: number[]) {
  return points.map((value, i) => ({ i, value }));
}

export default function Infrastructure() {
  const { data, isLoading } = useQuery({ queryKey: ['infra-hosts'], queryFn: fetchHosts });
  const [drawerHostId, setDrawerHostId] = useState<string | null>(null);

  const selected = useMemo(
    () => (data ?? []).find((h) => h.id === drawerHostId),
    [data, drawerHostId],
  );

  return (
    <div>
      <PageHeader title="Infrastructure" subtitle="Host monitoring" actions={<Link to="/kubernetes">Kubernetes →</Link>} />
      {isLoading && <LoadingState />}
      <Card title="Hosts">
        <table className="data-table">
          <thead><tr><th>Host</th><th>Zone</th><th>Status</th><th>CPU</th><th>Memory</th><th>Disk</th><th /></tr></thead>
          <tbody>
            {(data ?? []).map((h) => (
              <tr key={h.id}>
                <td>{h.name}</td>
                <td>{h.zone}</td>
                <td><Badge variant={h.status === 'healthy' ? 'healthy' : 'warning'}>{h.status}</Badge></td>
                <td>{h.cpuPercent}%</td>
                <td>{h.memoryPercent}%</td>
                <td>{h.diskPercent}%</td>
                <td>
                  <Button variant="ghost" size="sm" onClick={() => setDrawerHostId(h.id)}>Details</Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>

      {selected && (
        <aside className="infra-host-drawer" aria-label={`Host ${selected.name}`}>
          <div className="infra-host-drawer__header">
            <h3>{selected.name}</h3>
            <Button variant="ghost" size="sm" onClick={() => setDrawerHostId(null)}>Close</Button>
          </div>
          <p className="muted">{selected.zone} · <Badge variant={selected.status === 'healthy' ? 'healthy' : 'warning'}>{selected.status}</Badge></p>
          <div className="infra-sparklines">
            {[
              { label: 'CPU %', value: selected.cpuPercent },
              { label: 'Memory %', value: selected.memoryPercent },
              { label: 'Disk %', value: selected.diskPercent },
            ].map(({ label, value }) => {
              const points = Array.from({ length: 12 }, (_, i) =>
                Math.max(0, Math.min(100, value + Math.sin(i + value) * 8)),
              );
              return (
                <div key={label} className="infra-sparkline-card">
                  <span className="muted">{label}</span>
                  <strong>{value}%</strong>
                  <ResponsiveContainer width="100%" height={48}>
                    <LineChart data={sparkline(points)}>
                      <Line type="monotone" dataKey="value" stroke="var(--accent-primary)" dot={false} strokeWidth={2} />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              );
            })}
          </div>
          <Link to="/entities/$type/$id" params={{ type: 'host', id: selected.id }}>Open entity →</Link>
        </aside>
      )}
    </div>
  );
}
