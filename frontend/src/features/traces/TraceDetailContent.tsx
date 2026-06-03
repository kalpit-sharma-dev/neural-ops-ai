import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchTrace } from '../../api/observability';
import { getApiErrorMessage } from '../../api/client';
import { TraceWaterfall } from './TraceWaterfall';
import { TraceFlameGraph } from './TraceFlameGraph';
import { ProfilePanel } from './ProfilePanel';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { ErrorState, LoadingState } from '../../components/ui/PageStates';

type ViewMode = 'waterfall' | 'flame';

interface TraceDetailContentProps {
  traceId: string;
  showProfile?: boolean;
}

export function TraceDetailContent({ traceId, showProfile = true }: TraceDetailContentProps) {
  const [view, setView] = useState<ViewMode>('waterfall');

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['apm-trace', traceId],
    queryFn: () => fetchTrace(traceId),
    enabled: traceId.length > 0,
  });

  if (isLoading) return <LoadingState label="Loading trace…" />;
  if (error) return <ErrorState message={getApiErrorMessage(error)} onRetry={() => refetch()} />;
  if (!data) return null;

  return (
    <>
      <div className="trace-detail-content__toolbar">
        <Badge variant={data.status === 'ERROR' ? 'critical' : 'healthy'}>{data.status}</Badge>
        <span className="muted">
          {data.service} · {data.spanCount} spans · {data.totalMs}ms
        </span>
        <div className="tab-bar trace-detail-content__tabs">
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
      {showProfile && (
        <div style={{ marginTop: 24 }}>
          <ProfilePanel service={data.service} />
        </div>
      )}
    </>
  );
}
