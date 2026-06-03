import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import {
  createAlertPolicy,
  deleteAlertPolicy,
  fetchAlertPolicies,
  fetchAlertPolicyFeedback,
  fetchAlertPolicyScores,
  submitAlertPolicyFeedback,
  triggerAlertPolicy,
  type AlertPolicy,
} from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { StitchPageShell, SettingsBreadcrumb } from '../components/stitch';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Badge } from '../components/ui/Badge';
import { LoadingState } from '../components/ui/PageStates';
import { useState } from 'react';

export default function AlertPolicies() {
  const qc = useQueryClient();
  const { data, isLoading } = useQuery({ queryKey: ['alert-policies'], queryFn: fetchAlertPolicies });
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [form, setForm] = useState({
    name: '',
    servicePattern: '*',
    severity: 'P2',
    enabled: true,
  });

  const scoresQuery = useQuery({
    queryKey: ['alert-scores', expandedId],
    queryFn: () => fetchAlertPolicyScores(expandedId!),
    enabled: !!expandedId,
  });

  const feedbackQuery = useQuery({
    queryKey: ['alert-feedback', expandedId],
    queryFn: () => fetchAlertPolicyFeedback(expandedId!),
    enabled: !!expandedId,
  });

  const createMut = useMutation({
    mutationFn: () =>
      createAlertPolicy({
        ...form,
        routes: [{ channel: 'slack', target: '#oncall', after: '0m', priority: 1 }],
      }),
    onSuccess: () => {
      toast.success('Policy created');
      void qc.invalidateQueries({ queryKey: ['alert-policies'] });
      setForm({ name: '', servicePattern: '*', severity: 'P2', enabled: true });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const deleteMut = useMutation({
    mutationFn: deleteAlertPolicy,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alert-policies'] }),
  });

  const triggerMut = useMutation({
    mutationFn: (p: AlertPolicy) =>
      triggerAlertPolicy(p.id, { service: p.servicePattern.replace('*', 'payment-service'), severity: p.severity }),
    onSuccess: (res, p) => {
      toast.success(res.matched ? 'Policy matched — routes scheduled' : res.explanation ?? 'No match');
      if (expandedId === p.id) void scoresQuery.refetch();
    },
  });

  const feedbackMut = useMutation({
    mutationFn: (p: AlertPolicy) =>
      submitAlertPolicyFeedback(p.id, { service: 'payment-service', helpful: true, comment: 'Useful routing context' }),
    onSuccess: () => {
      toast.success('Feedback recorded');
      void feedbackQuery.refetch();
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  if (isLoading) return <LoadingState />;

  return (
    <StitchPageShell
      title="Alert policies"
      subtitle="Routing, escalation, runbook context, and materialized fatigue scores"
      breadcrumb={<SettingsBreadcrumb page="Alert policies" />}
    >
      <Card title="Create policy">
        <form
          className="form-stack"
          onSubmit={(e) => {
            e.preventDefault();
            createMut.mutate();
          }}
        >
          <Input label="Name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
          <Input
            label="Service pattern"
            value={form.servicePattern}
            onChange={(e) => setForm({ ...form, servicePattern: e.target.value })}
          />
          <Input label="Severity" value={form.severity} onChange={(e) => setForm({ ...form, severity: e.target.value })} />
          <Button type="submit" variant="primary" disabled={createMut.isPending}>
            Save policy
          </Button>
        </form>
      </Card>

      <Card title="Policies">
        {(data ?? []).map((p) => (
          <div key={p.id} style={{ marginBottom: 16 }}>
            <div className="list-row" style={{ flexWrap: 'wrap', gap: 8 }}>
              <span>
                <strong>{p.name}</strong>
                <span className="muted" style={{ marginLeft: 8 }}>
                  {p.servicePattern} · {p.severity}
                </span>
              </span>
              <Badge variant={p.enabled ? 'healthy' : 'info'}>{p.enabled ? 'Enabled' : 'Disabled'}</Badge>
              {p.context?.owner && <span className="muted">Owner: {p.context.owner}</span>}
              <Button size="sm" variant="secondary" onClick={() => triggerMut.mutate(p)}>
                Test trigger
              </Button>
              <Button
                size="sm"
                variant="ghost"
                data-testid={`policy-scores-${p.id}`}
                onClick={() => setExpandedId(expandedId === p.id ? null : p.id)}
              >
                {expandedId === p.id ? 'Hide scores' : 'Scores & feedback'}
              </Button>
              <Button size="sm" variant="ghost" onClick={() => deleteMut.mutate(p.id)}>
                Delete
              </Button>
            </div>
            {expandedId === p.id && (
              <div style={{ marginTop: 8, paddingLeft: 12, borderLeft: '2px solid var(--border-subtle)' }}>
                {scoresQuery.isLoading && <LoadingState />}
                {(scoresQuery.data ?? []).map((s) => (
                  <div key={`${s.policyId}-${s.bucketTs}`} className="list-row">
                    <span>{s.service}</span>
                    <Badge variant={s.fatigueScore > 5 ? 'critical' : 'warning'}>
                      fatigue {s.fatigueScore.toFixed(2)}
                    </Badge>
                    <span className="muted">{new Date(s.bucketTs).toLocaleString()}</span>
                  </div>
                ))}
                {(scoresQuery.data ?? []).length === 0 && !scoresQuery.isLoading && (
                  <p className="muted">No materialized scores yet — trigger the policy or run the materializer.</p>
                )}
                <Button size="sm" variant="secondary" style={{ marginTop: 8 }} onClick={() => feedbackMut.mutate(p)}>
                  Submit helpful feedback
                </Button>
                {(feedbackQuery.data ?? []).map((f) => (
                  <div key={f.id} className="list-row">
                    <Badge variant={f.helpful ? 'success' : 'warning'}>{f.helpful ? 'Helpful' : 'Noisy'}</Badge>
                    <span>{f.service}</span>
                    <span className="muted">{f.comment}</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        ))}
      </Card>
    </StitchPageShell>
  );
}
