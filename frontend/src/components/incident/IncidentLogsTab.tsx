import { useMemo, useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useVirtualizer } from '@tanstack/react-virtual';
import { formatDistanceToNow } from 'date-fns';
import { searchLogs } from '../../api/search';
import type { Incident, LogHit } from '../../api/types';
import { highlightMatches, hasStackTrace } from '../../utils/highlightText';
import { Badge } from '../ui/Badge';
import { ErrorState, LoadingState } from '../ui/PageStates';
import { LogDetailPanel } from '../logs/LogDetailPanel';

function severityVariant(sev: string) {
  if (sev === 'FATAL' || sev === 'CRITICAL') return 'critical';
  if (sev === 'ERROR') return 'error';
  if (sev === 'WARN') return 'warning';
  return 'info';
}

interface IncidentLogsTabProps {
  incident: Incident;
}

export function IncidentLogsTab({ incident }: IncidentLogsTabProps) {
  const [showAll, setShowAll] = useState(false);
  const [selected, setSelected] = useState<LogHit | null>(null);
  const parentRef = useRef<HTMLDivElement>(null);

  const endTime = incident.resolvedTime ?? new Date().toISOString();
  const services = incident.affectedServices ?? [];

  const logsQuery = useQuery({
    queryKey: ['incident-logs', incident.id, showAll],
    queryFn: async () => {
      const results = await Promise.all(
        (services.length ? services : ['payment-api']).map((svc) =>
          searchLogs({
            service: svc,
            severity: showAll ? undefined : 'ERROR',
            startTime: incident.startTime,
            endTime,
            size: 200,
          }),
        ),
      );
      const hits = results.flatMap((r) => r.hits);
      hits.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
      return { hits, total: hits.length, tookMs: results.reduce((s, r) => s + r.tookMs, 0) };
    },
  });

  const hits = useMemo(() => {
    if (showAll) return logsQuery.data?.hits ?? [];
    return (logsQuery.data?.hits ?? []).filter((h) => h.severity === 'ERROR' || h.severity === 'FATAL');
  }, [logsQuery.data, showAll]);

  const rowVirtualizer = useVirtualizer({
    count: hits.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 44,
    overscan: 15,
  });

  return (
    <div className={`incident-logs-tab ${selected ? 'incident-logs-tab--detail' : ''}`}>
      <div className="incident-logs-tab__toolbar">
        <p className="muted">
          Window: incident start → {incident.resolvedTime ? 'resolved' : 'now'} · Services:{' '}
          {services.join(', ') || 'all'}
        </p>
        <label className="checkbox-row">
          <input type="checkbox" checked={showAll} onChange={(e) => setShowAll(e.target.checked)} />
          Show all severities
        </label>
      </div>

      <div className="incident-logs-tab__stream" ref={parentRef}>
        {logsQuery.isLoading && <LoadingState label="Loading incident logs…" />}
        {logsQuery.error && <ErrorState message="Failed to load logs" onRetry={() => logsQuery.refetch()} />}
        {!logsQuery.isLoading && hits.length === 0 && (
          <p className="muted" style={{ padding: 16 }}>
            No logs in this window.
          </p>
        )}
        <div style={{ height: rowVirtualizer.getTotalSize(), position: 'relative' }}>
          {rowVirtualizer.getVirtualItems().map((virtualRow) => {
            const hit = hits[virtualRow.index];
            return (
              <div
                key={hit.id}
                className={`log-row severity-bar severity-bar--${hit.severity}`}
                style={{
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  width: '100%',
                  transform: `translateY(${virtualRow.start}px)`,
                }}
                onClick={() => setSelected(hit)}
              >
                <Badge variant={severityVariant(hit.severity) as 'error'}>{hit.severity.slice(0, 1)}</Badge>
                <span className="log-row__time" title={hit.timestamp}>
                  {formatDistanceToNow(new Date(hit.timestamp), { addSuffix: true })}
                </span>
                <Badge variant="info">{hit.service}</Badge>
                <span className="log-row__message">{highlightMatches(hit.message.slice(0, 120), 'error')}</span>
                {hasStackTrace(hit.message) && <span className="log-row__stack">ST</span>}
              </div>
            );
          })}
        </div>
      </div>

      <footer className="incident-logs-tab__footer muted">
        {hits.length} logs · {logsQuery.data?.tookMs ?? 0}ms
      </footer>

      {selected && (
        <div className="incident-logs-tab__detail">
          <LogDetailPanel log={selected} onClose={() => setSelected(null)} />
        </div>
      )}
    </div>
  );
}
