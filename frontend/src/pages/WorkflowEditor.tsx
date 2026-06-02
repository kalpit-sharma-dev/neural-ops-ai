import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate, useSearch } from '@tanstack/react-router';
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
import { ArrowDown, ArrowUp, Trash2 } from 'lucide-react';
import toast from 'react-hot-toast';
import {
  createWorkflow,
  deleteWorkflow,
  fetchWorkflows,
  updateWorkflow,
  type WorkflowGraph,
} from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
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

/**
 * Resolves the execution order of step labels by walking the canvas edges
 * (Kahn topological sort), so the connections the user draws are what gets
 * persisted. Disconnected/cyclic nodes fall back to their array order.
 */
function orderedStepLabels(nodes: Node[], edges: Edge[]): string[] {
  const labelOf = (n: Node) => String(n.data.label ?? 'Step');
  const byId = new Map(nodes.map((n) => [n.id, n]));
  const indegree = new Map(nodes.map((n) => [n.id, 0]));
  const adjacency = new Map<string, string[]>();
  edges.forEach((e) => {
    if (!byId.has(e.source) || !byId.has(e.target)) return;
    adjacency.set(e.source, [...(adjacency.get(e.source) ?? []), e.target]);
    indegree.set(e.target, (indegree.get(e.target) ?? 0) + 1);
  });
  const queue = nodes.map((n) => n.id).filter((id) => (indegree.get(id) ?? 0) === 0);
  const visited = new Set<string>();
  const result: string[] = [];
  while (queue.length) {
    const id = queue.shift()!;
    if (visited.has(id)) continue;
    visited.add(id);
    const node = byId.get(id);
    if (node) result.push(labelOf(node));
    (adjacency.get(id) ?? []).forEach((target) => {
      const remaining = (indegree.get(target) ?? 0) - 1;
      indegree.set(target, remaining);
      if (remaining <= 0 && !visited.has(target)) queue.push(target);
    });
  }
  nodes.forEach((n) => {
    if (!visited.has(n.id)) result.push(labelOf(n));
  });
  return result;
}

function inferStepType(label: string): string {
  const l = label.toLowerCase();
  if (l.includes('slack') || l.includes('notify')) return 'slack';
  if (l.includes('jira') || l.includes('ticket')) return 'jira';
  if (l.includes('page') || l.includes('oncall') || l.includes('on-call')) return 'pagerduty';
  return 'action';
}

const EDGE_CONDITIONS = [
  { value: 'success', label: 'On success' },
  { value: 'failure', label: 'On failure' },
  { value: 'always', label: 'Always' },
] as const;

function edgeCondition(edge: Edge): string {
  return String(edge.data?.condition ?? 'success');
}

/** Non-default conditions are surfaced as an edge label to visualize branches. */
function conditionLabel(cond: string): string | undefined {
  if (cond === 'success' || cond === '') return undefined;
  return EDGE_CONDITIONS.find((c) => c.value === cond)?.label ?? cond;
}

/** Serializes the live canvas into the persisted branching graph. */
function canvasToGraph(nodes: Node[], edges: Edge[]): WorkflowGraph {
  return {
    nodes: nodes.map((n) => ({
      id: n.id,
      type: String(n.data.stepType ?? inferStepType(String(n.data.label ?? ''))),
      label: String(n.data.label ?? 'Step'),
      x: n.position.x,
      y: n.position.y,
    })),
    edges: edges.map((e) => ({
      id: e.id,
      source: e.source,
      target: e.target,
      condition: edgeCondition(e),
    })),
  };
}

/** Rebuilds the canvas (positions + branching edges) from a stored graph. */
function graphToCanvas(graph: WorkflowGraph): { nodes: Node[]; edges: Edge[] } {
  return {
    nodes: graph.nodes.map((n) => ({
      id: n.id,
      position: { x: n.x, y: n.y },
      data: { label: n.label, stepType: n.type },
      type: 'default',
    })),
    edges: graph.edges.map((e) => {
      const cond = String(e.condition ?? 'success');
      return {
        id: e.id,
        source: e.source,
        target: e.target,
        data: { condition: cond },
        label: conditionLabel(cond),
      };
    }),
  };
}

export default function WorkflowEditor() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const { id: editingId } = useSearch({ strict: false }) as { id?: string };
  const { data: workflows, isLoading } = useQuery({ queryKey: ['workflows'], queryFn: fetchWorkflows });
  const editing = (workflows ?? []).find((w) => w.id === editingId);

  const [name, setName] = useState('');
  const [trigger, setTrigger] = useState('incident.p1');
  const [enabled, setEnabled] = useState(true);

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
  const [selectedEdgeId, setSelectedEdgeId] = useState<string | null>(null);

  // Seed the canvas from an existing workflow when editing (once per id).
  const seededId = useRef<string | null>(null);
  useEffect(() => {
    if (!editing || seededId.current === editing.id) return;
    seededId.current = editing.id;
    setName(editing.name);
    setTrigger(editing.trigger);
    setEnabled(editing.enabled);
    if (editing.graph && editing.graph.nodes.length > 0) {
      // Restore the authored branching graph exactly (positions + edges).
      const { nodes: gNodes, edges: gEdges } = graphToCanvas(editing.graph);
      setNodes(gNodes);
      setEdges(gEdges);
      return;
    }
    // Legacy workflows without a stored graph: lay steps out as a linear chain.
    const stepNodes = stepsToNodes(
      editing.steps.map((label, i) => ({ id: `s${i + 1}`, type: inferStepType(label), label })),
    );
    setNodes(stepNodes);
    setEdges(
      stepNodes.slice(1).map((node, i) => ({
        id: `e${i + 1}-${i + 2}`,
        source: stepNodes[i].id,
        target: node.id,
      })),
    );
  }, [editing, setNodes, setEdges]);

  // When leaving edit mode (id cleared), reset to a fresh canvas.
  useEffect(() => {
    if (editingId || seededId.current === null) return;
    seededId.current = null;
    setName('');
    setTrigger('incident.p1');
    setEnabled(true);
    const fresh = stepsToNodes([
      { id: 's1', type: 'slack', label: 'Notify Slack' },
      { id: 's2', type: 'jira', label: 'Create Jira ticket' },
    ]);
    setNodes(fresh);
    setEdges([{ id: 'e1-2', source: 's1', target: 's2' }]);
  }, [editingId, setNodes, setEdges]);

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge({ ...params, data: { condition: 'success' } }, eds)),
    [setEdges],
  );

  const onEdgeClick = useCallback((_: unknown, edge: Edge) => setSelectedEdgeId(edge.id), []);

  const setEdgeCondition = (edgeId: string, cond: string) => {
    setEdges((eds) =>
      eds.map((e) =>
        e.id === edgeId ? { ...e, data: { ...e.data, condition: cond }, label: conditionLabel(cond) } : e,
      ),
    );
  };

  const selectedEdge = edges.find((e) => e.id === selectedEdgeId) ?? null;
  const nodeLabelById = (id: string) => String(nodes.find((n) => n.id === id)?.data.label ?? id);

  const addStep = (type: string, label: string) => {
    const id = `s${Date.now()}`;
    setNodes((nds) => [...nds, { id, position: { x: 80, y: nds.length * 100 + 40 }, data: { label, stepType: type } }]);
  };

  const moveStep = (index: number, dir: -1 | 1) => {
    const next = index + dir;
    if (next < 0 || next >= nodes.length) return;
    setNodes((nds) => {
      const copy = [...nds];
      const [item] = copy.splice(index, 1);
      copy.splice(next, 0, item);
      return copy.map((n, i) => ({ ...n, position: { ...n.position, y: i * 100 + 40 } }));
    });
  };

  const resetForm = () => {
    setName('');
    setTrigger('incident.p1');
    setEnabled(true);
    seededId.current = null;
  };

  const saveMut = useMutation({
    mutationFn: () => {
      const payload = {
        name,
        trigger,
        enabled,
        steps: orderedStepLabels(nodes, edges),
        graph: canvasToGraph(nodes, edges),
      };
      return editing ? updateWorkflow(editing.id, payload) : createWorkflow(payload);
    },
    onSuccess: () => {
      toast.success(editing ? 'Workflow updated' : 'Workflow saved');
      void queryClient.invalidateQueries({ queryKey: ['workflows'] });
      if (!editing) resetForm();
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const deleteMut = useMutation({
    mutationFn: () => deleteWorkflow(editing!.id),
    onSuccess: () => {
      toast.success('Workflow deleted');
      void queryClient.invalidateQueries({ queryKey: ['workflows'] });
      void navigate({ to: '/workflows' });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const stepList = nodesToSteps(nodes);

  return (
    <div>
      <PageHeader
        title={editing ? `Edit workflow` : 'Workflow editor'}
        subtitle={editing ? editing.name : 'Visual automation builder'}
        actions={
          editing ? (
            <Button variant="ghost" size="sm" onClick={() => navigate({ to: '/workflows/editor', search: { id: undefined } })}>
              + New workflow
            </Button>
          ) : undefined
        }
      />
      <div className="dashboard-row-2">
        <Card title="Canvas">
          <div style={{ height: 420 }}>
            <ReactFlow
              nodes={nodes}
              edges={edges}
              onNodesChange={onNodesChange}
              onEdgesChange={onEdgesChange}
              onConnect={onConnect}
              onEdgeClick={onEdgeClick}
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
          {selectedEdge && (
            <div className="workflow-edge-editor">
              <div className="workflow-edge-editor__title">
                Connection: <strong>{nodeLabelById(selectedEdge.source)}</strong> →{' '}
                <strong>{nodeLabelById(selectedEdge.target)}</strong>
              </div>
              <label className="ui-field">
                <span className="ui-field__label">Run target when</span>
                <select
                  className="ui-field__control"
                  value={edgeCondition(selectedEdge)}
                  onChange={(e) => setEdgeCondition(selectedEdge.id, e.target.value)}
                >
                  {EDGE_CONDITIONS.map((c) => (
                    <option key={c.value} value={c.value}>
                      {c.label}
                    </option>
                  ))}
                </select>
              </label>
              <p className="muted workflow-edge-editor__hint">
                Connect one step to several to fan out into parallel branches; a step with multiple
                incoming connections waits for all of them (join).
              </p>
            </div>
          )}
        </Card>
        <div>
          <Card title="Step list">
            <ol className="workflow-step-list">
              {stepList.map((step, index) => (
                <li key={step.id} className="workflow-step-list__item">
                  <span>{step.label}</span>
                  <div className="workflow-step-list__actions">
                    <Button variant="ghost" size="sm" disabled={index === 0} onClick={() => moveStep(index, -1)} aria-label="Move up">
                      <ArrowUp size={14} />
                    </Button>
                    <Button variant="ghost" size="sm" disabled={index === stepList.length - 1} onClick={() => moveStep(index, 1)} aria-label="Move down">
                      <ArrowDown size={14} />
                    </Button>
                  </div>
                </li>
              ))}
            </ol>
          </Card>
          <Card title={editing ? 'Update workflow' : 'Save workflow'} style={{ marginTop: 16 }}>
            <div className="form-stack">
              <Input label="Name" placeholder="P1 incident response" value={name} onChange={(e) => setName(e.target.value)} />
              <Input label="Trigger" placeholder="incident.p1" value={trigger} onChange={(e) => setTrigger(e.target.value)} />
              <label className="workflow-enabled-toggle">
                <input type="checkbox" checked={enabled} onChange={(e) => setEnabled(e.target.checked)} />
                <span>Enabled</span>
              </label>
              <div style={{ display: 'flex', gap: 8 }}>
                <Button variant="primary" disabled={!name || saveMut.isPending} onClick={() => saveMut.mutate()}>
                  {editing ? 'Update' : 'Save'}
                </Button>
                {editing && (
                  <Button
                    variant="ghost"
                    disabled={deleteMut.isPending}
                    onClick={() => {
                      if (window.confirm(`Delete workflow "${editing.name}"? This cannot be undone.`)) {
                        deleteMut.mutate();
                      }
                    }}
                  >
                    <Trash2 size={14} /> Delete
                  </Button>
                )}
              </div>
            </div>
            {isLoading && <LoadingState />}
            {!editing && (
              <>
                <h4 style={{ marginTop: 24 }}>Existing</h4>
                {(workflows ?? []).map((w) => (
                  <p key={w.id} className="muted">{w.name} · {w.trigger}</p>
                ))}
              </>
            )}
          </Card>
        </div>
      </div>
    </div>
  );
}
