import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import toast from 'react-hot-toast';
import { Link } from '@tanstack/react-router';
import { createDashboard, fetchDashboards } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';

export default function Dashboards() {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({ queryKey: ['dashboards'], queryFn: fetchDashboards });
  const [name, setName] = useState('');

  const createMut = useMutation({
    mutationFn: () =>
      createDashboard({
        name,
        description: 'Custom observability dashboard',
        shared: false,
        tiles: [
          { id: 't1', type: 'metric', title: 'Latency P99', metric: 'latency_p99', position: { x: 0, y: 0, w: 6, h: 4 } },
          { id: 't2', type: 'metric', title: 'Error rate', metric: 'error_rate', position: { x: 6, y: 0, w: 6, h: 4 } },
          { id: 't3', type: 'metric', title: 'Throughput', metric: 'throughput', position: { x: 0, y: 4, w: 6, h: 4 } },
          { id: 't4', type: 'metric', title: 'CPU', metric: 'cpu_usage', position: { x: 6, y: 4, w: 6, h: 4 } },
        ],
      }),
    onSuccess: () => {
      toast.success('Dashboard created');
      void queryClient.invalidateQueries({ queryKey: ['dashboards'] });
      setName('');
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  return (
    <StitchPageShell
      title="Dashboards"
      subtitle="Custom observability dashboards"
      actions={
        <div style={{ display: 'flex', gap: 8 }}>
          <input placeholder="Dashboard name" value={name} onChange={(e) => setName(e.target.value)} />
          <Button variant="primary" disabled={!name} onClick={() => createMut.mutate()}>
            Create dashboard
          </Button>
        </div>
      }
    >
      {isLoading && <LoadingState />}
      <div className="dashboard-row-3">
        {(data ?? []).map((d) => (
          <Card key={d.id} title={d.name}>
            <p className="muted">{d.description ?? `${d.tiles.length} tiles`}</p>
            {d.shared && <Badge variant="info">Shared</Badge>}
            <Link to="/dashboards/$id" params={{ id: d.id }} search={{ edit: undefined }}>Open →</Link>
          </Card>
        ))}
      </div>
    </StitchPageShell>
  );
}
