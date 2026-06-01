import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useParams, useSearch } from '@tanstack/react-router';
import toast from 'react-hot-toast';
import { fetchDashboard, queryMetric, updateDashboard } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { DashboardBuilder } from '../features/dashboard/DashboardBuilder';
import { Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { Card } from '../components/ui/Card';
import { MetricCard } from '../components/ui/MetricCard';
import { Button } from '../components/ui/Button';
import { ErrorState, LoadingState, PageHeader } from '../components/ui/PageStates';

export default function DashboardView() {
  const { id = '' } = useParams({ strict: false });
  const search = useSearch({ strict: false }) as { edit?: string };
  const editMode = search.edit === '1';
  const queryClient = useQueryClient();

  const dashQuery = useQuery({ queryKey: ['dashboard', id], queryFn: () => fetchDashboard(id), enabled: !!id });

  const saveMut = useMutation({
    mutationFn: (tiles: NonNullable<typeof dashQuery.data>['tiles']) =>
      updateDashboard(id, { ...dashQuery.data!, tiles }),
    onSuccess: () => {
      toast.success('Dashboard saved');
      void queryClient.invalidateQueries({ queryKey: ['dashboard', id] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const tiles = dashQuery.data?.tiles ?? [];
  const metricTiles = tiles.filter((t) => t.type === 'metric' || t.type === 'timeseries');
  const firstMetric = metricTiles[0];

  const metricQuery = useQuery({
    queryKey: ['dash-metric', firstMetric?.metric],
    queryFn: () => queryMetric(firstMetric!.metric!, 'payment-service'),
    enabled: !!firstMetric?.metric && !editMode,
  });

  const chartData = (metricQuery.data?.points ?? []).map((p) => ({
    time: new Date(p.timestamp).toLocaleTimeString(),
    value: p.value,
  }));

  return (
    <div>
      <PageHeader
        title={dashQuery.data?.name ?? 'Dashboard'}
        subtitle={dashQuery.data?.description}
        actions={
          <div style={{ display: 'flex', gap: 8 }}>
            {editMode ? (
              <Link to="/dashboards/$id" params={{ id }} search={{ edit: undefined }}>Done editing</Link>
            ) : (
              <Link to="/dashboards/$id" params={{ id }} search={{ edit: '1' }}>
                <Button variant="secondary" size="sm">Edit layout</Button>
              </Link>
            )}
            <Link to="/dashboards">← All dashboards</Link>
          </div>
        }
      />
      {dashQuery.isLoading && <LoadingState />}
      {dashQuery.error && <ErrorState message="Dashboard not found" />}
      {dashQuery.data && editMode && (
        <DashboardBuilder
          dashboard={dashQuery.data}
          saving={saveMut.isPending}
          onSave={(t) => saveMut.mutate(t)}
        />
      )}
      {dashQuery.data && !editMode && (
        <div className="dashboard-builder__grid" style={{ gridTemplateColumns: 'repeat(12, 1fr)' }}>
          {tiles.map((tile) => (
            <div
              key={tile.id}
              style={{ gridColumn: `span ${tile.position?.w ?? 6}`, gridRow: `span ${tile.position?.h ?? 4}` }}
            >
              {tile.type === 'stat' ? (
                <MetricCard label={tile.title} value="1.2%" trend={1.2} />
              ) : (
                <Card title={tile.title}>
                  {metricQuery.isLoading ? <LoadingState /> : (
                    <ResponsiveContainer width="100%" height={220}>
                      <LineChart data={chartData}>
                        <XAxis dataKey="time" stroke="var(--text-muted)" fontSize={10} />
                        <YAxis stroke="var(--text-muted)" fontSize={10} />
                        <Tooltip />
                        <Line type="monotone" dataKey="value" stroke="var(--accent-primary)" dot={false} />
                      </LineChart>
                    </ResponsiveContainer>
                  )}
                </Card>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
