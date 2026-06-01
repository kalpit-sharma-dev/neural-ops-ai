import { useCallback, useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  ReactFlow,
  Background,
  Controls,
  addEdge,
  useEdgesState,
  useNodesState,
  type Connection,
  type Edge,
  type Node,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import toast from 'react-hot-toast';
import { createWorkflow, fetchWorkflows } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

const STEP_TYPES = [
  { type: 'slack', label: 'Notify Slack' },
  { type: 'jira', label: 'Create Jira ticket' },
  { type: 'pagerduty', label: 'Page on-call' },
];

function stepsToNodes(steps: { id: string; type: string; label: string }[]): Node[] {
  return steps.map((s, i) => ({
    id: s.id,
    position: { x: 80, y: i * 100 + 40 },
    data: { label: s.label, stepType: s.type },
    type: 'default',
  }));
}

function nodesToSteps(nodes: Node[]): { id: string; type: string; label: string }[] {
  return nodes.map((n) => ({
    id: n.id,
    type: String(n.data.stepType ?? 'action'),
    label: String(n.data.label ?? 'Step'),
  }));
}

export default function WorkflowEditor() {
  const queryClient = useQueryClient();
  const { data: workflows, isLoading } = useQuery({ queryKey: ['workflows'], queryFn: fetchWorkflows });
  const [name, setName] = useState('');
  const [trigger, setTrigger] = useState('incident.p1');

  const initialNodes = useMemo(
    () =>
      stepsToNodes([
        { id: 's1', type: 'slack', label: 'Notify Slack' },
        { id: 's2', type: 'jira', label: 'Create Jira ticket' },
      ]),
    [],
  );
  const initialEdges: Edge[] = [{ id: 'e1-2', source: 's1', target: 's2' }];

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges],
  );

  const addStep = (type: string, label: string) => {
    const id = `s${Date.now()}`;
    setNodes((nds) => [...nds, { id, position: { x: 80, y: nds.length * 100 + 40 }, data: { label, stepType: type } }]);
  };

  const createMut = useMutation({
    mutationFn: () =>
      createWorkflow({
        name,
        trigger,
        enabled: true,
        steps: nodesToSteps(nodes).map((s) => s.label),
      }),
    onSuccess: () => {
      toast.success('Workflow saved');
      setName('');
      void queryClient.invalidateQueries({ queryKey: ['workflows'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  return (
    <div>
      <PageHeader title="Workflow editor" subtitle="Visual automation builder" />
      <div className="dashboard-row-2">
        <Card title="Canvas">
          <div style={{ height: 420 }}>
            <ReactFlow
              nodes={nodes}
              edges={edges}
              onNodesChange={onNodesChange}
              onEdgesChange={onEdgesChange}
              onConnect={onConnect}
              fitView
            >
              <Background />
              <Controls />
            </ReactFlow>
          </div>
          <div style={{ display: 'flex', gap: 8, marginTop: 12, flexWrap: 'wrap' }}>
            {STEP_TYPES.map((s) => (
              <Button key={s.type} variant="secondary" size="sm" onClick={() => addStep(s.type, s.label)}>
                + {s.label}
              </Button>
            ))}
          </div>
        </Card>
        <Card title="Save workflow">
          <div className="form-stack">
            <input placeholder="Name" value={name} onChange={(e) => setName(e.target.value)} />
            <input placeholder="Trigger" value={trigger} onChange={(e) => setTrigger(e.target.value)} />
            <Button variant="primary" disabled={!name} onClick={() => createMut.mutate()}>Save</Button>
          </div>
          {isLoading && <LoadingState />}
          <h4 style={{ marginTop: 24 }}>Existing</h4>
          {(workflows ?? []).map((w) => (
            <p key={w.id} className="muted">{w.name} · {w.trigger}</p>
          ))}
        </Card>
      </div>
    </div>
  );
}
