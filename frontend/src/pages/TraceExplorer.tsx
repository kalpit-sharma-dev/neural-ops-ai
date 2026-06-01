import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { fetchServiceOperations, searchTraces } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { SearchInput } from '../components/ui/SearchInput';
import { ErrorState, LoadingState, PageHeader } from '../components/ui/PageStates';

export default function TraceExplorer() {
  const [service, setService] = useState('');
  const [status, setStatus] = useState('');
  const [search, setSearch] = useState(false);

  const opsQuery = useQuery({
    queryKey: ['service-operations', service],
    queryFn: () => fetchServiceOperations(service),
    enabled: service.length > 0,
  });

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['trace-search', service, status],
    queryFn: () => searchTraces({ service: service || undefined, status: status || undefined, limit: 50 }),
    enabled: search,
  });

  return (
    <div>
      <PageHeader
        title="Trace Explorer"
        subtitle="Search distributed traces and open waterfall views"
        actions={
          <Link to="/traces/compare" search={{ a: 'trace-demo-01', b: 'trace-demo-02' }}>Compare traces</Link>
        }
      />
      <div style={{ display: 'flex', gap: 12, marginBottom: 24, flexWrap: 'wrap' }}>
        <div style={{ flex: 1, minWidth: 200 }}>
          <SearchInput placeholder="Service filter…" value={service} onChange={(e) => setService(e.target.value)} shortcut="" />
        </div>
        <select value={status} onChange={(e) => setStatus(e.target.value)} className="ui-select">
          <option value="">All statuses</option>
          <option value="OK">OK</option>
          <option value="ERROR">ERROR</option>
        </select>
        <Button variant="primary" onClick={() => setSearch(true)}>Search</Button>
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
      {data && (
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
