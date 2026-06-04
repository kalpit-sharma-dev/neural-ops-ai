import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useMemo, useState } from 'react';
import toast from 'react-hot-toast';
import { Link } from '@tanstack/react-router';
import { fetchAlertPolicies, type AlertPolicy } from '../api/observability';
import {
  acknowledgeAlert,
  createAlertRule,
  createChannel,
  createSilence,
  previewAlertSilence,
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
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { DataExportMenu } from '../components/ui/DataExportMenu';
import { StitchPageShell } from '../components/stitch';
import { Input } from '../components/ui/Input';
import { Select } from '../components/ui/Select';
import { useI18n } from '../i18n/I18nProvider';

const TABS = ['Active', 'Rules', 'Channels', 'Silences', 'Escalation', 'History'] as const;

export default function Alerts() {
  const { t } = useI18n();
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

  const alertPoliciesQuery = useQuery({
    queryKey: ['obs-alert-policies'],
    queryFn: fetchAlertPolicies,
    enabled: tab === 'Active',
  });

  const policyForAlert = useMemo(() => {
    const policies = alertPoliciesQuery.data ?? [];
    return (alert: AlertRecord): AlertPolicy | undefined => {
      return policies.find(
        (p) =>
          p.enabled &&
          (p.servicePattern === '*' ||
            alert.service.toLowerCase().includes(p.servicePattern.replace('*', '').toLowerCase())),
      );
    };
  }, [alertPoliciesQuery.data]);

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
  const [selectedAlertIds, setSelectedAlertIds] = useState<Set<string>>(new Set());
  const [ruleErrors, setRuleErrors] = useState<Record<string, string>>({});

  const activeAlerts = useMemo(
    () =>
      (alertsQuery.data ?? []).filter((a) => a.status === 'FIRING' || a.status === 'ACKNOWLEDGED'),
    [alertsQuery.data],
  );

  const alertsByService = useMemo(() => {
    const groups = new Map<string, AlertRecord[]>();
    for (const alert of activeAlerts) {
      const key = alert.service || 'unknown';
      const list = groups.get(key) ?? [];
      list.push(alert);
      groups.set(key, list);
    }
    return [...groups.entries()].sort(([a], [b]) => a.localeCompare(b));
  }, [activeAlerts]);

  const clientSilencePreviewCount = useMemo(() => {
    const pattern = silenceForm.servicePattern.trim();
    if (!pattern) return activeAlerts.length;
    const re = new RegExp(`^${pattern.replace(/\*/g, '.*')}$`, 'i');
    return activeAlerts.filter((a) => re.test(a.service)).length;
  }, [activeAlerts, silenceForm.servicePattern]);

  const silencePreviewQuery = useQuery({
    queryKey: ['silence-preview', silenceForm.servicePattern],
    queryFn: () =>
      previewAlertSilence({
        servicePattern: silenceForm.servicePattern.trim() || '*',
      }),
    enabled: tab === 'Silences',
    staleTime: 10_000,
  });

  const silencePreviewCount =
    silencePreviewQuery.data?.matchedAlerts ?? clientSilencePreviewCount;

  const validateRuleForm = () => {
    const errors: Record<string, string> = {};
    if (!ruleForm.name.trim()) errors.name = 'Name is required';
    if (!ruleForm.servicePattern.trim()) errors.servicePattern = 'Service pattern is required (use * for all)';
    setRuleErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const bulkAckMut = useMutation({
    mutationFn: async (ids: string[]) => {
      await Promise.all(ids.map((id) => acknowledgeAlert(id)));
    },
    onSuccess: () => {
      toast.success('Alerts acknowledged');
      setSelectedAlertIds(new Set());
      void queryClient.invalidateQueries({ queryKey: ['alerts'] });
    },
  });

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
    <StitchPageShell
      title="Alerts"
      subtitle="Rules, channels, silences, and firing alerts"
      actions={
        <DataExportMenu
          getData={() => {
            if (tab === 'Active' || tab === 'History') return alertsQuery.data ?? [];
            if (tab === 'Rules') return rulesQuery.data ?? [];
            if (tab === 'Silences') return silencesQuery.data ?? [];
            if (tab === 'Channels') return channelsQuery.data ?? [];
            if (tab === 'Escalation') return escalationQuery.data ?? [];
            return [];
          }}
          filenamePrefix={`alerts-${tab.toLowerCase()}`}
          disabled={
            tab === 'Active' || tab === 'History'
              ? !alertsQuery.data?.length
              : tab === 'Rules'
                ? !rulesQuery.data?.length
                : tab === 'Silences'
                  ? !silencesQuery.data?.length
                  : tab === 'Channels'
                    ? !channelsQuery.data?.length
                    : tab === 'Escalation'
                      ? !escalationQuery.data?.length
                      : true
          }
        />
      }
    >
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
          {activeAlerts.length === 0 && !alertsQuery.isLoading && (
            <Card title="No active alerts">All clear — no firing alerts in the last window.</Card>
          )}
          {activeAlerts.length > 0 && (
            <div style={{ display: 'flex', gap: 8, marginBottom: 16, alignItems: 'center' }}>
              <label className="checkbox-row">
                <input
                  type="checkbox"
                  checked={selectedAlertIds.size === activeAlerts.length}
                  onChange={(e) => {
                    if (e.target.checked) setSelectedAlertIds(new Set(activeAlerts.map((a) => a.id)));
                    else setSelectedAlertIds(new Set());
                  }}
                />
                Select all
              </label>
              <Button
                variant="secondary"
                size="sm"
                disabled={selectedAlertIds.size === 0 || bulkAckMut.isPending}
                onClick={() => bulkAckMut.mutate([...selectedAlertIds])}
              >
                Acknowledge selected ({selectedAlertIds.size})
              </Button>
            </div>
          )}
          {alertsByService.map(([service, alerts]) => (
            <Card key={service} title={service} style={{ marginBottom: 16 }}>
              {alerts.map((alert) => (
                <div
                  key={alert.id}
                  className="ui-card"
                  style={{ marginBottom: 12, cursor: 'pointer' }}
                  onClick={() => setSelectedAlert(alert)}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 12 }}>
                    <div style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}>
                      <input
                        type="checkbox"
                        checked={selectedAlertIds.has(alert.id)}
                        onClick={(e) => e.stopPropagation()}
                        onChange={(e) => {
                          setSelectedAlertIds((prev) => {
                            const next = new Set(prev);
                            if (e.target.checked) next.add(alert.id);
                            else next.delete(alert.id);
                            return next;
                          });
                        }}
                        aria-label={`Select ${alert.title}`}
                      />
                      <div>
                        <Badge variant={alert.severity === 'P1' ? 'p1' : alert.severity === 'P2' ? 'p2' : 'info'}>
                          {alert.severity}
                        </Badge>
                        <strong style={{ marginLeft: 8 }}>{alert.title}</strong>
                        <p className="muted">{alert.status}</p>
                        <p>{alert.description}</p>
                        {(() => {
                          const policy = policyForAlert(alert);
                          if (!policy) return null;
                          return (
                            <div className="alert-row-context muted" data-testid={`alert-policy-context-${alert.id}`}>
                              <span>
                                Policy: <Link to="/settings/alert-policies" onClick={(e) => e.stopPropagation()}>{policy.name}</Link>
                              </span>
                              {policy.context?.owner && <span> · Owner: {policy.context.owner}</span>}
                              {policy.context?.runbookUrl && (
                                <span>
                                  {' '}
                                  · <a href={policy.context.runbookUrl} onClick={(e) => e.stopPropagation()}>Runbook</a>
                                </span>
                              )}
                            </div>
                          );
                        })()}
                      </div>
                    </div>
                    <div style={{ display: 'flex', gap: 8 }} onClick={(e) => e.stopPropagation()}>
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
            </Card>
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
            {(() => {
              const policy = policyForAlert(selectedAlert);
              if (!policy?.context) return null;
              return (
                <Card title="Policy context" style={{ marginTop: 16 }}>
                  <p>
                    <strong>Policy:</strong>{' '}
                    <Link to="/settings/alert-policies">{policy.name}</Link>
                  </p>
                  {policy.context.owner && <p>Owner: {policy.context.owner}</p>}
                  {policy.context.runbookUrl && (
                    <p>
                      Runbook: <a href={policy.context.runbookUrl}>{policy.context.runbookUrl}</a>
                    </p>
                  )}
                  <div className="incident-detail-actions" style={{ marginTop: 8 }}>
                    {policy.context.topologyLink && (
                      <Link to={policy.context.topologyLink}>Topology →</Link>
                    )}
                    {policy.context.tracePivotLink && (
                      <Link to={policy.context.tracePivotLink}>Traces →</Link>
                    )}
                    {policy.context.logPivotLink && <Link to={policy.context.logPivotLink}>Logs →</Link>}
                  </div>
                </Card>
              );
            })()}
          </>
        )}
      </Modal>

      {tab === 'Rules' && (
        <div className="dashboard-row-2">
          <Card title={editingRuleId ? 'Edit rule' : 'Create rule'}>
            <div className="form-stack">
              <Input
                label="Rule name"
                placeholder="High error rate"
                value={ruleForm.name}
                onChange={(e) => setRuleForm({ ...ruleForm, name: e.target.value })}
                error={ruleErrors.name}
              />
              <Input
                label="Service pattern"
                placeholder="payment-*"
                value={ruleForm.servicePattern}
                onChange={(e) => setRuleForm({ ...ruleForm, servicePattern: e.target.value })}
                error={ruleErrors.servicePattern}
              />
              <Select value={ruleForm.severity} onChange={(e) => setRuleForm({ ...ruleForm, severity: e.target.value })}>
                <option value="P1">P1</option>
                <option value="P2">P2</option>
                <option value="P3">P3</option>
                <option value="P4">P4</option>
              </Select>
              {editingRuleId ? (
                <>
                  <Button variant="primary" disabled={!ruleForm.name} onClick={() => { if (validateRuleForm()) updateRuleMut.mutate(); }}>Save</Button>
                  <Button variant="ghost" onClick={() => { setEditingRuleId(null); setRuleForm({ name: '', source: 'PROMETHEUS', servicePattern: '', severity: 'P2', enabled: true }); }}>Cancel</Button>
                </>
              ) : (
                <Button variant="primary" disabled={!ruleForm.name} onClick={() => { if (validateRuleForm()) createRuleMut.mutate(); }}>Create</Button>
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
              <Input label="Name" placeholder="Slack #alerts" value={channelForm.name} onChange={(e) => setChannelForm({ ...channelForm, name: e.target.value })} />
              <Select value={channelForm.channelType} onChange={(e) => setChannelForm({ ...channelForm, channelType: e.target.value })}>
                <option value="slack">Slack</option>
                <option value="pagerduty">PagerDuty</option>
                <option value="webhook">Webhook</option>
                <option value="email">Email</option>
              </Select>
              <Input label="Webhook URL" placeholder="https://…" value={channelForm.webhookUrl} onChange={(e) => setChannelForm({ ...channelForm, webhookUrl: e.target.value })} />
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
              <Input label="Name" value={escalationForm.name} onChange={(e) => setEscalationForm({ ...escalationForm, name: e.target.value })} />
              <Input label="Service pattern" value={escalationForm.servicePattern} onChange={(e) => setEscalationForm({ ...escalationForm, servicePattern: e.target.value })} />
              <Input label="Escalate after" placeholder="15m" value={escalationForm.after} onChange={(e) => setEscalationForm({ ...escalationForm, after: e.target.value })} />
              <Input label="Target" value={escalationForm.target} onChange={(e) => setEscalationForm({ ...escalationForm, target: e.target.value })} />
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
              <Input label="Service pattern" placeholder="* or payment-*" value={silenceForm.servicePattern} onChange={(e) => setSilenceForm({ ...silenceForm, servicePattern: e.target.value })} />
              <Input label="Reason" value={silenceForm.reason} onChange={(e) => setSilenceForm({ ...silenceForm, reason: e.target.value })} />
              <p className="muted" style={{ margin: 0 }} data-testid="silence-preview">
                {silencePreviewQuery.isFetching
                  ? t('page.alerts.silencePreviewLoading')
                  : (
                    <>
                      {t('page.alerts.silencePreview')}{' '}
                      <strong>{silencePreviewCount}</strong> {t('page.alerts.silencePreviewActive')}
                    </>
                  )}
              </p>
              <Select value={silenceForm.duration} onChange={(e) => setSilenceForm({ ...silenceForm, duration: e.target.value })}>
                <option value="30m">30 minutes</option>
                <option value="1h">1 hour</option>
                <option value="4h">4 hours</option>
                <option value="24h">24 hours</option>
              </Select>
              <Button variant="primary" onClick={() => createSilenceMut.mutate()}>{t('page.alerts.createSilence')}</Button>
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
    </StitchPageShell>
  );
}
