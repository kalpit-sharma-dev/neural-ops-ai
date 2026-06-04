import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { fetchSyntheticMonitors, fetchSyntheticRuns } from '../api/observability';
import { Badge } from '../components/ui/Badge';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';
import { DataExportMenu } from '../components/ui/DataExportMenu';

export default function Synthetic() {
  const [selected, setSelected] = useState('');
  const monitorsQuery = useQuery({ queryKey: ['synthetic-monitors'], queryFn: fetchSyntheticMonitors });
  const runsQuery = useQuery({
    queryKey: ['synthetic-runs', selected],
    queryFn: () => fetchSyntheticRuns(selected),
    enabled: !!selected,
  });

  const locationBars = useMemo(() => {
    const byLoc = new Map<string, { location: string; latencyMs: number; count: number }>();
    for (const r of runsQuery.data ?? []) {
      const row = byLoc.get(r.location) ?? { location: r.location, latencyMs: 0, count: 0 };
      row.latencyMs += r.latencyMs;
      row.count += 1;
      byLoc.set(r.location, row);
    }
    return [...byLoc.values()]
      .map((r) => ({ ...r, avgMs: Math.round(r.latencyMs / Math.max(r.count, 1)) }))
      .sort((a, b) => b.avgMs - a.avgMs);
  }, [runsQuery.data]);

  const maxLatency = Math.max(...locationBars.map((l) => l.avgMs), 1);

  return (
    <StitchPageShell
      title="Synthetic Monitoring"
      subtitle="HTTP and browser checks with geo latency view"
      actions={
        <DataExportMenu
          getData={() => {
            if (selected && runsQuery.data?.length) {
              return runsQuery.data.map((r) => ({ monitorId: selected, ...r }));
            }
            return monitorsQuery.data ?? [];
          }}
          filenamePrefix={selected ? `synthetic-runs-${selected}` : 'synthetic-monitors'}
        />
      }
    >
      {monitorsQuery.isLoading && <LoadingState />}
      {(monitorsQuery.data ?? []).map((m) => (
        <Card key={m.id} title={m.name}>
          <p>{m.url} · every {m.interval}</p>
          <Badge variant={m.lastStatus === 'OK' ? 'healthy' : 'critical'}>{m.lastStatus}</Badge>
          <button type="button" className="pill" onClick={() => setSelected(m.id)}>View runs & map</button>
        </Card>
      ))}

      {runsQuery.data && locationBars.length > 0 && (
        <Card title="Geo / latency map" style={{ marginTop: 24 }}>
          <p className="muted">Average latency by probe location (placeholder map)</p>
          <div className="synthetic-geo-map" role="img" aria-label="Latency by location">
            {locationBars.map((loc) => (
              <div key={loc.location} className="synthetic-geo-row">
                <span className="synthetic-geo-label">{loc.location}</span>
                <div className="synthetic-geo-bar-track">
                  <div
                    className="synthetic-geo-bar"
                    style={{ width: `${(loc.avgMs / maxLatency) * 100}%` }}
                    title={`${loc.avgMs}ms avg · ${loc.count} runs`}
                  />
                </div>
                <span className="muted">{loc.avgMs}ms</span>
              </div>
            ))}
          </div>
        </Card>
      )}

      {runsQuery.data && (
        <Card title="Run history" style={{ marginTop: 16 }}>
          <table className="data-table">
            <thead><tr><th>Status</th><th>Location</th><th>Latency</th><th>Time</th></tr></thead>
            <tbody>
              {runsQuery.data.map((r) => (
                <tr key={r.id}>
                  <td>{r.status}</td>
                  <td>{r.location}</td>
                  <td>{r.latencyMs}ms</td>
                  <td>{new Date(r.ranAt).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}
    </StitchPageShell>
  );
}
