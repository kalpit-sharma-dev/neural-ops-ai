import { useQuery } from '@tanstack/react-query';
import { Link, useSearch } from '@tanstack/react-router';
import { fetchTrace } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { TraceWaterfall } from '../features/traces/TraceWaterfall';
import { Badge } from '../components/ui/Badge';
import { ErrorState, LoadingState, PageHeader } from '../components/ui/PageStates';

export default function TraceCompare() {
  const search = useSearch({ strict: false }) as { a?: string; b?: string };
  const traceA = search.a ?? '';
  const traceB = search.b ?? '';

  const queryA = useQuery({
    queryKey: ['apm-trace', traceA],
    queryFn: () => fetchTrace(traceA),
    enabled: traceA.length > 0,
  });

  const queryB = useQuery({
    queryKey: ['apm-trace', traceB],
    queryFn: () => fetchTrace(traceB),
    enabled: traceB.length > 0,
  });

  const loading = queryA.isLoading || queryB.isLoading;
  const error = queryA.error ?? queryB.error;

  return (
    <div>
      <PageHeader
        title="Compare traces"
        subtitle="Side-by-side waterfall comparison"
        actions={<Link to="/traces">← Trace explorer</Link>}
      />

      {(!traceA || !traceB) && (
        <p className="muted">
          Pass two trace IDs via query params: <code>/traces/compare?a=trace-demo-01&amp;b=trace-demo-02</code>
        </p>
      )}

      {loading && <LoadingState />}
      {error && <ErrorState message={getApiErrorMessage(error)} />}

      {queryA.data && queryB.data && (
        <div className="trace-compare-grid">
          <section>
            <div style={{ display: 'flex', gap: 8, marginBottom: 12, alignItems: 'center' }}>
              <h3 style={{ margin: 0 }}>{traceA}</h3>
              <Badge variant={queryA.data.status === 'ERROR' ? 'critical' : 'healthy'}>{queryA.data.status}</Badge>
              <span className="muted">{queryA.data.totalMs}ms</span>
            </div>
            <TraceWaterfall spans={queryA.data.spans} traceTotalMs={queryA.data.totalMs} />
          </section>
          <section>
            <div style={{ display: 'flex', gap: 8, marginBottom: 12, alignItems: 'center' }}>
              <h3 style={{ margin: 0 }}>{traceB}</h3>
              <Badge variant={queryB.data.status === 'ERROR' ? 'critical' : 'healthy'}>{queryB.data.status}</Badge>
              <span className="muted">{queryB.data.totalMs}ms</span>
            </div>
            <TraceWaterfall spans={queryB.data.spans} traceTotalMs={queryB.data.totalMs} />
          </section>
        </div>
      )}
    </div>
  );
}
