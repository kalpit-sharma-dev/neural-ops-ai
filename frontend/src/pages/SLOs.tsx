import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import toast from 'react-hot-toast';
import { Link } from '@tanstack/react-router';
import { createSLO, fetchSLOs, updateSLOBurnAlert } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { LoadingState } from '../components/ui/PageStates';
import { DataExportMenu } from '../components/ui/DataExportMenu';
import { StitchPageShell } from '../components/stitch';

export default function SLOs() {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({ queryKey: ['slos'], queryFn: fetchSLOs });

  const [form, setForm] = useState({
    name: '',
    service: '',
    sliQuery: 'sum(rate(http_requests_total{status!~"5.."}[5m])) / sum(rate(http_requests_total[5m]))',
    target: 99.9,
    windowDays: 30,
  });

  const createMut = useMutation({
    mutationFn: () => createSLO(form),
    onSuccess: () => {
      toast.success('SLO created');
      void queryClient.invalidateQueries({ queryKey: ['slos'] });
      setForm({ ...form, name: '', service: '' });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const burnMut = useMutation({
    mutationFn: ({ id, enabled, threshold }: { id: string; enabled: boolean; threshold: number }) =>
      updateSLOBurnAlert(id, { enabled, threshold }),
    onSuccess: () => {
      toast.success('Burn alert updated');
      void queryClient.invalidateQueries({ queryKey: ['slos'] });
    },
  });

  return (
    <StitchPageShell
      title="SLOs"
      subtitle="Service level objectives and error budgets"
      actions={
        <DataExportMenu getData={() => data ?? []} filenamePrefix="slos" disabled={!data?.length} />
      }
    >
      <div className="dashboard-row-2" style={{ marginBottom: 24 }}>
        <Card title="Create SLO">
          <div className="form-stack">
            <Input label="Name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
            <Input label="Service" value={form.service} onChange={(e) => setForm({ ...form, service: e.target.value })} />
            <Input
              label="SLI query"
              value={form.sliQuery}
              onChange={(e) => setForm({ ...form, sliQuery: e.target.value })}
            />
            <Input
              label="Target %"
              type="number"
              value={String(form.target)}
              onChange={(e) => setForm({ ...form, target: Number(e.target.value) })}
            />
            <Button variant="primary" disabled={!form.name || !form.service} onClick={() => createMut.mutate()}>
              Create SLO
            </Button>
          </div>
        </Card>
      </div>
      {isLoading && <LoadingState />}
      {(data ?? []).map((slo) => (
        <Card key={slo.id} title={slo.name}>
          <div style={{ display: 'flex', gap: 16, alignItems: 'center', flexWrap: 'wrap' }}>
            <Badge variant={slo.status === 'OK' ? 'healthy' : 'critical'}>{slo.status}</Badge>
            <span>{slo.service}</span>
            <span>Target {slo.target}%</span>
            <span>Error budget {slo.errorBudget.toFixed(2)}%</span>
            <span>Burn {slo.burnRate.toFixed(2)}</span>
            <Button
              variant="secondary"
              size="sm"
              onClick={() =>
                burnMut.mutate({
                  id: slo.id,
                  enabled: true,
                  threshold: 2.0,
                })
              }
            >
              Enable burn alert
            </Button>
            <Link to="/alerts">View alerts →</Link>
          </div>
        </Card>
      ))}
    </StitchPageShell>
  );
}
