import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import toast from 'react-hot-toast';
import { Link } from '@tanstack/react-router';
import {
  acknowledgeAlert,
  createAlertRule,
  createChannel,
  createSilence,
  deleteAlertRule,
  deleteChannel,
  fetchAlertRules,
  fetchAlerts,
  fetchChannels,
  fetchSilences,
  suppressAlert,
  testChannel,
  updateAlertRule,
  updateChannel,
  createEscalationPolicy,
  fetchEscalationPolicies,
  deleteEscalationPolicy,
  type AlertRecord,
  type AlertRule,
} from '../api/alerts';
import { Modal } from '../components/ui/Modal';
import { getApiErrorMessage } from '../api/client';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { ErrorState, LoadingState, PageHeader } from '../components/ui/PageStates';
import { Select } from '../components/ui/Select';

const TABS = ['Active', 'Rules', 'Channels', 'Silences', 'Escalation', 'History'] as const;

export default function Alerts() {
  const [tab, setTab] = useState<(typeof TABS)[number]>('Active');
  const [selectedAlert, setSelectedAlert] = useState<AlertRecord | null>(null);
  const queryClient = useQueryClient();

  const alertsQuery = useQuery({
    queryKey: ['alerts', tab],
    queryFn: () =>
      fetchAlerts({
        size: 50,
        status: tab === 'History' ? 'RESOLVED' : undefined,
      }),
    refetchInterval: 30_000,
    enabled: tab === 'Active' || tab === 'History',
  });

  const rulesQuery = useQuery({ queryKey: ['alert-rules'], queryFn: fetchAlertRules, enabled: tab === 'Rules' });
  const channelsQuery = useQuery({ queryKey: ['alert-channels'], queryFn: fetchChannels, enabled: tab === 'Channels' });
  const silencesQuery = useQuery({ queryKey: ['alert-silences'], queryFn: fetchSilences, enabled: tab === 'Silences' });
  const escalationQuery = useQuery({
    queryKey: ['escalation-policies'],
    queryFn: fetchEscalationPolicies,
    enabled: tab === 'Escalation',
  });

  const [escalationForm, setEscalationForm] = useState({
    name: '',
    servicePattern: '*',
    after: '15m',
    target: 'L2 on-call',
  });

  const [ruleForm, setRuleForm] = useState({ name: '', source: 'PROMETHEUS', servicePattern: '', severity: 'P2', enabled: true });
  const [editingRuleId, setEditingRuleId] = useState<string | null>(null);
  const [channelForm, setChannelForm] = useState({ name: '', channelType: 'slack', webhookUrl: '', enabled: true });
  const [editingChannelId, setEditingChannelId] = useState<string | null>(null);
  const [silenceForm, setSilenceForm] = useState({ servicePattern: '', reason: '', duration: '1h' });

  const createRuleMut = useMutation({
    mutationFn: () => createAlertRule(ruleForm as Omit<AlertRule, 'id'>),
    onSuccess: () => {
      toast.success('Rule created');
      void queryClient.invalidateQueries({ queryKey: ['alert-rules'] });
      setRuleForm({ name: '', source: 'PROMETHEUS', servicePattern: '', severity: 'P2', enabled: true });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const deleteRuleMut = useMutation({
    mutationFn: deleteAlertRule,
    onSuccess: () => {
      toast.success('Rule deleted');
      void queryClient.invalidateQueries({ queryKey: ['alert-rules'] });
    },
  });

  const updateRuleMut = useMutation({
    mutationFn: () =>
      updateAlertRule(editingRuleId!, {
        name: ruleForm.name,
        source: ruleForm.source,
        servicePattern: ruleForm.servicePattern,
        severity: ruleForm.severity,
        enabled: ruleForm.enabled,
      }),
    onSuccess: () => {
      toast.success('Rule updated');
      setEditingRuleId(null);
      setRuleForm({ name: '', source: 'PROMETHEUS', servicePattern: '', severity: 'P2', enabled: true });
      void queryClient.invalidateQueries({ queryKey: ['alert-rules'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const updateChannelMut = useMutation({
    mutationFn: () =>
      updateChannel(editingChannelId!, {
        name: channelForm.name,
        channelType: channelForm.channelType,
        config: { webhookUrl: channelForm.webhookUrl },
        enabled: channelForm.enabled,
      }),
    onSuccess: () => {
      toast.success('Channel updated');
      setEditingChannelId(null);
      setChannelForm({ name: '', channelType: 'slack', webhookUrl: '', enabled: true });
      void queryClient.invalidateQueries({ queryKey: ['alert-channels'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const deleteChannelMut = useMutation({
    mutationFn: deleteChannel,
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['alert-channels'] }),
  });

  const createChannelMut = useMutation({
    mutationFn: () =>
      createChannel({
        name: channelForm.name,
        channelType: channelForm.channelType,
        config: { webhookUrl: channelForm.webhookUrl },
        enabled: channelForm.enabled,
      }),
    onSuccess: () => {
      toast.success('Channel created');
      void queryClient.invalidateQueries({ queryKey: ['alert-channels'] });
      setChannelForm({ name: '', channelType: 'slack', webhookUrl: '', enabled: true });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const testChannelMut = useMutation({
    mutationFn: testChannel,
    onSuccess: () => toast.success('Test notification sent'),
  });

  const createSilenceMut = useMutation({
    mutationFn: () => createSilence(silenceForm),
    onSuccess: () => {
      toast.success('Silence created');
      void queryClient.invalidateQueries({ queryKey: ['alert-silences'] });
    },
  });

  const createEscalationMut = useMutation({
    mutationFn: () =>
      createEscalationPolicy({
        name: escalationForm.name,
        servicePattern: escalationForm.servicePattern,
        enabled: true,
        levels: [{ level: 1, after: escalationForm.after, target: escalationForm.target, channelType: 'pagerduty' }],
      }),
    onSuccess: () => {
      toast.success('Escalation policy created');
      void queryClient.invalidateQueries({ queryKey: ['escalation-policies'] });
      setEscalationForm({ name: '', servicePattern: '*', after: '15m', target: 'L2 on-call' });
    },
  });

  const deleteEscalationMut = useMutation({
    mutationFn: deleteEscalationPolicy,
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['escalation-policies'] }),
  });

  const ackMut = useMutation({
    mutationFn: acknowledgeAlert,
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['alerts'] }),
  });

  const suppressMut = useMutation({
    mutationFn: (id: string) => suppressAlert(id, '30m', 'Suppressed from UI'),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['alerts'] }),
  });

  return (
    <div>
      <PageHeader title="Alerts" subtitle="Rules, channels, silences, and firing alerts" />
      <div className="tab-bar" style={{ marginBottom: 24 }}>
        {TABS.map((t) => (
          <button
            key={t}
            type="button"
            className={`tab-bar__item ${tab === t ? 'tab-bar__item--active' : ''}`}
            onClick={() => setTab(t)}
          >
            {t}
          </button>
        ))}
      </div>

      {tab === 'Active' && (
        <>
          {alertsQuery.isLoading && <LoadingState />}
          {alertsQuery.error && <ErrorState message={getApiErrorMessage(alertsQuery.error)} onRetry={() => alertsQuery.refetch()} />}
          {(alertsQuery.data ?? []).filter((a) => a.status === 'FIRING' || a.status === 'ACKNOWLEDGED').length === 0 && !alertsQuery.isLoading && (
            <Card title="No active alerts">All clear — no firing alerts in the last window.</Card>
          )}
          {(alertsQuery.data ?? []).filter((a) => a.status === 'FIRING' || a.status === 'ACKNOWLEDGED').map((alert) => (
            <div key={alert.id} className="ui-card" style={{ marginBottom: 12, cursor: 'pointer' }} onClick={() => setSelectedAlert(alert)}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 12 }}>
                <div>
                  <Badge variant={alert.severity === 'P1' ? 'p1' : alert.severity === 'P2' ? 'p2' : 'info'}>
                    {alert.severity}
                  </Badge>
                  <strong style={{ marginLeft: 8 }}>{alert.title}</strong>
                  <p className="muted">{alert.service} · {alert.status}</p>
                  <p>{alert.description}</p>
                </div>
                <div style={{ display: 'flex', gap: 8 }}>
                  {alert.status === 'FIRING' && (
                    <>
                      <Button variant="secondary" size="sm" onClick={() => ackMut.mutate(alert.id)}>Ack</Button>
                      <Button variant="ghost" size="sm" onClick={() => suppressMut.mutate(alert.id)}>Suppress</Button>
                    </>
                  )}
                </div>
              </div>
            </div>
          ))}
        </>
      )}

      {tab === 'History' && (
        <>
          {alertsQuery.isLoading && <LoadingState />}
          {(alertsQuery.data ?? []).map((alert) => (
            <div key={alert.id} className="ui-card" style={{ marginBottom: 12, cursor: 'pointer' }} onClick={() => setSelectedAlert(alert)}>
              <div style={{ display: 'flex', justifyContent: 'space-between', gap: 12 }}>
                <div>
                  <Badge variant="info">{alert.status}</Badge>
                  <strong style={{ marginLeft: 8 }}>{alert.title}</strong>
                  <p className="muted">{alert.service} · {new Date(alert.firedAt).toLocaleString()}</p>
                </div>
              </div>
            </div>
          ))}
        </>
      )}

      <Modal open={selectedAlert != null} onOpenChange={(open) => !open && setSelectedAlert(null)} title={selectedAlert?.title ?? 'Alert'}>
        {selectedAlert && (
          <>
            <p><strong>Service:</strong> {selectedAlert.service}</p>
            <p><strong>Severity:</strong> {selectedAlert.severity}</p>
            <p><strong>Status:</strong> {selectedAlert.status}</p>
            <p>{selectedAlert.description}</p>
            <p className="muted">Fired {new Date(selectedAlert.firedAt).toLocaleString()}</p>
            {selectedAlert.linkedIncidentId && (
              <Link to="/incidents/$id" params={{ id: selectedAlert.linkedIncidentId }}>View linked incident →</Link>
            )}
          </>
        )}
      </Modal>

      {tab === 'Rules' && (
        <div className="dashboard-row-2">
          <Card title={editingRuleId ? 'Edit rule' : 'Create rule'}>
            <div className="form-stack">
              <input placeholder="Rule name" value={ruleForm.name} onChange={(e) => setRuleForm({ ...ruleForm, name: e.target.value })} />
              <input placeholder="Service pattern" value={ruleForm.servicePattern} onChange={(e) => setRuleForm({ ...ruleForm, servicePattern: e.target.value })} />
              <Select value={ruleForm.severity} onChange={(e) => setRuleForm({ ...ruleForm, severity: e.target.value })}>
                <option value="P1">P1</option>
                <option value="P2">P2</option>
                <option value="P3">P3</option>
                <option value="P4">P4</option>
              </Select>
              {editingRuleId ? (
                <>
                  <Button variant="primary" disabled={!ruleForm.name} onClick={() => updateRuleMut.mutate()}>Save</Button>
                  <Button variant="ghost" onClick={() => { setEditingRuleId(null); setRuleForm({ name: '', source: 'PROMETHEUS', servicePattern: '', severity: 'P2', enabled: true }); }}>Cancel</Button>
                </>
              ) : (
                <Button variant="primary" disabled={!ruleForm.name} onClick={() => createRuleMut.mutate()}>Create</Button>
              )}
            </div>
          </Card>
          <Card title="Rules">
            {rulesQuery.isLoading && <LoadingState />}
            {(rulesQuery.data ?? []).map((rule) => (
              <div key={rule.id} className="list-row">
                <span>{rule.name}</span>
                <Badge variant={rule.enabled ? 'healthy' : 'info'}>{rule.enabled ? 'On' : 'Off'}</Badge>
                <Button variant="ghost" size="sm" onClick={() => {
                  setEditingRuleId(rule.id);
                  setRuleForm({ name: rule.name, source: rule.source, servicePattern: rule.servicePattern ?? '', severity: rule.severity, enabled: rule.enabled });
                }}>Edit</Button>
                <Button variant="ghost" size="sm" onClick={() => deleteRuleMut.mutate(rule.id)}>Delete</Button>
              </div>
            ))}
          </Card>
        </div>
      )}

      {tab === 'Channels' && (
        <div className="dashboard-row-2">
          <Card title={editingChannelId ? 'Edit channel' : 'Add channel'}>
            <div className="form-stack">
              <input placeholder="Name" value={channelForm.name} onChange={(e) => setChannelForm({ ...channelForm, name: e.target.value })} />
              <Select value={channelForm.channelType} onChange={(e) => setChannelForm({ ...channelForm, channelType: e.target.value })}>
                <option value="slack">Slack</option>
                <option value="pagerduty">PagerDuty</option>
                <option value="webhook">Webhook</option>
                <option value="email">Email</option>
              </Select>
              <input placeholder="Webhook URL" value={channelForm.webhookUrl} onChange={(e) => setChannelForm({ ...channelForm, webhookUrl: e.target.value })} />
              {editingChannelId ? (
                <>
                  <Button variant="primary" disabled={!channelForm.name} onClick={() => updateChannelMut.mutate()}>Save</Button>
                  <Button variant="ghost" onClick={() => { setEditingChannelId(null); setChannelForm({ name: '', channelType: 'slack', webhookUrl: '', enabled: true }); }}>Cancel</Button>
                </>
              ) : (
                <Button variant="primary" disabled={!channelForm.name} onClick={() => createChannelMut.mutate()}>Create</Button>
              )}
            </div>
          </Card>
          <Card title="Channels">
            {(channelsQuery.data ?? []).map((ch) => (
              <div key={ch.id} className="list-row">
                <span>{ch.name}</span>
                <Badge variant="info">{ch.channelType}</Badge>
                <Button variant="ghost" size="sm" onClick={() => {
                  setEditingChannelId(ch.id);
                  setChannelForm({
                    name: ch.name,
                    channelType: ch.channelType,
                    webhookUrl: String(ch.config?.webhookUrl ?? ''),
                    enabled: ch.enabled,
                  });
                }}>Edit</Button>
                <Button variant="ghost" size="sm" onClick={() => testChannelMut.mutate(ch.id)}>Test</Button>
                <Button variant="ghost" size="sm" onClick={() => deleteChannelMut.mutate(ch.id)}>Delete</Button>
              </div>
            ))}
          </Card>
        </div>
      )}

      {tab === 'Escalation' && (
        <div className="dashboard-row-2">
          <Card title="Create escalation policy">
            <div className="form-stack">
              <input placeholder="Name" value={escalationForm.name} onChange={(e) => setEscalationForm({ ...escalationForm, name: e.target.value })} />
              <input placeholder="Service pattern" value={escalationForm.servicePattern} onChange={(e) => setEscalationForm({ ...escalationForm, servicePattern: e.target.value })} />
              <input placeholder="Escalate after (e.g. 15m)" value={escalationForm.after} onChange={(e) => setEscalationForm({ ...escalationForm, after: e.target.value })} />
              <input placeholder="Target" value={escalationForm.target} onChange={(e) => setEscalationForm({ ...escalationForm, target: e.target.value })} />
              <Button variant="primary" disabled={!escalationForm.name} onClick={() => createEscalationMut.mutate()}>Create</Button>
            </div>
          </Card>
          <Card title="Policies">
            {(escalationQuery.data ?? []).map((p) => (
              <div key={p.id} className="list-row">
                <span>{p.name}</span>
                <span className="muted">{p.servicePattern || 'all services'}</span>
                <Button variant="ghost" size="sm" onClick={() => deleteEscalationMut.mutate(p.id)}>Delete</Button>
              </div>
            ))}
          </Card>
        </div>
      )}

      {tab === 'Silences' && (
        <div className="dashboard-row-2">
          <Card title="Maintenance window">
            <div className="form-stack">
              <input placeholder="Service pattern" value={silenceForm.servicePattern} onChange={(e) => setSilenceForm({ ...silenceForm, servicePattern: e.target.value })} />
              <input placeholder="Reason" value={silenceForm.reason} onChange={(e) => setSilenceForm({ ...silenceForm, reason: e.target.value })} />
              <Select value={silenceForm.duration} onChange={(e) => setSilenceForm({ ...silenceForm, duration: e.target.value })}>
                <option value="30m">30 minutes</option>
                <option value="1h">1 hour</option>
                <option value="4h">4 hours</option>
                <option value="24h">24 hours</option>
              </Select>
              <Button variant="primary" onClick={() => createSilenceMut.mutate()}>Create silence</Button>
            </div>
          </Card>
          <Card title="Active silences">
            {(silencesQuery.data ?? []).map((s) => (
              <div key={s.id} className="list-row">
                <span>{s.servicePattern || s.alertNamePattern || 'All'}</span>
                <span className="muted">{s.reason}</span>
              </div>
            ))}
          </Card>
        </div>
      )}
    </div>
  );
}
