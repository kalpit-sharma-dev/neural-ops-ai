import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { Bar, BarChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { fetchRUMSessions } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Card } from '../components/ui/Card';
import { DomainEmptyState } from '../components/ui/DomainEmptyState';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { StitchPageShell, KpiRow } from '../components/stitch';
import { DataExportMenu } from '../components/ui/DataExportMenu';
import { chartCartesianDefaults } from '../lib/chartTheme';

export default function RUM() {
  const { data, isLoading, error, refetch, isFetched } = useQuery({ queryKey: ['rum-sessions'], queryFn: fetchRUMSessions });
  const chartDefaults = chartCartesianDefaults();

  const vitals = useMemo(() => {
    const sessions = data ?? [];
    if (sessions.length === 0) return { lcp: [], fid: [], cls: [] };
    const bucket = (values: number[], label: string) => {
      const sorted = [...values].sort((a, b) => a - b);
      const p50 = sorted[Math.floor(sorted.length * 0.5)] ?? 0;
      const p75 = sorted[Math.floor(sorted.length * 0.75)] ?? 0;
      const p95 = sorted[Math.floor(sorted.length * 0.95)] ?? 0;
      return [
        { bucket: 'p50', value: p50, label },
        { bucket: 'p75', value: p75, label },
        { bucket: 'p95', value: p95, label },
      ];
    };
    const lcpVals = sessions.map((s) => s.lcp);
    const fidVals = sessions.map((s) => (s as { fid?: number }).fid ?? s.durationMs / 10000);
    const clsVals = sessions.map((s) => (s as { cls?: number }).cls ?? s.errors * 0.01);
    return {
      lcp: bucket(lcpVals, 'LCP (s)'),
      fid: bucket(fidVals, 'FID (ms)'),
      cls: bucket(clsVals, 'CLS'),
    };
  }, [data]);

  const hasSessions = (data?.length ?? 0) > 0;

  return (
    <StitchPageShell
      title="Real User Monitoring"
      subtitle="Browser sessions, Core Web Vitals, session replay"
      actions={
        <>
          <DataExportMenu getData={() => data ?? []} filenamePrefix="rum-sessions" disabled={!data?.length} />
          <a href="/rum/neuralops-rum.js" download="neuralops-rum.js" className="muted">
            Download RUM SDK
          </a>
        </>
      }
    >
      {isLoading && <LoadingState />}
      {error && <ErrorState message={getApiErrorMessage(error)} onRetry={() => refetch()} />}
      {!isLoading && !error && !hasSessions && isFetched && <DomainEmptyState domain="rum" />}

      {hasSessions && (
        <>
          <KpiRow columns={3}>
            {[
              { title: 'LCP', data: vitals.lcp },
              { title: 'FID (proxy)', data: vitals.fid },
              { title: 'CLS (proxy)', data: vitals.cls },
            ].map(({ title, data: series }) => (
              <Card key={title} title={title}>
                <ResponsiveContainer width="100%" height={160}>
                  <BarChart data={series}>
                    <XAxis dataKey="bucket" {...chartDefaults.axis} />
                    <YAxis {...chartDefaults.axis} />
                    <Tooltip contentStyle={chartDefaults.tooltipStyle} />
                    <Bar dataKey="value" fill="var(--accent-primary)" />
                  </BarChart>
                </ResponsiveContainer>
              </Card>
            ))}
          </KpiRow>

          <Card title="Sessions" style={{ marginTop: 24 }}>
            <table className="data-table">
              <thead>
                <tr>
                  <th>User</th><th>Page</th><th>Device</th><th>Country</th><th>LCP</th><th>Errors</th><th>Duration</th><th />
                </tr>
              </thead>
              <tbody>
                {(data ?? []).map((s) => (
                  <tr key={s.id}>
                    <td>{s.userId}</td>
                    <td>{s.page}</td>
                    <td>{s.device}</td>
                    <td>{s.country}</td>
                    <td>{s.lcp}s</td>
                    <td>{s.errors}</td>
                    <td>{Math.round(s.durationMs / 1000)}s</td>
                    <td>
                      <Link to="/rum/sessions/$sessionId/replay" params={{ sessionId: s.id }}>Replay</Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
        </>
      )}
    </StitchPageShell>
  );
}
