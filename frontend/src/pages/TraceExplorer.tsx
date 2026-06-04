import { useMemo, useState, type KeyboardEvent } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { AnimatePresence, motion } from 'framer-motion';
import { Bar, BarChart, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { ExternalLink } from 'lucide-react';
import { DataExportMenu } from '../components/ui/DataExportMenu';
import { fetchServiceOperations, searchTraces, type TraceSummary } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { TraceDetailPanel } from '../components/traces/TraceDetailPanel';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { DomainEmptyState } from '../components/ui/DomainEmptyState';
import { SearchInput } from '../components/ui/SearchInput';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';
import { chartCartesianDefaults } from '../lib/chartTheme';
import { getChartColors } from '../lib/chartColors';
import { useTimeBounds } from '../hooks/useTimeBounds';

function selectTraceRow(e: KeyboardEvent, trace: TraceSummary, onSelect: (t: TraceSummary) => void) {
  if (e.key === 'Enter' || e.key === ' ') {
    e.preventDefault();
    onSelect(trace);
  }
}

export default function TraceExplorer() {
  const [service, setService] = useState('');
  const [status, setStatus] = useState('');
  const [selected, setSelected] = useState<TraceSummary | null>(null);
  const { key: rangeKey, resolve: resolveRange } = useTimeBounds();

  const opsQuery = useQuery({
    queryKey: ['service-operations', service],
    queryFn: () => fetchServiceOperations(service),
    enabled: service.length > 0,
  });

  const { data, isLoading, error, refetch, isFetched } = useQuery({
    queryKey: ['trace-search', service, status, rangeKey],
    queryFn: () => {
      const { startIso, endIso } = resolveRange();
      return searchTraces({
        service: service || undefined,
        status: status || undefined,
        start: startIso,
        end: endIso,
        limit: 50,
      });
    },
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
    <StitchPageShell
      title="Trace Explorer"
      subtitle="Search distributed traces and open waterfall views"
      actions={
        <>
          <DataExportMenu getData={() => data ?? []} filenamePrefix="traces" disabled={!data?.length} />
          <Link to="/traces/compare" search={{ a: undefined, b: undefined }}>Compare traces</Link>
        </>
      }
    >
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
        <div className={`trace-explorer ${selected ? 'trace-explorer--detail' : ''}`}>
          <div className="ui-card trace-explorer__list">
            <p className="muted">{data.length} traces · click a row for inline detail</p>
            {data.map((t) => (
              <div
                key={t.traceId}
                role="button"
                tabIndex={0}
                className={`trace-row ${selected?.traceId === t.traceId ? 'trace-row--selected' : ''}`}
                onClick={() => setSelected(t)}
                onKeyDown={(e) => selectTraceRow(e, t, setSelected)}
              >
                <Badge variant={t.status === 'ERROR' ? 'critical' : 'healthy'} className="trace-row__status">
                  {t.status}
                </Badge>
                <span className="trace-row__service">{t.service}</span>
                <span className="trace-row__operation">{t.operation}</span>
                <span className="trace-row__meta muted">
                  {t.durationMs}ms · {t.spanCount} spans
                </span>
                <Link
                  to="/traces/$traceId"
                  params={{ traceId: t.traceId }}
                  className="trace-row__open"
                  title="Open full trace page"
                  onClick={(e) => e.stopPropagation()}
                >
                  <ExternalLink size={14} aria-hidden />
                  Open
                </Link>
              </div>
            ))}
          </div>

          <AnimatePresence>
            {selected && (
              <motion.div
                className="trace-detail-panel-wrap"
                initial={{ x: 40, opacity: 0 }}
                animate={{ x: 0, opacity: 1 }}
                exit={{ x: 40, opacity: 0 }}
              >
                <TraceDetailPanel trace={selected} onClose={() => setSelected(null)} />
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      )}
    </StitchPageShell>
  );
}
