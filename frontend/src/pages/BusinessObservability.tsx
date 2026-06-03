import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import {
  createRUMFunnel,
  createSyntheticBrowserTest,
  createSyntheticMobileTest,
  enableBusinessKPIPack,
  fetchBusinessKPIPacks,
  fetchRUMFunnels,
  fetchSyntheticBrowserTests,
  fetchSyntheticMobileTests,
  fetchSyntheticPrivateLocations,
  registerSyntheticPrivateLocation,
} from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { StitchPageShell } from '../components/stitch';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Textarea } from '../components/ui/Textarea';
import { Badge } from '../components/ui/Badge';
import { LoadingState } from '../components/ui/PageStates';

const TABS = ['RUM funnels', 'Synthetic studio', 'Business KPI packs'] as const;

export default function BusinessObservability() {
  const qc = useQueryClient();
  const [tab, setTab] = useState<(typeof TABS)[number]>('RUM funnels');

  const [funnelName, setFunnelName] = useState('Signup funnel');
  const [funnelSteps, setFunnelSteps] = useState('Landing|page_landing\nSignup|signup_complete\nActivated|activation_complete');

  const [browserName, setBrowserName] = useState('');
  const [browserURL, setBrowserURL] = useState('https://demo.neuralops.ai');
  const [browserScript, setBrowserScript] = useState("await page.goto(url);\nawait page.click('[data-testid=cta]');");

  const [mobileName, setMobileName] = useState('');
  const [mobilePlatform, setMobilePlatform] = useState('ios');
  const [locName, setLocName] = useState('');

  const funnelsQuery = useQuery({ queryKey: ['rum-funnels'], queryFn: fetchRUMFunnels, enabled: tab === 'RUM funnels' });
  const browserQuery = useQuery({
    queryKey: ['synthetic-browser'],
    queryFn: fetchSyntheticBrowserTests,
    enabled: tab === 'Synthetic studio',
  });
  const mobileQuery = useQuery({
    queryKey: ['synthetic-mobile'],
    queryFn: fetchSyntheticMobileTests,
    enabled: tab === 'Synthetic studio',
  });
  const locQuery = useQuery({
    queryKey: ['synthetic-private-loc'],
    queryFn: fetchSyntheticPrivateLocations,
    enabled: tab === 'Synthetic studio',
  });
  const kpiQuery = useQuery({
    queryKey: ['business-kpi-packs'],
    queryFn: fetchBusinessKPIPacks,
    enabled: tab === 'Business KPI packs',
  });

  const createFunnelMut = useMutation({
    mutationFn: () => {
      const steps = funnelSteps
        .split('\n')
        .map((line) => line.trim())
        .filter(Boolean)
        .map((line) => {
          const [name, event] = line.split('|').map((s) => s.trim());
          return { name: name || line, event: event || name || line };
        });
      return createRUMFunnel({ name: funnelName, steps });
    },
    onSuccess: () => {
      toast.success('Funnel created');
      void qc.invalidateQueries({ queryKey: ['rum-funnels'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const browserMut = useMutation({
    mutationFn: () => createSyntheticBrowserTest({ name: browserName, url: browserURL, script: browserScript }),
    onSuccess: () => {
      toast.success('Browser test created');
      setBrowserName('');
      void qc.invalidateQueries({ queryKey: ['synthetic-browser'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const mobileMut = useMutation({
    mutationFn: () => createSyntheticMobileTest({ name: mobileName, platform: mobilePlatform }),
    onSuccess: () => {
      toast.success('Mobile test created');
      setMobileName('');
      void qc.invalidateQueries({ queryKey: ['synthetic-mobile'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const locMut = useMutation({
    mutationFn: () => registerSyntheticPrivateLocation({ name: locName, region: 'on-prem' }),
    onSuccess: () => {
      toast.success('Private location registered');
      setLocName('');
      void qc.invalidateQueries({ queryKey: ['synthetic-private-loc'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const enableKpiMut = useMutation({
    mutationFn: enableBusinessKPIPack,
    onSuccess: () => {
      toast.success('KPI pack enabled');
      void qc.invalidateQueries({ queryKey: ['business-kpi-packs'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  return (
    <StitchPageShell
      title="Business observability"
      subtitle="RUM funnels, synthetic browser/mobile checks, and KPI pack library."
    >
      <div className="tab-bar" style={{ marginBottom: 24 }}>
        {TABS.map((t) => (
          <button
            key={t}
            type="button"
            className={`tab-bar__item ${tab === t ? 'tab-bar__item--active' : ''}`}
            onClick={() => setTab(t)}
            data-testid={`biz-obs-tab-${t.replace(/\s+/g, '-').toLowerCase()}`}
          >
            {t}
          </button>
        ))}
      </div>

      {tab === 'RUM funnels' && (
        <>
          <Card title="Create funnel" data-testid="rum-funnel-create">
            <Input label="Name" value={funnelName} onChange={(e) => setFunnelName(e.target.value)} />
            <Textarea
              label="Steps (name|event per line)"
              value={funnelSteps}
              onChange={(e) => setFunnelSteps(e.target.value)}
              rows={4}
              style={{ marginTop: 12 }}
            />
            <Button data-testid="rum-funnel-submit" style={{ marginTop: 12 }} onClick={() => createFunnelMut.mutate()}>
              Create funnel
            </Button>
          </Card>
          {funnelsQuery.isLoading && <LoadingState />}
          {(funnelsQuery.data ?? []).map((f) => (
            <Card key={f.id} title={f.name} style={{ marginTop: 16 }} data-testid={`rum-funnel-${f.id}`}>
              <p>
                Overall conversion: <strong>{f.overallConversionPct.toFixed(1)}%</strong>
              </p>
              <div className="funnel-steps" style={{ marginTop: 12 }}>
                {f.steps.map((s, i) => (
                  <div key={i} style={{ marginBottom: 8 }}>
                    <Badge variant="info">{s.name}</Badge> {s.count.toLocaleString()} users · {s.conversionPct.toFixed(1)}% step
                    conversion
                    <div
                      style={{
                        height: 6,
                        background: 'var(--accent)',
                        width: `${Math.max(8, s.conversionPct)}%`,
                        marginTop: 4,
                        borderRadius: 4,
                      }}
                    />
                  </div>
                ))}
              </div>
            </Card>
          ))}
        </>
      )}

      {tab === 'Synthetic studio' && (
        <>
          <Card title="Browser test" data-testid="syn-browser-create">
            <Input label="Name" value={browserName} onChange={(e) => setBrowserName(e.target.value)} />
            <Input label="URL" value={browserURL} onChange={(e) => setBrowserURL(e.target.value)} style={{ marginTop: 8 }} />
            <Textarea label="Script" value={browserScript} onChange={(e) => setBrowserScript(e.target.value)} rows={3} style={{ marginTop: 8 }} />
            <Button data-testid="syn-browser-submit" style={{ marginTop: 12 }} onClick={() => browserMut.mutate()} disabled={!browserName}>
              Create browser test
            </Button>
          </Card>
          <Card title="Mobile test" style={{ marginTop: 16 }} data-testid="syn-mobile-create">
            <Input label="Name" value={mobileName} onChange={(e) => setMobileName(e.target.value)} />
            <Input label="Platform" value={mobilePlatform} onChange={(e) => setMobilePlatform(e.target.value)} style={{ marginTop: 8 }} />
            <Button data-testid="syn-mobile-submit" style={{ marginTop: 12 }} onClick={() => mobileMut.mutate()} disabled={!mobileName}>
              Create mobile test
            </Button>
          </Card>
          <Card title="Private location" style={{ marginTop: 16 }} data-testid="syn-loc-create">
            <Input label="Name" value={locName} onChange={(e) => setLocName(e.target.value)} />
            <Button data-testid="syn-loc-submit" style={{ marginTop: 12 }} onClick={() => locMut.mutate()} disabled={!locName}>
              Register location
            </Button>
          </Card>
          {(browserQuery.data ?? []).length > 0 && (
            <Card title="Browser tests" style={{ marginTop: 24 }}>
              {(browserQuery.data ?? []).map((t) => (
                <div key={t.id} style={{ marginBottom: 8 }}>
                  <strong>{t.name}</strong> · <Badge variant={t.lastStatus === 'OK' ? 'healthy' : 'info'}>{t.lastStatus}</Badge>
                  <p className="muted">{t.url}</p>
                </div>
              ))}
            </Card>
          )}
          {(mobileQuery.data ?? []).length > 0 && (
            <Card title="Mobile tests" style={{ marginTop: 16 }}>
              {(mobileQuery.data ?? []).map((t) => (
                <div key={t.id}>
                  {t.name} ({t.platform}) · {t.bundleId}
                </div>
              ))}
            </Card>
          )}
          {(locQuery.data ?? []).length > 0 && (
            <Card title="Private locations" style={{ marginTop: 16 }}>
              {(locQuery.data ?? []).map((l) => (
                <div key={l.id}>
                  {l.name} · {l.region} · <Badge variant="healthy">{l.status}</Badge>
                </div>
              ))}
            </Card>
          )}
        </>
      )}

      {tab === 'Business KPI packs' && (
        <>
          {kpiQuery.isLoading && <LoadingState />}
          {(kpiQuery.data ?? []).map((p) => (
            <Card key={p.id} title={p.name} style={{ marginBottom: 16 }} data-testid={`kpi-pack-${p.id}`}>
              <Badge variant="info">{p.category}</Badge>{' '}
              {p.enabled ? <Badge variant="success">Enabled</Badge> : <Badge variant="warning">Disabled</Badge>}
              <p style={{ marginTop: 8 }}>{p.description}</p>
              <ul className="insight-list">
                {p.kpis.map((k) => (
                  <li key={k.key}>
                    {k.label}: target {k.target} {k.unit}
                  </li>
                ))}
              </ul>
              <p className="muted">Connectors: {p.connectors.join(', ')}</p>
              {!p.enabled && (
                <Button
                  data-testid={`kpi-enable-${p.id}`}
                  style={{ marginTop: 12 }}
                  onClick={() => enableKpiMut.mutate(p.id)}
                  disabled={enableKpiMut.isPending}
                >
                  Enable pack
                </Button>
              )}
            </Card>
          ))}
        </>
      )}
    </StitchPageShell>
  );
}
