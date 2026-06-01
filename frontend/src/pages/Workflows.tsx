import { Link } from '@tanstack/react-router';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import toast from 'react-hot-toast';
import { createWorkflow, fetchWorkflows } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function Workflows() {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({ queryKey: ['workflows'], queryFn: fetchWorkflows });
  const [name, setName] = useState('');
  const [trigger, setTrigger] = useState('incident.p1');

  const createMut = useMutation({
    mutationFn: () =>
      createWorkflow({
        name,
        trigger,
        enabled: true,
        steps: ['Notify Slack', 'Create Jira ticket', 'Page on-call'],
      }),
    onSuccess: () => {
      toast.success('Workflow created');
      setName('');
      void queryClient.invalidateQueries({ queryKey: ['workflows'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  return (
    <div>
      <PageHeader title="Workflows" subtitle="Automation and remediation" actions={<Link to="/workflows/editor">Visual editor →</Link>} />
      <Card title="Create workflow" style={{ marginBottom: 24 }}>
        <div className="form-stack">
          <input placeholder="Name" value={name} onChange={(e) => setName(e.target.value)} />
          <input placeholder="Trigger (e.g. incident.p1)" value={trigger} onChange={(e) => setTrigger(e.target.value)} />
          <Button variant="primary" disabled={!name} onClick={() => createMut.mutate()}>Create</Button>
        </div>
      </Card>
      {isLoading && <LoadingState />}
      {(data ?? []).map((w) => (
        <Card key={w.id} title={w.name}>
          <p>Trigger: {w.trigger}</p>
          <Badge variant={w.enabled ? 'healthy' : 'info'}>{w.enabled ? 'Enabled' : 'Disabled'}</Badge>
          <ol>{w.steps.map((s) => <li key={s}>{s}</li>)}</ol>
        </Card>
      ))}
    </div>
  );
}
