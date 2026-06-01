import { useCallback, useEffect, useMemo, useState } from 'react';
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
import { chartCartesianDefaults } from '../lib/chartTheme';

type Tile = NonNullable<Awaited<ReturnType<typeof fetchDashboard>>['tiles']>[number];

function layoutKey(dashboardId: string) {
  return `neuralops-dashboard-layout-${dashboardId}`;
}

export default function DashboardView() {
  const { id = '' } = useParams({ strict: false });
  const search = useSearch({ strict: false }) as { edit?: string };
  const editMode = search.edit === '1';
  const queryClient = useQueryClient();
  const [dragIndex, setDragIndex] = useState<number | null>(null);
  const [tileOrder, setTileOrder] = useState<string[] | null>(null);

  const dashQuery = useQuery({ queryKey: ['dashboard', id], queryFn: () => fetchDashboard(id), enabled: !!id });

  useEffect(() => {
    if (!id || !dashQuery.data) return;
    try {
      const saved = localStorage.getItem(layoutKey(id));
      if (saved) setTileOrder(JSON.parse(saved) as string[]);
      else setTileOrder(dashQuery.data.tiles.map((t) => t.id));
    } catch {
      setTileOrder(dashQuery.data.tiles.map((t) => t.id));
    }
  }, [id, dashQuery.data]);

  const orderedTiles = useMemo(() => {
    const tiles = dashQuery.data?.tiles ?? [];
    if (!tileOrder) return tiles;
    const byId = new Map(tiles.map((t) => [t.id, t]));
    return tileOrder.map((tid) => byId.get(tid)).filter((t): t is Tile => Boolean(t));
  }, [dashQuery.data?.tiles, tileOrder]);

  const persistOrder = useCallback(
    (order: string[]) => {
      setTileOrder(order);
      localStorage.setItem(layoutKey(id), JSON.stringify(order));
    },
    [id],
  );

  const saveMut = useMutation({
    mutationFn: (tiles: Tile[]) => updateDashboard(id, { ...dashQuery.data!, tiles }),
    onSuccess: () => {
      toast.success('Dashboard saved');
      void queryClient.invalidateQueries({ queryKey: ['dashboard', id] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const metricTiles = orderedTiles.filter((t) => t.type === 'metric' || t.type === 'timeseries');
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

  const chartDefaults = chartCartesianDefaults();

  const onDrop = (targetIndex: number) => {
    if (dragIndex == null || dragIndex === targetIndex || !tileOrder) return;
    const next = [...tileOrder];
    const [moved] = next.splice(dragIndex, 1);
    next.splice(targetIndex, 0, moved);
    persistOrder(next);
    setDragIndex(null);
    const byId = new Map((dashQuery.data?.tiles ?? []).map((t) => [t.id, t]));
    const reordered = next.map((tid) => byId.get(tid)).filter((t): t is Tile => Boolean(t));
    saveMut.mutate(reordered);
  };

  return (
    <div>
      <PageHeader
        title={dashQuery.data?.name ?? 'Dashboard'}
        subtitle={dashQuery.data?.description ?? 'Drag tiles to reorder — layout saved locally'}
        actions={
          <div style={{ display: 'flex', gap: 8 }}>
            {editMode ? (
              <Link to="/dashboards/$id" params={{ id }} search={{ edit: undefined }}>Done editing</Link>
            ) : (
              <Link to="/dashboards/$id" params={{ id }} search={{ edit: '1' }}>
                <Button variant="secondary" size="sm">Edit tiles (builder)</Button>
              </Link>
            )}
            <Link to="/dashboards">← All dashboards</Link>
          </div>
        }
      />
      {dashQuery.isLoading && <LoadingState />}
      {dashQuery.error && <ErrorState message="Dashboard not found" onRetry={() => dashQuery.refetch()} />}
      {dashQuery.data && editMode && (
        <DashboardBuilder
          dashboard={dashQuery.data}
          saving={saveMut.isPending}
          onSave={(t) => saveMut.mutate(t)}
        />
      )}
      {dashQuery.data && !editMode && (
        <div className="dashboard-builder__grid" style={{ gridTemplateColumns: 'repeat(12, 1fr)' }}>
          {orderedTiles.map((tile, index) => (
            <div
              key={tile.id}
              draggable
              onDragStart={() => setDragIndex(index)}
              onDragOver={(e) => e.preventDefault()}
              onDrop={() => onDrop(index)}
              className={`dashboard-tile-draggable ${dragIndex === index ? 'dashboard-tile-draggable--dragging' : ''}`}
              style={{ gridColumn: `span ${tile.position?.w ?? 6}`, gridRow: `span ${tile.position?.h ?? 4}` }}
            >
              {tile.type === 'stat' ? (
                <MetricCard label={tile.title} value="1.2%" trend={1.2} />
              ) : (
                <Card title={tile.title}>
                  {metricQuery.isLoading ? <LoadingState /> : (
                    <ResponsiveContainer width="100%" height={220}>
                      <LineChart data={chartData}>
                        <XAxis dataKey="time" {...chartDefaults.axis} />
                        <YAxis {...chartDefaults.axis} />
                        <Tooltip contentStyle={chartDefaults.tooltipStyle} />
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
