import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import toast from 'react-hot-toast';
import {
  createLogMetricRule,
  createLogParsingRule,
  fetchLogMetricRules,
  fetchLogParsingRules,
  fetchLogTierPolicy,
  updateLogTierPolicy,
} from '../api/observability';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';

export default function LogSettings() {
  const qc = useQueryClient();
  const metricsQuery = useQuery({ queryKey: ['log-metric-rules'], queryFn: fetchLogMetricRules });
  const parsingQuery = useQuery({ queryKey: ['log-parsing-rules'], queryFn: fetchLogParsingRules });
  const tierQuery = useQuery({ queryKey: ['log-tier-policy'], queryFn: fetchLogTierPolicy });
  const [metricForm, setMetricForm] = useState({ name: '', pattern: '', service: '', enabled: true });
  const [parseForm, setParseForm] = useState({ name: '', pattern: '', field: 'message', enabled: true });

  const createMetric = useMutation({
    mutationFn: () => createLogMetricRule(metricForm),
    onSuccess: () => { toast.success('Rule saved'); void qc.invalidateQueries({ queryKey: ['log-metric-rules'] }); },
  });
  const createParse = useMutation({
    mutationFn: () => createLogParsingRule(parseForm),
    onSuccess: () => { toast.success('Rule saved'); void qc.invalidateQueries({ queryKey: ['log-parsing-rules'] }); },
  });

  const tierMut = useMutation({
    mutationFn: updateLogTierPolicy,
    onSuccess: () => {
      toast.success('Tier policy saved');
      void qc.invalidateQueries({ queryKey: ['log-tier-policy'] });
    },
  });
  const tier = tierQuery.data;

  return (
    <StitchPageShell
      title="Log Settings"
      subtitle="Parsing rules and log-based metrics"
    >
      <Card title="Hot / warm / cold tiering" data-testid="log-tier-policy-card">
        {tier && (
          <div className="form-stack">
            <Input
              label="Hot retention (days)"
              type="number"
              value={String(tier.hotRetentionDays)}
              data-testid="log-tier-hot-days"
              onChange={(e) =>
                tierMut.mutate({ ...tier, hotRetentionDays: Number(e.target.value) || tier.hotRetentionDays })
              }
            />
            <Input
              label="Warm retention (days)"
              type="number"
              value={String(tier.warmRetentionDays)}
              data-testid="log-tier-warm-days"
              onChange={(e) =>
                tierMut.mutate({ ...tier, warmRetentionDays: Number(e.target.value) || tier.warmRetentionDays })
              }
            />
            <Input
              label="Cold retention (days)"
              type="number"
              value={String(tier.coldRetentionDays)}
              data-testid="log-tier-cold-days"
              onChange={(e) =>
                tierMut.mutate({ ...tier, coldRetentionDays: Number(e.target.value) || tier.coldRetentionDays })
              }
            />
            <p className="muted">Restore SLA: {tier.restoreSlaHours}h</p>
          </div>
        )}
      </Card>
      <div className="dashboard-row-2">
        <Card title="Log metric rules">
          <div className="form-stack">
            <Input label="Name" value={metricForm.name} onChange={(e) => setMetricForm({ ...metricForm, name: e.target.value })} />
            <Input label="Pattern (regex)" value={metricForm.pattern} onChange={(e) => setMetricForm({ ...metricForm, pattern: e.target.value })} />
            <Button variant="primary" onClick={() => createMetric.mutate()}>Add rule</Button>
          </div>
          {(metricsQuery.data ?? []).map((r) => <div key={r.id} className="list-row">{r.name} — <code>{r.pattern}</code></div>)}
        </Card>
        <Card title="Parsing rules">
          <div className="form-stack">
            <Input label="Name" value={parseForm.name} onChange={(e) => setParseForm({ ...parseForm, name: e.target.value })} />
            <Input label="Grok/regex" value={parseForm.pattern} onChange={(e) => setParseForm({ ...parseForm, pattern: e.target.value })} />
            <Button variant="primary" onClick={() => createParse.mutate()}>Add rule</Button>
          </div>
          {parsingQuery.isLoading && <LoadingState />}
          {(parsingQuery.data ?? []).map((r) => <div key={r.id} className="list-row">{r.name} → {r.field}</div>)}
        </Card>
      </div>
    </StitchPageShell>
  );
}
