import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  fetchCloudAssetTopology,
  fetchCloudAssets,
  fetchNetworkAnomalies,
  fetchNetworkDevices,
  fetchNetworkFlows,
  fetchNetworkTopology,
  fetchNetFlowRecords,
  fetchSDWANTunnels,
  fetchWirelessLinks,
} from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { StitchPageShell } from '../components/stitch';
import { Card } from '../components/ui/Card';
import { Badge } from '../components/ui/Badge';
import { Select } from '../components/ui/Select';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { FinOpsPanel } from '../components/finops/FinOpsPanel';

const TABS = ['Cloud assets', 'FinOps', 'Network (NPM)'] as const;
const NPM_SUBTABS = ['Flows', 'NetFlow', 'SD-WAN', 'Wireless'] as const;
const PROVIDERS = ['', 'aws', 'gcp', 'azure'] as const;

function healthVariant(h: string): 'healthy' | 'degraded' | 'down' | 'warning' {
  if (h === 'healthy' || h === 'up') return 'healthy';
  if (h === 'degraded') return 'degraded';
  if (h === 'down') return 'down';
  return 'warning';
}

function severityVariant(s: string): 'critical' | 'warning' | 'info' {
  if (s === 'high' || s === 'critical') return 'critical';
  if (s === 'medium') return 'warning';
  return 'info';
}

export default function CloudFinOpsNetwork() {
  const [tab, setTab] = useState<(typeof TABS)[number]>('Cloud assets');
  const [npmTab, setNpmTab] = useState<(typeof NPM_SUBTABS)[number]>('Flows');
  const [provider, setProvider] = useState<(typeof PROVIDERS)[number]>('');
  const [costScope, setCostScope] = useState('all');

  const assetsQuery = useQuery({
    queryKey: ['cloud-assets', provider],
    queryFn: () => fetchCloudAssets(provider || undefined),
    enabled: tab === 'Cloud assets',
  });
  const cloudTopoQuery = useQuery({
    queryKey: ['cloud-topology'],
    queryFn: fetchCloudAssetTopology,
    enabled: tab === 'Cloud assets',
  });
  const flowsQuery = useQuery({
    queryKey: ['network-flows'],
    queryFn: () => fetchNetworkFlows(25),
    enabled: tab === 'Network (NPM)' && npmTab === 'Flows',
  });
  const devicesQuery = useQuery({
    queryKey: ['network-devices'],
    queryFn: fetchNetworkDevices,
    enabled: tab === 'Network (NPM)' && npmTab === 'Flows',
  });
  const netTopoQuery = useQuery({
    queryKey: ['network-topology'],
    queryFn: fetchNetworkTopology,
    enabled: tab === 'Network (NPM)' && npmTab === 'Flows',
  });
  const netAnomQuery = useQuery({
    queryKey: ['network-anomalies'],
    queryFn: fetchNetworkAnomalies,
    enabled: tab === 'Network (NPM)' && npmTab === 'Flows',
  });
  const netflowQuery = useQuery({
    queryKey: ['netflow-records'],
    queryFn: fetchNetFlowRecords,
    enabled: tab === 'Network (NPM)' && npmTab === 'NetFlow',
  });
  const sdwanQuery = useQuery({
    queryKey: ['sdwan-tunnels'],
    queryFn: fetchSDWANTunnels,
    enabled: tab === 'Network (NPM)' && npmTab === 'SD-WAN',
  });
  const wirelessQuery = useQuery({
    queryKey: ['wireless-links'],
    queryFn: fetchWirelessLinks,
    enabled: tab === 'Network (NPM)' && npmTab === 'Wireless',
  });

  return (
    <StitchPageShell
      title="Cloud, FinOps & Network"
      subtitle="Multi-cloud asset graph, cost intelligence, and NPM path observability."
    >
      <div className="tab-bar" style={{ marginBottom: 24 }}>
        {TABS.map((t) => (
          <button
            key={t}
            type="button"
            className={`tab-bar__item ${tab === t ? 'tab-bar__item--active' : ''}`}
            onClick={() => setTab(t)}
            data-testid={`cfn-tab-${t.replace(/\s+/g, '-').toLowerCase()}`}
          >
            {t}
          </button>
        ))}
      </div>

      <div style={{ marginBottom: 16, maxWidth: 240 }}>
        <Select
          label="Cloud provider filter"
          value={provider}
          onChange={(e) => setProvider(e.target.value as (typeof PROVIDERS)[number])}
        >
          <option value="">All providers</option>
          <option value="aws">AWS</option>
          <option value="gcp">GCP</option>
          <option value="azure">Azure</option>
        </Select>
      </div>

      {tab === 'Cloud assets' && (
        <>
          {assetsQuery.isLoading && <LoadingState />}
          {assetsQuery.error && <ErrorState message={getApiErrorMessage(assetsQuery.error)} onRetry={() => assetsQuery.refetch()} />}
          <div style={{ display: 'grid', gap: 16 }}>
            {(assetsQuery.data ?? []).map((a) => (
              <Card key={a.id} title={a.name} data-testid={`cloud-asset-${a.id}`}>
                <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                  <Badge variant="info">{a.provider.toUpperCase()}</Badge>
                  <Badge variant="info">{a.type}</Badge>
                  <Badge variant={healthVariant(a.status)}>{a.status}</Badge>
                </div>
                <p className="muted" style={{ marginTop: 8 }}>
                  {a.region} · {a.accountId}
                  {a.monthlyUsd != null ? ` · $${a.monthlyUsd.toFixed(0)}/mo` : ''}
                </p>
              </Card>
            ))}
          </div>
          {cloudTopoQuery.data && (
            <Card title="Asset topology" style={{ marginTop: 24 }} data-testid="cloud-topology-card">
              <p className="muted">{cloudTopoQuery.data.nodes.length} nodes · {cloudTopoQuery.data.edges.length} relationships</p>
              <ul className="insight-list" style={{ marginTop: 12 }}>
                {cloudTopoQuery.data.edges.map((e, i) => (
                  <li key={i}>
                    <code>{e.source}</code> —{e.relation}→ <code>{e.target}</code>
                  </li>
                ))}
              </ul>
            </Card>
          )}
        </>
      )}

      {tab === 'FinOps' && (
        <FinOpsPanel costScope={costScope} provider={provider} onScopeChange={setCostScope} />
      )}

      {tab === 'Network (NPM)' && (
        <>
          <div className="tab-bar" style={{ marginBottom: 16 }}>
            {NPM_SUBTABS.map((st) => (
              <button
                key={st}
                type="button"
                className={`tab-bar__item ${npmTab === st ? 'tab-bar__item--active' : ''}`}
                onClick={() => setNpmTab(st)}
                data-testid={`npm-subtab-${st.toLowerCase()}`}
              >
                {st}
              </button>
            ))}
          </div>

          {npmTab === 'Flows' && (
            <>
              {devicesQuery.isLoading && <LoadingState />}
              <Card title="Devices" data-testid="npm-devices-card">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Type</th>
                  <th>Site</th>
                  <th>Status</th>
                  <th>CPU</th>
                  <th>Uptime</th>
                </tr>
              </thead>
              <tbody>
                {(devicesQuery.data ?? []).map((d) => (
                  <tr key={d.id}>
                    <td>{d.name}</td>
                    <td>{d.type}</td>
                    <td>{d.site}</td>
                    <td>
                      <Badge variant={healthVariant(d.status)}>{d.status}</Badge>
                    </td>
                    <td>{d.cpuUtil}%</td>
                    <td>{d.uptimePct}%</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
          <Card title="Hot flows" style={{ marginTop: 24 }} data-testid="npm-flows-card">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Source</th>
                  <th>Dest</th>
                  <th>Latency</th>
                  <th>Loss</th>
                  <th>Jitter</th>
                </tr>
              </thead>
              <tbody>
                {(flowsQuery.data ?? []).map((f) => (
                  <tr key={f.id}>
                    <td>
                      {f.source}:{f.port}
                    </td>
                    <td>{f.destination}</td>
                    <td>{f.latencyMs} ms</td>
                    <td>{f.lossPct}%</td>
                    <td>{f.jitterMs} ms</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
          {netTopoQuery.data && (
            <Card title="Network topology" style={{ marginTop: 24 }} data-testid="npm-topology-card">
              <ul className="insight-list">
                {netTopoQuery.data.edges.map((e, i) => (
                  <li key={i}>
                    <code>{e.source}</code> → <code>{e.target}</code> · {e.latencyMs}ms latency · {e.lossPct}% loss ·{' '}
                    {e.utilPct}% util
                  </li>
                ))}
              </ul>
            </Card>
          )}
          <Card title="NPM anomalies" style={{ marginTop: 24 }} data-testid="npm-anomalies-card">
            {(netAnomQuery.data ?? []).map((a) => (
              <div key={a.id} style={{ marginBottom: 12 }}>
                <Badge variant={severityVariant(a.severity)}>{a.severity}</Badge> {a.link} — {a.metric}: {a.value} (baseline{' '}
                {a.baseline})
                <p className="muted">{a.description}</p>
              </div>
            ))}
          </Card>
            </>
          )}

          {npmTab === 'NetFlow' && (
            <Card title="NetFlow exports" data-testid="npm-netflow-card">
              {netflowQuery.isLoading && <LoadingState />}
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Exporter</th>
                    <th>Src</th>
                    <th>Dst</th>
                    <th>App</th>
                    <th>Bytes</th>
                  </tr>
                </thead>
                <tbody>
                  {(netflowQuery.data ?? []).map((r) => (
                    <tr key={r.id}>
                      <td>{r.exporter}</td>
                      <td>{r.srcIp}</td>
                      <td>{r.dstIp}</td>
                      <td>{r.application}</td>
                      <td>{r.bytes.toLocaleString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </Card>
          )}

          {npmTab === 'SD-WAN' && (
            <Card title="SD-WAN tunnels" data-testid="npm-sdwan-card">
              {sdwanQuery.isLoading && <LoadingState />}
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Site</th>
                    <th>Provider</th>
                    <th>Latency</th>
                    <th>Loss</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {(sdwanQuery.data ?? []).map((t) => (
                    <tr key={t.id}>
                      <td>{t.site}</td>
                      <td>{t.provider}</td>
                      <td>{t.latencyMs} ms</td>
                      <td>{t.lossPct}%</td>
                      <td>
                        <Badge variant={healthVariant(t.status)}>{t.status}</Badge>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </Card>
          )}

          {npmTab === 'Wireless' && (
            <Card title="Wireless links" data-testid="npm-wireless-card">
              {wirelessQuery.isLoading && <LoadingState />}
              <table className="data-table">
                <thead>
                  <tr>
                    <th>AP</th>
                    <th>SSID</th>
                    <th>Client</th>
                    <th>RSSI</th>
                    <th>Mbps</th>
                  </tr>
                </thead>
                <tbody>
                  {(wirelessQuery.data ?? []).map((w) => (
                    <tr key={w.id}>
                      <td>{w.apName}</td>
                      <td>{w.ssid}</td>
                      <td>{w.clientMac}</td>
                      <td>{w.rssiDbm} dBm</td>
                      <td>{w.throughputMbps}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </Card>
          )}
        </>
      )}
    </StitchPageShell>
  );
}
