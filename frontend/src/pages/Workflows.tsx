import { Link } from '@tanstack/react-router';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useMemo, useState } from 'react';
import toast from 'react-hot-toast';
import { Pencil, Play, Plus, Power, Trash2, Workflow as WorkflowIcon, Zap } from 'lucide-react';
import {
  createWorkflow,
  deleteWorkflow,
  fetchWorkflows,
  triggerWorkflows,
  updateWorkflow,
  type Workflow,
} from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { EmptyState } from '../components/ui/EmptyState';
import { Input } from '../components/ui/Input';
import { MetricCard } from '../components/ui/MetricCard';
import { LoadingState } from '../components/ui/PageStates';
import { DataExportMenu } from '../components/ui/DataExportMenu';
import { KpiRow, StitchPageShell } from '../components/stitch';

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

  const runMut = useMutation({
    mutationFn: (workflowTrigger: string) =>
      triggerWorkflows({ trigger: workflowTrigger, context: { source: 'manual-test' } }),
    onSuccess: () => toast.success('Trigger fired — matching workflows executed'),
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['workflows'] });

  const toggleMut = useMutation({
    mutationFn: (w: Workflow) =>
      updateWorkflow(w.id, { name: w.name, trigger: w.trigger, enabled: !w.enabled, steps: w.steps }),
    onSuccess: (_d, w) => {
      toast.success(w.enabled ? 'Workflow disabled' : 'Workflow enabled');
      void invalidate();
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const deleteMut = useMutation({
    mutationFn: (id: string) => deleteWorkflow(id),
    onSuccess: () => {
      toast.success('Workflow deleted');
      void invalidate();
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const mutating = toggleMut.isPending || deleteMut.isPending;

  const workflows = data ?? [];
  const stats = useMemo(
    () => ({
      total: workflows.length,
      enabled: workflows.filter((w) => w.enabled).length,
      triggers: new Set(workflows.map((w) => w.trigger)).size,
    }),
    [workflows],
  );

  return (
    <StitchPageShell
      title="Workflows"
      subtitle="Automation and remediation"
      actions={
        <>
          <DataExportMenu getData={() => workflows} filenamePrefix="workflows" disabled={!workflows.length} />
          <Link to="/workflows/editor" search={{ id: undefined }}>Visual editor →</Link>
        </>
      }
    >
      {workflows.length > 0 && (
        <KpiRow columns={3}>
          <MetricCard label="Workflows" value={stats.total} />
          <MetricCard label="Enabled" value={stats.enabled} accent="success" />
          <MetricCard label="Distinct triggers" value={stats.triggers} />
        </KpiRow>
      )}

      <Card title="Create workflow">
        <div style={{ display: 'flex', gap: 12, alignItems: 'flex-end', flexWrap: 'wrap' }}>
          <Input
            label="Name"
            value={name}
            data-testid="workflows-create-name"
            onChange={(e) => setName(e.target.value)}
          />
          <Input
            label="Trigger"
            placeholder="incident.p1"
            value={trigger}
            data-testid="workflows-create-trigger"
            onChange={(e) => setTrigger(e.target.value)}
          />
          <Button
            variant="primary"
            disabled={!name}
            data-testid="workflows-create-btn"
            onClick={() => createMut.mutate()}
          >
            <Plus size={15} /> Create
          </Button>
        </div>
      </Card>

      {isLoading && <LoadingState />}
      {!isLoading && workflows.length === 0 && (
        <EmptyState
          title="No workflows yet"
          description="Create an automation to notify, ticket, and page on-call when a trigger fires."
          icon={<WorkflowIcon size={32} />}
        />
      )}

      {workflows.length > 0 && (
        <div className="workflow-grid">
          {workflows.map((w) => (
            <Card key={w.id} className="workflow-card">
              <div className="workflow-card__head">
                <span className="workflow-card__title">
                  <WorkflowIcon size={16} className="workflow-card__icon" aria-hidden />
                  {w.name}
                </span>
                <Badge variant={w.enabled ? 'healthy' : 'info'}>{w.enabled ? 'Enabled' : 'Disabled'}</Badge>
              </div>

              <span className="workflow-trigger-chip">
                <Zap size={12} aria-hidden /> {w.trigger}
              </span>

              <ol className="workflow-pipeline">
                {w.steps.map((step, i) => (
                  <li key={`${step}-${i}`} className="workflow-step">
                    <span className="workflow-step__index">{i + 1}</span>
                    <span className="workflow-step__label">{step}</span>
                  </li>
                ))}
              </ol>

              <div className="workflow-card__actions">
                <Button
                  variant="secondary"
                  size="sm"
                  disabled={!w.enabled || runMut.isPending}
                  onClick={() => runMut.mutate(w.trigger)}
                  data-testid={`workflows-test-run-${w.id}`}
                >
                  <Play size={13} /> Test run
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  disabled={mutating}
                  onClick={() => toggleMut.mutate(w)}
                  aria-label={w.enabled ? 'Disable workflow' : 'Enable workflow'}
                  data-testid={`workflows-toggle-${w.id}`}
                >
                  <Power size={13} /> {w.enabled ? 'Disable' : 'Enable'}
                </Button>
                <Link to="/workflows/editor" search={{ id: w.id }} className="pill" data-testid={`workflows-edit-${w.id}`}>
                  <Pencil size={12} aria-hidden /> Edit
                </Link>
                <Button
                  variant="ghost"
                  size="sm"
                  disabled={mutating}
                  onClick={() => {
                    if (window.confirm(`Delete workflow "${w.name}"? This cannot be undone.`)) {
                      deleteMut.mutate(w.id);
                    }
                  }}
                  aria-label="Delete workflow"
                  data-testid={`workflows-delete-${w.id}`}
                >
                  <Trash2 size={13} /> Delete
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}
    </StitchPageShell>
  );
}
