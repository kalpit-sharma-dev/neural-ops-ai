import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { searchTrace } from '../../api/search';
import type { Incident } from '../../api/types';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { SearchInput } from '../ui/SearchInput';
import { ErrorState, LoadingState } from '../ui/PageStates';

interface IncidentTracesTabProps {
  incident: Incident;
}

export function IncidentTracesTab({ incident }: IncidentTracesTabProps) {
  const [traceId, setTraceId] = useState('');
  const [searchId, setSearchId] = useState('');

  const traceQuery = useQuery({
    queryKey: ['incident-trace', searchId],
    queryFn: () => searchTrace(searchId),
    enabled: searchId.length > 0,
  });

  const suggestedTraces = incident.rootCauseAnalysis?.evidence
    ?.filter((e) => e.type === 'TRACE' || e.description.toLowerCase().includes('trace'))
    .slice(0, 3);

  return (
    <div className="incident-traces-tab">
      <div className="incident-traces-tab__search">
        <SearchInput
          placeholder="Search trace ID from incident evidence…"
          value={traceId}
          onChange={(e) => setTraceId(e.target.value)}
          shortcut=""
        />
        <Button variant="primary" onClick={() => setSearchId(traceId.trim())}>
          Load trace
        </Button>
      </div>

      {suggestedTraces && suggestedTraces.length > 0 && (
        <div className="incident-traces-suggestions">
          <span className="muted">From evidence:</span>
          {suggestedTraces.map((ev) => (
            <button
              key={ev.id}
              type="button"
              className="pill"
              onClick={() => {
                setTraceId(ev.description.slice(0, 36));
                setSearchId(ev.description.slice(0, 36));
              }}
            >
              {ev.description.slice(0, 40)}…
            </button>
          ))}
        </div>
      )}

      {traceQuery.isLoading && <LoadingState label="Loading trace journey…" />}
      {traceQuery.error && <ErrorState message="Trace not found" onRetry={() => traceQuery.refetch()} />}

      {traceQuery.data && (
        <div className="trace-journey">
          <p className="muted">
            {traceQuery.data.total} spans · {traceQuery.data.tookMs}ms · affected services:{' '}
            {incident.affectedServices?.join(', ')}
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

      {!searchId && (
        <p className="muted">
          Enter a correlated trace ID to view the distributed journey across{' '}
          {incident.affectedServices?.length ?? 0} affected services.
        </p>
      )}
    </div>
  );
}
