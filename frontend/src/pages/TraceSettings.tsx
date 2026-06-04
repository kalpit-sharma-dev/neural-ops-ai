import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { fetchTraceRetention, updateTraceRetention } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';

export default function TraceSettings() {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({ queryKey: ['trace-retention'], queryFn: fetchTraceRetention });

  const saveMut = useMutation({
    mutationFn: updateTraceRetention,
    onSuccess: () => {
      toast.success('Retention policy saved');
      void queryClient.invalidateQueries({ queryKey: ['trace-retention'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  if (isLoading || !data) return <LoadingState />;

  return (
    <StitchPageShell
      title="Trace sampling & retention"
      subtitle="Control head/tail sampling and ClickHouse retention"
    >
      <Card title="Policy">
        <div className="form-stack">
          <label>
            Retention (days)
            <input
              type="number"
              min={1}
              max={365}
              defaultValue={data.retentionDays}
              id="retention-days"
            />
          </label>
          <label>
            Head sample rate (0–1)
            <input type="number" min={0} max={1} step={0.01} defaultValue={data.headSampleRate} id="head-rate" />
          </label>
          <label>
            Tail sample rate for errors (0–1)
            <input type="number" min={0} max={1} step={0.01} defaultValue={data.tailSampleRate} id="tail-rate" />
          </label>
          <Button
            variant="primary"
            onClick={() => {
              const retentionDays = Number((document.getElementById('retention-days') as HTMLInputElement).value);
              const headSampleRate = Number((document.getElementById('head-rate') as HTMLInputElement).value);
              const tailSampleRate = Number((document.getElementById('tail-rate') as HTMLInputElement).value);
              saveMut.mutate({ retentionDays, headSampleRate, tailSampleRate });
            }}
          >
            Save policy
          </Button>
        </div>
        <p className="muted" style={{ marginTop: 16 }}>
          Ingestion applies head sampling per trace ID; error traces use tail sample rate. Retention TTL is applied on ClickHouse span store.
        </p>
      </Card>
    </StitchPageShell>
  );
}
