import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import toast from 'react-hot-toast';
import { Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { createNotebook, executeNotebook, fetchNotebooks, queryPromQL } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';
import { chartCartesianDefaults } from '../lib/chartTheme';
import { useTimeBounds } from '../hooks/useTimeBounds';

type CellType = 'markdown' | 'log' | 'promql';

interface CellOutput {
  cellId: string;
  type: string;
  output: string;
  durationMs: number;
  chartData?: { time: string; value: number }[];
}

export default function Notebooks() {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({ queryKey: ['notebooks'], queryFn: fetchNotebooks });
  const [name, setName] = useState('');
  const [results, setResults] = useState<Record<string, CellOutput[]>>({});
  const [runningCell, setRunningCell] = useState<string | null>(null);
  const { resolve: resolveRange } = useTimeBounds();

  const createMut = useMutation({
    mutationFn: () =>
      createNotebook({
        name,
        cells: [
          { id: 'c1', type: 'markdown', content: '# Analysis notebook' },
          { id: 'c2', type: 'log', content: 'service:payment-service status:ERROR' },
          { id: 'c3', type: 'promql', content: 'rate(http_requests_total[5m])' },
        ],
      }),
    onSuccess: () => {
      toast.success('Notebook created');
      setName('');
      void queryClient.invalidateQueries({ queryKey: ['notebooks'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const runMut = useMutation({
    mutationFn: executeNotebook,
    onSuccess: (res, notebookId) => {
      setResults((prev) => ({
        ...prev,
        [notebookId]: res.map((r) => ({ ...r, chartData: undefined })),
      }));
      toast.success('Notebook executed');
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const runCell = async (notebookId: string, cellId: string, type: CellType, content: string) => {
    setRunningCell(cellId);
    try {
      if (type === 'promql') {
        const { startIso, endIso } = resolveRange();
        const series = await queryPromQL(content, { start: startIso, end: endIso });
        const chartData = (series.points ?? []).map((p) => ({
          time: new Date(p.timestamp).toLocaleTimeString(),
          value: p.value,
        }));
        setResults((prev) => {
          const list = prev[notebookId] ?? [];
          const filtered = list.filter((r) => r.cellId !== cellId);
          return {
            ...prev,
            [notebookId]: [
              ...filtered,
              {
                cellId,
                type,
                output: `${series.points.length} points`,
                durationMs: 0,
                chartData,
              },
            ],
          };
        });
        toast.success('PromQL cell ran');
      } else {
        const res = await executeNotebook(notebookId);
        const match = res.find((r) => r.cellId === cellId);
        setResults((prev) => {
          const list = (prev[notebookId] ?? []).filter((r) => r.cellId !== cellId);
          if (match) list.push({ ...match, chartData: undefined });
          return { ...prev, [notebookId]: list };
        });
      }
    } catch (e) {
      toast.error(getApiErrorMessage(e));
    } finally {
      setRunningCell(null);
    }
  };

  const chartDefaults = chartCartesianDefaults();

  return (
    <StitchPageShell title="Notebooks" subtitle="Saved analysis notebooks">
      <Card title="New notebook" style={{ marginBottom: 24 }}>
        <div className="form-stack">
          <Input label="Notebook name" placeholder="Weekly error review" value={name} onChange={(e) => setName(e.target.value)} />
          <Button variant="primary" disabled={!name} onClick={() => createMut.mutate()}>Create</Button>
        </div>
      </Card>
      {isLoading && <LoadingState />}
      {(data ?? []).map((n) => (
        <Card
          key={n.id}
          title={n.name}
          action={
            <Button variant="secondary" size="sm" disabled={runMut.isPending} onClick={() => runMut.mutate(n.id)}>
              Run all cells
            </Button>
          }
        >
          <p className="muted">{n.cells.length} cells · updated {new Date(n.updatedAt).toLocaleDateString()}</p>
          {n.cells.map((c) => {
            const cellType = (['markdown', 'log', 'promql'].includes(c.type) ? c.type : 'log') as CellType;
            const out = results[n.id]?.find((r) => r.cellId === c.id);
            return (
              <div key={c.id} className="notebook-cell">
                <div className="notebook-cell__header">
                  <BadgeType type={cellType} />
                  <Button
                    variant="secondary"
                    size="sm"
                    disabled={runningCell === c.id}
                    onClick={() => void runCell(n.id, c.id, cellType, c.content)}
                  >
                    Run
                  </Button>
                </div>
                {cellType === 'markdown' ? (
                  <pre className="code-block">{c.content}</pre>
                ) : (
                  <pre className="code-block" style={{ fontSize: 12 }}>{c.content}</pre>
                )}
                {out && (
                  <>
                    <pre className="code-block" style={{ fontSize: 12, background: 'var(--surface-2)' }}>
                      {out.output}
                      {out.durationMs > 0 && `\n(${out.durationMs}ms)`}
                    </pre>
                    {out.chartData && out.chartData.length > 0 && (
                      <ResponsiveContainer width="100%" height={180}>
                        <LineChart data={out.chartData}>
                          <XAxis dataKey="time" {...chartDefaults.axis} />
                          <YAxis {...chartDefaults.axis} />
                          <Tooltip contentStyle={chartDefaults.tooltipStyle} />
                          <Line type="monotone" dataKey="value" stroke="var(--accent-primary)" dot={false} />
                        </LineChart>
                      </ResponsiveContainer>
                    )}
                  </>
                )}
              </div>
            );
          })}
        </Card>
      ))}
    </StitchPageShell>
  );
}

function BadgeType({ type }: { type: CellType }) {
  return <span className={`pill pill--${type}`}>{type}</span>;
}
