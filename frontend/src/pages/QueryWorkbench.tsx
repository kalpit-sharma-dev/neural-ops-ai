import { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import {
  unifiedQuery,
  validateUnifiedQuery,
  explainUnifiedQuery,
  fetchUnifiedQueryFunctions,
  fetchSavedQueries,
  createSavedQuery,
  deleteSavedQuery,
  fetchAlertPolicies,
  createAlertPolicy,
  deleteAlertPolicy,
  fetchAlertSuppressions,
  createAlertSuppression,
  deleteAlertSuppression,
  type UnifiedQueryHit,
  type QueryExplainResponse,
} from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { StitchPageShell } from '../components/stitch';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Select } from '../components/ui/Select';
import { Badge } from '../components/ui/Badge';

export default function QueryWorkbench() {
  const qc = useQueryClient();
  const [query, setQuery] = useState('error rate payment-service');
  const [from, setFrom] = useState<'all' | 'logs' | 'metrics' | 'traces' | 'events'>('all');
  const [service, setService] = useState('payment-service');
  const [traceId, setTraceId] = useState('');
  const [txnId, setTxnId] = useState('');
  const [hits, setHits] = useState<UnifiedQueryHit[]>([]);
  const [explain, setExplain] = useState<QueryExplainResponse | null>(null);
  const [saveName, setSaveName] = useState('');

  const savedQuery = useQuery({ queryKey: ['saved-queries'], queryFn: fetchSavedQueries });

  const [policyName, setPolicyName] = useState('P1 payment routing');
  const [policyPattern, setPolicyPattern] = useState('payment-*');
  const [policySeverity, setPolicySeverity] = useState('P1');

  const [suppPattern, setSuppPattern] = useState('payment-*');
  const [suppReason, setSuppReason] = useState('maintenance window');
  const [suppDuration, setSuppDuration] = useState('30m');

  const functionsQuery = useQuery({ queryKey: ['query-functions'], queryFn: fetchUnifiedQueryFunctions });
  const policyQuery = useQuery({ queryKey: ['alerts-policies'], queryFn: fetchAlertPolicies });
  const suppressionQuery = useQuery({ queryKey: ['alerts-suppressions'], queryFn: fetchAlertSuppressions });

  const explainMut = useMutation({
    mutationFn: () => explainUnifiedQuery({ query, from, service, traceId: traceId || undefined, txnId: txnId || undefined, limit: 50 }),
    onSuccess: (data) => {
      setExplain(data);
      toast.success(`Plan: ${data.stores.join(', ')} (~${data.estimatedMs}ms)`);
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const runMut = useMutation({
    mutationFn: async () => {
      const validation = await validateUnifiedQuery({ query, from, service, traceId: traceId || undefined, txnId: txnId || undefined, limit: 50 });
      if (!validation.valid) throw new Error(validation.message);
      if (validation.explain) setExplain(validation.explain);
      return unifiedQuery({ query, from, service, traceId: traceId || undefined, txnId: txnId || undefined, limit: 50 });
    },
    onSuccess: (data) => {
      setHits(data.hits ?? []);
      toast.success(`Query returned ${data.count} hit(s)`);
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const saveQueryMut = useMutation({
    mutationFn: () =>
      createSavedQuery({
        id: '',
        name: saveName.trim(),
        query,
        from,
        service,
        traceId: traceId || undefined,
        txnId: txnId || undefined,
      }),
    onSuccess: () => {
      toast.success('Query saved');
      setSaveName('');
      void qc.invalidateQueries({ queryKey: ['saved-queries'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const deleteSavedMut = useMutation({
    mutationFn: deleteSavedQuery,
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['saved-queries'] }),
  });

  const createPolicyMut = useMutation({
    mutationFn: () =>
      createAlertPolicy({
        name: policyName,
        servicePattern: policyPattern,
        severity: policySeverity,
        enabled: true,
        routes: [
          { channel: 'slack', target: '#oncall-core', after: '0m', priority: 1 },
          { channel: 'pagerduty', target: 'primary', after: '5m', priority: 2 },
        ],
      }),
    onSuccess: () => {
      toast.success('Alert policy created');
      void qc.invalidateQueries({ queryKey: ['alerts-policies'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const deletePolicyMut = useMutation({
    mutationFn: deleteAlertPolicy,
    onSuccess: () => {
      toast.success('Alert policy deleted');
      void qc.invalidateQueries({ queryKey: ['alerts-policies'] });
    },
  });

  const createSuppressionMut = useMutation({
    mutationFn: () => createAlertSuppression({ servicePattern: suppPattern, reason: suppReason, duration: suppDuration }),
    onSuccess: () => {
      toast.success('Suppression created');
      void qc.invalidateQueries({ queryKey: ['alerts-suppressions'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const deleteSuppressionMut = useMutation({
    mutationFn: deleteAlertSuppression,
    onSuccess: () => {
      toast.success('Suppression deleted');
      void qc.invalidateQueries({ queryKey: ['alerts-suppressions'] });
    },
  });

  const groupedBySignal = useMemo(() => {
    const bySignal = new Map<string, UnifiedQueryHit[]>();
    for (const hit of hits) {
      const list = bySignal.get(hit.signal) ?? [];
      list.push(hit);
      bySignal.set(hit.signal, list);
    }
    return [...bySignal.entries()];
  }, [hits]);

  return (
    <StitchPageShell title="Query Workbench" subtitle="NexQL cross-signal search with explain plan and saved queries">
      <Card title="Unified query">
        <div className="form-stack">
          <Input label="Query" value={query} onChange={(e) => setQuery(e.target.value)} />
          <Input label="Service" value={service} onChange={(e) => setService(e.target.value)} />
          <Input label="Trace ID (join)" value={traceId} onChange={(e) => setTraceId(e.target.value)} placeholder="optional cross-signal join" />
          <Input label="Txn ID (join)" value={txnId} onChange={(e) => setTxnId(e.target.value)} placeholder="optional UPI/ledger correlation" />
          <Select value={from} onChange={(e) => setFrom(e.target.value as typeof from)}>
            <option value="all">All signals</option>
            <option value="logs">Logs</option>
            <option value="metrics">Metrics</option>
            <option value="traces">Traces</option>
            <option value="events">Events</option>
          </Select>
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
            <Button variant="secondary" onClick={() => explainMut.mutate()} disabled={explainMut.isPending || !query.trim()}>
              Explain plan
            </Button>
            <Button variant="primary" onClick={() => runMut.mutate()} disabled={runMut.isPending || !query.trim()}>
              Run query
            </Button>
          </div>
          {explain && (
            <div className="muted" data-testid="query-explain">
              Stores: {explain.stores.join(', ')} · Steps: {explain.steps.length} · Join keys: {(explain.joinKeys ?? []).join(', ') || 'none'} · Est. {explain.estimatedMs}ms
            </div>
          )}
          <p className="muted" style={{ margin: 0 }}>
            Functions: {(functionsQuery.data ?? []).join(', ')}
          </p>
          <div style={{ display: 'flex', gap: 8, alignItems: 'flex-end' }}>
            <Input label="Save as" value={saveName} onChange={(e) => setSaveName(e.target.value)} />
            <Button variant="ghost" onClick={() => saveQueryMut.mutate()} disabled={!saveName.trim() || saveQueryMut.isPending}>
              Save query
            </Button>
          </div>
          {(savedQuery.data ?? []).map((sq) => (
            <div key={sq.id} className="list-row">
              <button type="button" className="link-button" onClick={() => { setQuery(sq.query); setService(sq.service ?? ''); setFrom((sq.from as typeof from) || 'all'); setTraceId(sq.traceId ?? ''); setTxnId(sq.txnId ?? ''); }}>
                {sq.name}
              </button>
              <Button variant="ghost" size="sm" onClick={() => deleteSavedMut.mutate(sq.id)}>Delete</Button>
            </div>
          ))}
        </div>
      </Card>

      {groupedBySignal.map(([signal, signalHits]) => (
        <Card key={signal} title={`${signal.toUpperCase()} (${signalHits.length})`}>
          {signalHits.map((h) => (
            <div key={h.id} className="list-row">
              <span>
                <strong>{h.title}</strong>
                <span className="muted" style={{ marginLeft: 8 }}>{h.summary}</span>
              </span>
              <span style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                {h.severity && <Badge variant={h.severity === 'P1' ? 'critical' : h.severity === 'P2' ? 'warning' : 'info'}>{h.severity}</Badge>}
                <span className="muted">{new Date(h.timestamp).toLocaleString()}</span>
              </span>
            </div>
          ))}
        </Card>
      ))}

      <div className="dashboard-row-2">
        <Card title="Alert policies">
          <div className="form-stack">
            <Input label="Name" value={policyName} onChange={(e) => setPolicyName(e.target.value)} />
            <Input label="Service pattern" value={policyPattern} onChange={(e) => setPolicyPattern(e.target.value)} />
            <Select value={policySeverity} onChange={(e) => setPolicySeverity(e.target.value)}>
              <option value="P1">P1</option>
              <option value="P2">P2</option>
              <option value="P3">P3</option>
            </Select>
            <Button variant="primary" onClick={() => createPolicyMut.mutate()} disabled={!policyName.trim()}>
              Create policy
            </Button>
            {(policyQuery.data ?? []).map((p) => (
              <div key={p.id} className="list-row">
                <span>{p.name} <span className="muted">({p.servicePattern})</span></span>
                <Button variant="ghost" size="sm" onClick={() => deletePolicyMut.mutate(p.id)}>Delete</Button>
              </div>
            ))}
          </div>
        </Card>

        <Card title="Alert suppressions">
          <div className="form-stack">
            <Input label="Service pattern" value={suppPattern} onChange={(e) => setSuppPattern(e.target.value)} />
            <Input label="Reason" value={suppReason} onChange={(e) => setSuppReason(e.target.value)} />
            <Select value={suppDuration} onChange={(e) => setSuppDuration(e.target.value)}>
              <option value="30m">30m</option>
              <option value="1h">1h</option>
              <option value="4h">4h</option>
            </Select>
            <Button variant="primary" onClick={() => createSuppressionMut.mutate()} disabled={!suppReason.trim()}>
              Create suppression
            </Button>
            {(suppressionQuery.data ?? []).map((s) => (
              <div key={s.id} className="list-row">
                <span>{s.servicePattern || '*'} <span className="muted">until {new Date(s.endsAt).toLocaleString()}</span></span>
                <Button variant="ghost" size="sm" onClick={() => deleteSuppressionMut.mutate(s.id)}>Delete</Button>
              </div>
            ))}
          </div>
        </Card>
      </div>
    </StitchPageShell>
  );
}
