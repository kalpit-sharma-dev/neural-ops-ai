import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from '@tanstack/react-router';
import { fetchTrace } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { TraceWaterfall } from '../features/traces/TraceWaterfall';
import { TraceFlameGraph } from '../features/traces/TraceFlameGraph';
import { ProfilePanel } from '../features/traces/ProfilePanel';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { ErrorState, LoadingState, PageHeader } from '../components/ui/PageStates';

type ViewMode = 'waterfall' | 'flame';

export default function TraceDetail() {
  const { traceId = '' } = useParams({ strict: false });
  const [view, setView] = useState<ViewMode>('waterfall');

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['apm-trace', traceId],
    queryFn: () => fetchTrace(traceId),
    enabled: traceId.length > 0,
  });

  return (
    <div>
      <PageHeader
        title={`Trace ${traceId}`}
        subtitle="PurePath-style trace analysis"
        actions={
          <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <Link to="/traces/settings">Sampling</Link>
            <Link to="/traces/compare" search={{ a: traceId, b: undefined }}>Compare</Link>
            <Link to="/traces">← Back</Link>
          </div>
        }
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message={getApiErrorMessage(error)} onRetry={() => refetch()} />}
      {data && (
        <>
          <div style={{ display: 'flex', gap: 12, marginBottom: 16, alignItems: 'center', flexWrap: 'wrap' }}>
            <Badge variant={data.status === 'ERROR' ? 'critical' : 'healthy'}>{data.status}</Badge>
            <span className="muted">{data.service} · {data.spanCount} spans · {data.totalMs}ms</span>
            <div className="tab-bar" style={{ marginLeft: 'auto' }}>
              <Button variant={view === 'waterfall' ? 'primary' : 'ghost'} size="sm" onClick={() => setView('waterfall')}>
                Waterfall
              </Button>
              <Button variant={view === 'flame' ? 'primary' : 'ghost'} size="sm" onClick={() => setView('flame')}>
                Flame graph
              </Button>
            </div>
          </div>
          {view === 'waterfall' ? (
            <TraceWaterfall spans={data.spans} traceTotalMs={data.totalMs} traceId={traceId} />
          ) : (
            <TraceFlameGraph spans={data.spans} traceTotalMs={data.totalMs} />
          )}
          <div style={{ marginTop: 24 }}>
            <ProfilePanel service={data.service} />
          </div>
        </>
      )}
    </div>
  );
}
