import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { Bar, BarChart, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { fetchServiceOperations, searchTraces } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { DomainEmptyState } from '../components/ui/DomainEmptyState';
import { SearchInput } from '../components/ui/SearchInput';
import { ErrorState, LoadingState, PageHeader } from '../components/ui/PageStates';
import { chartCartesianDefaults } from '../lib/chartTheme';
import { getChartColors } from '../lib/chartColors';

export default function TraceExplorer() {
  const [service, setService] = useState('');
  const [status, setStatus] = useState('');

  const opsQuery = useQuery({
    queryKey: ['service-operations', service],
    queryFn: () => fetchServiceOperations(service),
    enabled: service.length > 0,
  });

  const { data, isLoading, error, refetch, isFetched } = useQuery({
    queryKey: ['trace-search', service, status],
    queryFn: () => searchTraces({ service: service || undefined, status: status || undefined, limit: 50 }),
  });

  const heatmapData = useMemo(() => {
    const buckets = new Map<string, { service: string; count: number; errorCount: number }>();
    for (const t of data ?? []) {
      const row = buckets.get(t.service) ?? { service: t.service, count: 0, errorCount: 0 };
      row.count += 1;
      if (t.status === 'ERROR') row.errorCount += 1;
      buckets.set(t.service, row);
    }
    return [...buckets.values()].sort((a, b) => b.count - a.count).slice(0, 12);
  }, [data]);

  const chartDefaults = chartCartesianDefaults();
  const colors = getChartColors();

  return (
    <div>
      <PageHeader
        title="Trace Explorer"
        subtitle="Search distributed traces and open waterfall views"
        actions={<Link to="/traces/compare" search={{ a: undefined, b: undefined }}>Compare traces</Link>}
      />
      <div style={{ display: 'flex', gap: 12, marginBottom: 24, flexWrap: 'wrap' }}>
        <div style={{ flex: 1, minWidth: 200 }}>
          <SearchInput placeholder="Service filter…" value={service} onChange={(e) => setService(e.target.value)} shortcut="" />
        </div>
        <select value={status} onChange={(e) => setStatus(e.target.value)} className="ui-select" aria-label="Status filter">
          <option value="">All statuses</option>
          <option value="OK">OK</option>
          <option value="ERROR">ERROR</option>
        </select>
        <Button variant="secondary" onClick={() => refetch()}>Refresh</Button>
      </div>

      {service && opsQuery.data && opsQuery.data.length > 0 && (
        <div className="ui-card" style={{ marginBottom: 16 }}>
          <p className="muted">Operations for {service}</p>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
            {opsQuery.data.map((op) => (
              <Badge key={op} variant="info">{op}</Badge>
            ))}
          </div>
        </div>
      )}

      {isLoading && <LoadingState />}
      {error && <ErrorState message={getApiErrorMessage(error)} onRetry={() => refetch()} />}
      {isFetched && (data?.length ?? 0) === 0 && !error && <DomainEmptyState domain="traces" />}

      {heatmapData.length > 0 && (
        <div className="ui-card" style={{ marginBottom: 24 }}>
          <h3 style={{ marginTop: 0 }}>Latency heatmap (trace volume by service)</h3>
          <p className="muted">Bar height = trace count · color intensity = error ratio</p>
          <ResponsiveContainer width="100%" height={220}>
            <BarChart data={heatmapData} layout="vertical" margin={{ left: 100 }}>
              <XAxis type="number" {...chartDefaults.axis} />
              <YAxis type="category" dataKey="service" width={96} {...chartDefaults.axis} />
              <Tooltip contentStyle={chartDefaults.tooltipStyle} />
              <Bar dataKey="count" name="Traces">
                {heatmapData.map((entry, i) => (
                  <Cell
                    key={entry.service}
                    fill={colors[i % colors.length]}
                    fillOpacity={0.4 + (entry.errorCount / Math.max(entry.count, 1)) * 0.6}
                  />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}

      {data && data.length > 0 && (
        <div className="ui-card">
          <p className="muted">{data.length} traces</p>
          {data.map((t) => (
            <Link key={t.traceId} to="/traces/$traceId" params={{ traceId: t.traceId }} className="log-row">
              <Badge variant={t.status === 'ERROR' ? 'critical' : 'healthy'}>{t.status}</Badge>
              <span>{t.service}</span>
              <span className="log-row__message">{t.operation}</span>
              <span className="muted">{t.durationMs}ms · {t.spanCount} spans</span>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
