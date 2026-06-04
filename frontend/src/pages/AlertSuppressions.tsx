import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { useState } from 'react';
import {
  createAlertSuppression,
  deleteAlertSuppression,
  fetchAlertSuppressions,
} from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { StitchPageShell } from '../components/stitch';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { LoadingState } from '../components/ui/PageStates';

export default function AlertSuppressions() {
  const qc = useQueryClient();
  const { data, isLoading } = useQuery({ queryKey: ['alert-suppressions'], queryFn: fetchAlertSuppressions });
  const [form, setForm] = useState({ servicePattern: '*', reason: '', duration: '1h' });

  const createMut = useMutation({
    mutationFn: () => createAlertSuppression(form),
    onSuccess: () => {
      toast.success('Suppression created');
      void qc.invalidateQueries({ queryKey: ['alert-suppressions'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const deleteMut = useMutation({
    mutationFn: deleteAlertSuppression,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alert-suppressions'] }),
  });

  if (isLoading) return <LoadingState />;

  return (
    <StitchPageShell
      title="Alert suppressions"
      subtitle="Maintenance windows and noise reduction rules"
    >
      <Card title="New suppression">
        <form
          className="form-stack"
          onSubmit={(e) => {
            e.preventDefault();
            createMut.mutate();
          }}
        >
          <Input
            label="Service pattern"
            value={form.servicePattern}
            onChange={(e) => setForm({ ...form, servicePattern: e.target.value })}
          />
          <Input label="Reason" value={form.reason} onChange={(e) => setForm({ ...form, reason: e.target.value })} />
          <Input label="Duration" value={form.duration} onChange={(e) => setForm({ ...form, duration: e.target.value })} />
          <Button type="submit" variant="primary">
            Create
          </Button>
        </form>
      </Card>
      <Card title="Active suppressions">
        {(data ?? []).map((s) => (
          <div key={s.id} className="list-row">
            <span>
              <strong>{s.servicePattern}</strong> — {s.reason}
            </span>
            <span className="muted">
              until {new Date(s.endsAt).toLocaleString()}
            </span>
            <Button size="sm" variant="ghost" onClick={() => deleteMut.mutate(s.id)}>
              Remove
            </Button>
          </div>
        ))}
      </Card>
    </StitchPageShell>
  );
}
