import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchAPMCodeErrors, fetchTrace } from '../../api/observability';
import { getApiErrorMessage } from '../../api/client';
import { TraceWaterfall } from './TraceWaterfall';
import { TraceFlameGraph } from './TraceFlameGraph';
import { ProfilePanel } from './ProfilePanel';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { Card } from '../../components/ui/Card';
import { ErrorState, LoadingState } from '../../components/ui/PageStates';
import { DataExportMenu } from '../../components/ui/DataExportMenu';
import { useI18n } from '../../i18n/I18nProvider';

type ViewMode = 'waterfall' | 'flame';

interface TraceDetailContentProps {
  traceId: string;
  showProfile?: boolean;
}

export function TraceDetailContent({ traceId, showProfile = true }: TraceDetailContentProps) {
  const { t } = useI18n();
  const [view, setView] = useState<ViewMode>('waterfall');

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['apm-trace', traceId],
    queryFn: () => fetchTrace(traceId),
    enabled: traceId.length > 0,
  });

  const codeErrorsQuery = useQuery({
    queryKey: ['apm-code-errors', data?.service],
    queryFn: () => fetchAPMCodeErrors(data!.service),
    enabled: !!data?.service && data.status === 'ERROR',
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
          <DataExportMenu getData={() => [data]} filenamePrefix={`trace-${traceId}`} />
        </div>
      </div>
      {view === 'waterfall' ? (
        <TraceWaterfall spans={data.spans} traceTotalMs={data.totalMs} traceId={traceId} />
      ) : (
        <TraceFlameGraph spans={data.spans} traceTotalMs={data.totalMs} />
      )}
      {data.status === 'ERROR' && (
        <Card title={t('page.apm.codeErrors')} style={{ marginTop: 24 }}>
          {codeErrorsQuery.isLoading && <LoadingState label={t('common.loading')} />}
          {!codeErrorsQuery.isLoading && (codeErrorsQuery.data?.length ?? 0) === 0 && (
            <p className="muted">{t('page.apm.codeErrorsEmpty')}</p>
          )}
          {(codeErrorsQuery.data ?? []).map((err) => (
            <div key={`${err.file}:${err.line}`} className="list-row" data-testid="apm-code-error">
              <span>
                <strong>{err.file}:{err.line}</strong>
                <span className="muted" style={{ marginLeft: 8 }}>{err.message}</span>
              </span>
              <span className="muted">
                {err.count}× · {err.release ?? '—'} · {err.commitSha?.slice(0, 7) ?? '—'}
              </span>
            </div>
          ))}
        </Card>
      )}
      {showProfile && (
        <div style={{ marginTop: 24 }}>
          <ProfilePanel service={data.service} />
        </div>
      )}
    </>
  );
}
