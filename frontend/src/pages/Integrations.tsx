import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import toast from 'react-hot-toast';
import {
  connectIntegration,
  disconnectIntegration,
  fetchIntegrations,
  integrationOAuthStartUrl,
} from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';
import { Link } from '@tanstack/react-router';

const CONFIG_FIELDS: Record<string, { key: string; label: string; secret?: boolean }[]> = {
  jira: [
    { key: 'oauthClientId', label: 'OAuth client ID' },
    { key: 'oauthClientSecret', label: 'OAuth client secret', secret: true },
    { key: 'baseUrl', label: 'Base URL' },
    { key: 'email', label: 'Email' },
    { key: 'apiToken', label: 'API token (optional)', secret: true },
    { key: 'projectKey', label: 'Project key' },
  ],
  slack: [
    { key: 'oauthClientId', label: 'OAuth client ID' },
    { key: 'oauthClientSecret', label: 'OAuth client secret', secret: true },
    { key: 'webhookUrl', label: 'Webhook URL (optional)', secret: true },
  ],
  pagerduty: [{ key: 'routingKey', label: 'Routing key', secret: true }],
  servicenow: [
    { key: 'instanceUrl', label: 'Instance URL (https://xxx.service-now.com)' },
    { key: 'oauthClientId', label: 'OAuth client ID' },
    { key: 'oauthClientSecret', label: 'OAuth client secret', secret: true },
  ],
  aws: [
    { key: 'prometheusUrl', label: 'Prometheus URL (optional)' },
    { key: 'cloudwatchProxyUrl', label: 'CloudWatch proxy URL' },
  ],
  azure: [
    { key: 'prometheusUrl', label: 'Prometheus URL (optional)' },
    { key: 'monitorProxyUrl', label: 'Monitor proxy URL' },
  ],
  gcp: [
    { key: 'prometheusUrl', label: 'Prometheus URL (optional)' },
    { key: 'monitorProxyUrl', label: 'Monitoring proxy URL' },
  ],
};

const OAUTH_KEYS = new Set(['jira', 'slack', 'servicenow']);

export default function Integrations() {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({ queryKey: ['integrations'], queryFn: fetchIntegrations });
  const [modalKey, setModalKey] = useState<string | null>(null);
  const [editMode, setEditMode] = useState(false);
  const [form, setForm] = useState<Record<string, string>>({});

  const connectMut = useMutation({
    mutationFn: ({ id, config }: { id: string; config: Record<string, string> }) => connectIntegration(id, config),
    onSuccess: () => {
      toast.success(editMode ? 'Integration updated' : 'Integration connected');
      setModalKey(null);
      setEditMode(false);
      setForm({});
      void queryClient.invalidateQueries({ queryKey: ['integrations'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const disconnectMut = useMutation({
    mutationFn: disconnectIntegration,
    onSuccess: () => {
      toast.success('Integration disconnected');
      void queryClient.invalidateQueries({ queryKey: ['integrations'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const openConnect = (integrationKey: string, editing = false) => {
    setModalKey(integrationKey);
    setEditMode(editing);
    const fields = CONFIG_FIELDS[integrationKey] ?? [];
    setForm(Object.fromEntries(fields.map((f) => [f.key, ''])));
  };

  return (
    <div>
      <PageHeader title="Integrations" subtitle="Third-party connections" actions={<Link to="/settings">← Settings</Link>} />
      {isLoading && <LoadingState />}
      <div className="settings-grid">
        {(data ?? []).map((i) => {
          const key = i.integrationKey ?? i.id;
          return (
            <Card key={i.id} title={i.name}>
              <Badge variant={i.connected ? 'healthy' : 'info'}>{i.status}</Badge>
              <p className="muted">{i.type}</p>
              {i.configPublic && Object.keys(i.configPublic).length > 0 && (
                <ul className="muted" style={{ fontSize: 12, marginTop: 8 }}>
                  {Object.entries(i.configPublic).map(([k, v]) => (
                    <li key={k}>{k}: {v}</li>
                  ))}
                </ul>
              )}
              <div style={{ display: 'flex', gap: 8, marginTop: 12, flexWrap: 'wrap' }}>
                {!i.connected && (
                  <>
                    <Button variant="secondary" size="sm" onClick={() => openConnect(key)}>
                      Connect
                    </Button>
                    {OAUTH_KEYS.has(key) && (
                      <Button variant="secondary" size="sm" onClick={() => { window.location.href = integrationOAuthStartUrl(key); }}>
                        OAuth
                      </Button>
                    )}
                  </>
                )}
                {i.connected && (
                  <>
                    <Button variant="secondary" size="sm" onClick={() => openConnect(key, true)}>Edit</Button>
                    <Button variant="secondary" size="sm" onClick={() => disconnectMut.mutate(key)}>Disconnect</Button>
                  </>
                )}
              </div>
            </Card>
          );
        })}
      </div>

      {modalKey && (
        <div className="modal-overlay" role="dialog" aria-modal="true">
          <Card title={editMode ? 'Edit integration' : 'Connect integration'} style={{ maxWidth: 480, margin: '10vh auto' }}>
            <div className="form-stack">
              {(CONFIG_FIELDS[modalKey] ?? []).map((f) => (
                <label key={f.key}>
                  {f.label}
                  <input
                    type={f.secret ? 'password' : 'text'}
                    value={form[f.key] ?? ''}
                    onChange={(e) => setForm((prev) => ({ ...prev, [f.key]: e.target.value }))}
                  />
                </label>
              ))}
              <div style={{ display: 'flex', gap: 8 }}>
                <Button variant="primary" onClick={() => connectMut.mutate({ id: modalKey, config: form })}>
                  Save & connect
                </Button>
                <Button variant="secondary" onClick={() => { setModalKey(null); setEditMode(false); }}>Cancel</Button>
              </div>
            </div>
          </Card>
        </div>
      )}
    </div>
  );
}
