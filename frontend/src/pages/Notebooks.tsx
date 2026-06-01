import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import toast from 'react-hot-toast';
import { createNotebook, executeNotebook, fetchNotebooks } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function Notebooks() {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({ queryKey: ['notebooks'], queryFn: fetchNotebooks });
  const [name, setName] = useState('');
  const [results, setResults] = useState<Record<string, { cellId: string; type: string; output: string; durationMs: number }[]>>({});

  const createMut = useMutation({
    mutationFn: () =>
      createNotebook({
        name,
        cells: [
          { id: 'c1', type: 'markdown', content: '# Analysis notebook' },
          { id: 'c2', type: 'query', content: 'service:payment-service status:ERROR' },
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
      setResults((prev) => ({ ...prev, [notebookId]: res }));
      toast.success('Notebook executed');
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  return (
    <div>
      <PageHeader title="Notebooks" subtitle="Saved analysis notebooks" />
      <Card title="New notebook" style={{ marginBottom: 24 }}>
        <div className="form-stack">
          <input placeholder="Notebook name" value={name} onChange={(e) => setName(e.target.value)} />
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
            const out = results[n.id]?.find((r) => r.cellId === c.id);
            return (
              <div key={c.id} style={{ marginBottom: 12 }}>
                <pre className="code-block" style={{ fontSize: 12 }}>{c.content}</pre>
                {out && (
                  <pre className="code-block" style={{ fontSize: 12, background: 'var(--surface-2)' }}>
                    {out.output}
                    {'\n'}({out.durationMs}ms)
                  </pre>
                )}
              </div>
            );
          })}
        </Card>
      ))}
    </div>
  );
}
