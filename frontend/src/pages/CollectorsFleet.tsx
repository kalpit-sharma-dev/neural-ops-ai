import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import {
  fetchCollectorFleet,
  createCollectorAgent,
  upgradeCollectorAgent,
  fetchCollectorPipelines,
  createCollectorPipeline,
  validateCollectorPipeline,
  type CollectorPipelineStage,
} from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { StitchPageShell } from '../components/stitch';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { Select } from '../components/ui/Select';
import { Button } from '../components/ui/Button';
import { Badge } from '../components/ui/Badge';

export default function CollectorsFleet() {
  const qc = useQueryClient();
  const fleetQuery = useQuery({ queryKey: ['collectors-fleet'], queryFn: fetchCollectorFleet });
  const pipelinesQuery = useQuery({ queryKey: ['collectors-pipelines'], queryFn: fetchCollectorPipelines });

  const [agent, setAgent] = useState({
    name: '',
    environment: 'prod',
    version: '1.2.0',
    status: 'healthy',
    lastHeartbeatAt: new Date().toISOString(),
    policyId: 'default',
  });
  const [pipeline, setPipeline] = useState({
    name: 'custom-pipeline',
    description: 'parse + mask + route',
    enabled: true,
    stages: [
      { id: 'parse-json', type: 'parse', config: { format: 'json', fields: '' }, enabled: true },
      { id: 'mask-secrets', type: 'mask', config: { format: '', fields: 'token,password' }, enabled: true },
    ] as CollectorPipelineStage[],
  });
  const [validationMessage, setValidationMessage] = useState('');

  const stageTypes = ['parse', 'mask', 'filter', 'route', 'sample', 'enrich'];

  const addStage = () => {
    setPipeline({
      ...pipeline,
      stages: [
        ...pipeline.stages,
        { id: `stage-${pipeline.stages.length + 1}`, type: 'filter', config: { format: '', fields: '' }, enabled: true },
      ],
    });
  };

  const moveStage = (index: number, dir: -1 | 1) => {
    const next = [...pipeline.stages];
    const target = index + dir;
    if (target < 0 || target >= next.length) return;
    [next[index], next[target]] = [next[target], next[index]];
    setPipeline({ ...pipeline, stages: next });
  };

  const removeStage = (index: number) => {
    setPipeline({ ...pipeline, stages: pipeline.stages.filter((_, i) => i !== index) });
  };

  const updateStage = (index: number, patch: Partial<CollectorPipelineStage>) => {
    setPipeline({
      ...pipeline,
      stages: pipeline.stages.map((s, i) => (i === index ? { ...s, ...patch } : s)),
    });
  };

  const createAgentMut = useMutation({
    mutationFn: () => createCollectorAgent(agent),
    onSuccess: () => {
      toast.success('Collector agent enrolled');
      setAgent({ ...agent, name: '' });
      void qc.invalidateQueries({ queryKey: ['collectors-fleet'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const [canaryVersion, setCanaryVersion] = useState('1.3.0');

  const upgradeMut = useMutation({
    mutationFn: ({ id, version }: { id: string; version?: string }) => upgradeCollectorAgent(id, version),
    onSuccess: () => {
      toast.success('Upgrade scheduled');
      void qc.invalidateQueries({ queryKey: ['collectors-fleet'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const createPipelineMut = useMutation({
    mutationFn: () => createCollectorPipeline(pipeline),
    onSuccess: () => {
      toast.success('Pipeline saved');
      void qc.invalidateQueries({ queryKey: ['collectors-pipelines'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const validatePipelineMut = useMutation({
    mutationFn: async () => {
      const id = (pipelinesQuery.data?.[0]?.id ?? 'preview').toString();
      return validateCollectorPipeline(id, pipeline);
    },
    onSuccess: (res) => {
      setValidationMessage(res.message);
      toast.success(res.valid ? 'Pipeline is valid' : 'Pipeline has validation issues');
    },
  });

  return (
    <StitchPageShell title="Collectors Fleet" subtitle="Agent enrollment, health, and pipeline lifecycle">
      <div className="dashboard-row-2">
        <Card title="Enroll collector agent">
          <div className="form-stack">
            <Input label="Agent name" value={agent.name} onChange={(e) => setAgent({ ...agent, name: e.target.value })} />
            <Select value={agent.environment} onChange={(e) => setAgent({ ...agent, environment: e.target.value })}>
              <option value="prod">prod</option>
              <option value="staging">staging</option>
              <option value="dev">dev</option>
            </Select>
            <Input label="Version" value={agent.version} onChange={(e) => setAgent({ ...agent, version: e.target.value })} />
            <Button variant="primary" onClick={() => createAgentMut.mutate()} disabled={!agent.name.trim()}>
              Enroll agent
            </Button>
          </div>
        </Card>
        <Card
          title="Fleet inventory"
          action={
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <span className="muted">Canary target</span>
              <Input value={canaryVersion} onChange={(e) => setCanaryVersion(e.target.value)} aria-label="Canary version" />
            </div>
          }
        >
          {(fleetQuery.data ?? []).map((f) => (
            <div key={f.id} className="list-row">
              <span>
                {f.name} <span className="muted">v{f.version} · {f.environment}</span>
              </span>
              <span style={{ display: 'flex', gap: 8 }}>
                <Badge variant={f.status === 'healthy' ? 'healthy' : 'warning'}>{f.status}</Badge>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => upgradeMut.mutate({ id: f.id })}
                  disabled={upgradeMut.isPending}
                >
                  Rolling upgrade
                </Button>
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => upgradeMut.mutate({ id: f.id, version: canaryVersion })}
                  disabled={upgradeMut.isPending}
                >
                  Canary
                </Button>
              </span>
            </div>
          ))}
        </Card>
      </div>

      <div className="dashboard-row-2">
        <Card title="Pipeline composer (COLL-10)">
          <div className="form-stack">
            <Input label="Pipeline name" value={pipeline.name} onChange={(e) => setPipeline({ ...pipeline, name: e.target.value })} />
            <Input label="Description" value={pipeline.description} onChange={(e) => setPipeline({ ...pipeline, description: e.target.value })} />
            <div data-testid="pipeline-stages">
              {pipeline.stages.map((stage, index) => (
                <div key={stage.id} className="list-row" style={{ flexDirection: 'column', alignItems: 'stretch', gap: 8 }}>
                  <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                    <Badge variant="info">{index + 1}</Badge>
                    <Select value={stage.type} onChange={(e) => updateStage(index, { type: e.target.value })}>
                      {stageTypes.map((t) => (
                        <option key={t} value={t}>{t}</option>
                      ))}
                    </Select>
                    <Input
                      aria-label={`Stage ${index + 1} id`}
                      value={stage.id}
                      onChange={(e) => updateStage(index, { id: e.target.value })}
                    />
                    <Button variant="ghost" size="sm" onClick={() => moveStage(index, -1)} disabled={index === 0}>↑</Button>
                    <Button variant="ghost" size="sm" onClick={() => moveStage(index, 1)} disabled={index === pipeline.stages.length - 1}>↓</Button>
                    <Button variant="ghost" size="sm" onClick={() => removeStage(index)}>Remove</Button>
                  </div>
                  <Input
                    label="Config fields (comma-separated keys)"
                    value={stage.config?.fields ?? ''}
                    onChange={(e) => updateStage(index, { config: { ...stage.config, fields: e.target.value } })}
                  />
                </div>
              ))}
            </div>
            <Button variant="secondary" onClick={addStage}>Add stage</Button>
            <Button variant="secondary" onClick={() => validatePipelineMut.mutate()}>
              Validate pipeline
            </Button>
            {validationMessage && <p className="muted">{validationMessage}</p>}
            <Button variant="primary" onClick={() => createPipelineMut.mutate()}>
              Save pipeline
            </Button>
          </div>
        </Card>
        <Card title="Pipelines">
          {(pipelinesQuery.data ?? []).map((p) => (
            <div key={p.id} className="list-row">
              <span>{p.name} <span className="muted">({p.stages.length} stages)</span></span>
              <Badge variant={p.enabled ? 'healthy' : 'info'}>{p.enabled ? 'Enabled' : 'Disabled'}</Badge>
            </div>
          ))}
        </Card>
      </div>
    </StitchPageShell>
  );
}
