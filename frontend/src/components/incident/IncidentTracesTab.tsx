import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { searchTrace } from '../../api/search';
import type { Incident } from '../../api/types';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { ErrorState, LoadingState } from '../ui/PageStates';

interface IncidentTracesTabProps {
  incident: Incident;
}

const TRACE_ID_RE = /\b(trace-[a-z0-9-]+|[0-9a-f]{32})\b/i;

function extractTraceIds(incident: Incident): string[] {
  const ids = new Set<string>();
  for (const ev of incident.timeline ?? []) {
    const m = ev.description?.match(TRACE_ID_RE);
    if (m) ids.add(m[1]);
  }
  for (const ev of incident.rootCauseAnalysis?.evidence ?? []) {
    if (ev.type === 'TRACE' && ev.snippet) ids.add(ev.snippet.trim());
    const m = ev.description.match(TRACE_ID_RE);
    if (m) ids.add(m[1]);
  }
  return [...ids];
}

export function IncidentTracesTab({ incident }: IncidentTracesTabProps) {
  const [traceId, setTraceId] = useState('');
  const [searchId, setSearchId] = useState('');

  const deepLinks = useMemo(() => extractTraceIds(incident), [incident]);

  const traceQuery = useQuery({
    queryKey: ['incident-trace', searchId],
    queryFn: () => searchTrace(searchId),
    enabled: searchId.length > 0,
  });

  return (
    <div className="incident-traces-tab">
      {deepLinks.length > 0 && (
        <div className="incident-traces-suggestions" style={{ marginBottom: 16 }}>
          <span className="muted">Correlated traces</span>
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginTop: 8 }}>
            {deepLinks.map((id) => (
              <Link key={id} to="/traces/$traceId" params={{ traceId: id }} className="pill">
                {id.length > 24 ? `${id.slice(0, 24)}…` : id}
              </Link>
            ))}
          </div>
        </div>
      )}

      <div className="incident-traces-tab__search">
        <Input
          placeholder="Trace ID…"
          value={traceId}
          onChange={(e) => setTraceId(e.target.value)}
          hint="Open full waterfall from evidence or paste an ID"
        />
        <Button variant="primary" onClick={() => setSearchId(traceId.trim())}>
          Load journey
        </Button>
        {traceId.trim() && (
          <Link to="/traces/$traceId" params={{ traceId: traceId.trim() }}>
            <Button variant="secondary">Open trace detail →</Button>
          </Link>
        )}
      </div>

      {traceQuery.isLoading && <LoadingState label="Loading trace journey…" />}
      {traceQuery.error && <ErrorState message="Trace not found" onRetry={() => traceQuery.refetch()} />}

      {traceQuery.data && (
        <div className="trace-journey">
          <p className="muted">
            {traceQuery.data.total} spans · {traceQuery.data.tookMs}ms ·{' '}
            <Link to="/traces/$traceId" params={{ traceId: searchId }}>View waterfall</Link>
          </p>
          {traceQuery.data.hits.map((hit, idx) => (
            <div key={hit.id} className="trace-journey__step">
              <div className="trace-journey__connector">{idx > 0 && <span />}</div>
              <div className={`log-row severity-bar severity-bar--${hit.severity}`}>
                <Badge variant="info">{hit.severity}</Badge>
                <Badge variant="info">{hit.service}</Badge>
                <span className="log-row__message">{hit.message}</span>
              </div>
            </div>
          ))}
        </div>
      )}

      {!searchId && deepLinks.length === 0 && (
        <p className="muted">
          No trace IDs in incident evidence yet. Paste a trace ID or pick one from logs with a traceId field.
        </p>
      )}
    </div>
  );
}
