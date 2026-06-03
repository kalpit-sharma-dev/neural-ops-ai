import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchCloudDashboards, fetchCloudMetrics, fetchServerlessFunctions } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Badge } from '../components/ui/Badge';
import { Card } from '../components/ui/Card';
import { EmptyState } from '../components/ui/EmptyState';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';

const PROVIDERS = ['aws', 'azure', 'gcp'] as const;

export default function CloudMonitoring() {
  const [provider, setProvider] = useState<(typeof PROVIDERS)[number]>('aws');
  const [selectedMetric, setSelectedMetric] = useState<string | null>(null);
  const { data, isLoading, error, refetch, isFetched } = useQuery({
    queryKey: ['cloud-dashboards', provider],
    queryFn: () => fetchCloudDashboards(provider),
  });
  const { data: series, isLoading: metricsLoading } = useQuery({
    queryKey: ['cloud-metrics', provider, selectedMetric],
    queryFn: () => fetchCloudMetrics(provider, selectedMetric!),
    enabled: !!selectedMetric,
  });

  const { data: serverless, isLoading: serverlessLoading } = useQuery({
    queryKey: ['serverless-functions'],
    queryFn: fetchServerlessFunctions,
  });

  return (
    <StitchPageShell title="Cloud monitoring" subtitle="AWS, Azure, and GCP observability dashboards">
      <div className="tab-bar" style={{ marginBottom: 24 }}>
        {PROVIDERS.map((p) => (
          <button
            key={p}
            type="button"
            className={`tab-bar__item ${provider === p ? 'tab-bar__item--active' : ''}`}
            onClick={() => { setProvider(p); setSelectedMetric(null); }}
          >
            {p.toUpperCase()}
          </button>
        ))}
      </div>
      {isLoading && <LoadingState />}
      {error && <ErrorState message={getApiErrorMessage(error)} onRetry={() => refetch()} />}
      {!error && isFetched && (data?.length ?? 0) === 0 && (
        <EmptyState title={`No ${provider.toUpperCase()} dashboards`} description="Connect this cloud provider integration to import observability dashboards." />
      )}
      {(data ?? []).map((d) => (
        <Card key={d.id} title={d.name}>
          <p className="muted">Region: {d.region}</p>
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginTop: 12 }}>
            {d.metrics.map((m) => (
              <button
                key={m}
                type="button"
                className={`tab-bar__item ${selectedMetric === m ? 'tab-bar__item--active' : ''}`}
                onClick={() => setSelectedMetric(m)}
              >
                <Badge variant="info">{m}</Badge>
              </button>
            ))}
          </div>
        </Card>
      ))}
      {selectedMetric && (
        <Card title={`Metric: ${selectedMetric}`} style={{ marginTop: 24 }}>
          {metricsLoading && <LoadingState />}
          {series && (
            <>
              <p className="muted">Source: {series.source} · {series.points.length} points</p>
              <div className="metric-sparkline" style={{ display: 'flex', alignItems: 'flex-end', gap: 2, height: 80, marginTop: 12 }}>
                {series.points.map((p) => (
                  <div
                    key={p.timestamp}
                    title={`${p.timestamp}: ${p.value}`}
                    style={{
                      flex: 1,
                      background: 'var(--accent)',
                      height: `${Math.min(100, (p.value / Math.max(...series.points.map((x) => x.value), 1)) * 100)}%`,
                      minHeight: 2,
                    }}
                  />
                ))}
              </div>
            </>
          )}
        </Card>
      )}

      <Card title="Serverless functions" style={{ marginTop: 24 }} data-testid="serverless-functions">
        {serverlessLoading && <LoadingState />}
        {!serverlessLoading && (
          <table className="data-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Provider</th>
                <th>Runtime</th>
                <th>Invocations (24h)</th>
                <th>Error %</th>
                <th>P95 ms</th>
                <th>Cold start %</th>
              </tr>
            </thead>
            <tbody>
              {(serverless ?? []).map((fn) => (
                <tr key={fn.id}>
                  <td>{fn.name}</td>
                  <td>{fn.provider}</td>
                  <td>{fn.runtime}</td>
                  <td>{fn.invocations24h.toLocaleString()}</td>
                  <td>{fn.errorRatePct.toFixed(2)}</td>
                  <td>{fn.p95DurationMs.toFixed(0)}</td>
                  <td>{fn.coldStartPct.toFixed(1)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Card>
    </StitchPageShell>
  );
}
