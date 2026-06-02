import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate, useSearch } from '@tanstack/react-router';
import { fetchTrace, searchTraces } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { TraceWaterfall } from '../features/traces/TraceWaterfall';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';

export default function TraceCompare() {
  const navigate = useNavigate();
  const search = useSearch({ strict: false }) as { a?: string; b?: string };
  const [pickerA, setPickerA] = useState(search.a ?? '');
  const [pickerB, setPickerB] = useState(search.b ?? '');
  const traceA = search.a ?? '';
  const traceB = search.b ?? '';

  const recentQuery = useQuery({
    queryKey: ['trace-compare-recent'],
    queryFn: () => searchTraces({ limit: 20 }),
  });

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

  const applyCompare = () => {
    void navigate({ to: '/traces/compare', search: { a: pickerA.trim(), b: pickerB.trim() } });
  };

  return (
    <StitchPageShell
      title="Compare traces"
      subtitle="Side-by-side waterfall comparison"
      actions={<Link to="/traces">← Trace explorer</Link>}
    >
      <div className="ui-card trace-compare-picker" style={{ marginBottom: 24 }}>
        <div className="form-stack" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr auto', gap: 12, alignItems: 'end' }}>
          <Input label="Trace A" placeholder="trace id" value={pickerA} onChange={(e) => setPickerA(e.target.value)} />
          <Input label="Trace B" placeholder="trace id" value={pickerB} onChange={(e) => setPickerB(e.target.value)} />
          <Button variant="primary" disabled={!pickerA.trim() || !pickerB.trim()} onClick={applyCompare}>
            Compare
          </Button>
        </div>
        {recentQuery.data && recentQuery.data.length > 0 && (
          <div style={{ marginTop: 16 }}>
            <span className="muted">Recent traces — click to fill picker</span>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginTop: 8 }}>
              {recentQuery.data.slice(0, 10).map((t) => (
                <button
                  key={t.traceId}
                  type="button"
                  className="pill"
                  onClick={() => {
                    if (!pickerA) setPickerA(t.traceId);
                    else setPickerB(t.traceId);
                  }}
                >
                  {t.service} · {t.durationMs}ms
                </button>
              ))}
            </div>
          </div>
        )}
      </div>

      {(!traceA || !traceB) && (
        <p className="muted">Select two traces above or from recent results.</p>
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
            <TraceWaterfall spans={queryA.data.spans} traceTotalMs={queryA.data.totalMs} traceId={traceA} />
          </section>
          <section>
            <div style={{ display: 'flex', gap: 8, marginBottom: 12, alignItems: 'center' }}>
              <h3 style={{ margin: 0 }}>{traceB}</h3>
              <Badge variant={queryB.data.status === 'ERROR' ? 'critical' : 'healthy'}>{queryB.data.status}</Badge>
              <span className="muted">{queryB.data.totalMs}ms</span>
            </div>
            <TraceWaterfall spans={queryB.data.spans} traceTotalMs={queryB.data.totalMs} traceId={traceB} />
          </section>
        </div>
      )}
    </StitchPageShell>
  );
}
