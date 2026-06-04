import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import {
  exportEvents,
  exportToBI,
  exportToWarehouse,
  fetchABACPolicies,
  fetchBranding,
  fetchDataResidency,
  fetchMSPTenants,
  createMSPTenant,
  fetchMultiRegionStatus,
  updateABACPolicies,
  updateBranding,
  updateDataResidency,
  type ExportJob,
} from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { StitchPageShell } from '../components/stitch';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { Badge } from '../components/ui/Badge';
import { LoadingState } from '../components/ui/PageStates';
import { DataExportMenu } from '../components/ui/DataExportMenu';

const TABS = ['ABAC', 'Residency', 'Multi-region', 'Branding', 'MSP tenants', 'Exports'] as const;

export default function EnterpriseGovernance() {
  const qc = useQueryClient();
  const [tab, setTab] = useState<(typeof TABS)[number]>('ABAC');
  const [exportDest, setExportDest] = useState('s3://neuralops-exports/prod');
  const [lastExport, setLastExport] = useState<ExportJob | null>(null);

  const abacQuery = useQuery({ queryKey: ['abac'], queryFn: fetchABACPolicies, enabled: tab === 'ABAC' });
  const residencyQuery = useQuery({ queryKey: ['residency'], queryFn: fetchDataResidency, enabled: tab === 'Residency' });
  const regionsQuery = useQuery({ queryKey: ['multi-region'], queryFn: fetchMultiRegionStatus, enabled: tab === 'Multi-region' });
  const brandingQuery = useQuery({ queryKey: ['branding'], queryFn: fetchBranding, enabled: tab === 'Branding' });
  const mspQuery = useQuery({ queryKey: ['msp-tenants'], queryFn: fetchMSPTenants, enabled: tab === 'MSP tenants' });
  const [mspName, setMspName] = useState('');
  const [mspSlug, setMspSlug] = useState('');
  const [mspRegion, setMspRegion] = useState('ap-south-1');

  const createMspMut = useMutation({
    mutationFn: () =>
      createMSPTenant({
        name: mspName.trim(),
        slug: mspSlug.trim() || mspName.trim().toLowerCase().replace(/\s+/g, '-'),
        plan: 'enterprise',
        region: mspRegion,
      }),
    onSuccess: () => {
      toast.success('MSP tenant registered');
      setMspName('');
      setMspSlug('');
      void qc.invalidateQueries({ queryKey: ['msp-tenants'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const [abacEnabled, setAbacEnabled] = useState(true);
  const [primaryRegion, setPrimaryRegion] = useState('us-east-1');
  const [productName, setProductName] = useState('NeuralOps');

  const saveAbacMut = useMutation({
    mutationFn: () =>
      updateABACPolicies({
        enabled: abacEnabled,
        rules: abacQuery.data?.rules ?? [],
        updatedAt: abacQuery.data?.updatedAt ?? new Date().toISOString(),
      }),
    onSuccess: () => {
      toast.success('ABAC policies saved');
      void qc.invalidateQueries({ queryKey: ['abac'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const saveResidencyMut = useMutation({
    mutationFn: () =>
      updateDataResidency({
        primaryRegion,
        allowedRegions: residencyQuery.data?.allowedRegions ?? [primaryRegion],
        piiStorageRegion: residencyQuery.data?.piiStorageRegion ?? primaryRegion,
        crossBorderDenied: residencyQuery.data?.crossBorderDenied ?? true,
        updatedAt: residencyQuery.data?.updatedAt ?? new Date().toISOString(),
      }),
    onSuccess: () => {
      toast.success('Residency policy saved');
      void qc.invalidateQueries({ queryKey: ['residency'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const saveBrandingMut = useMutation({
    mutationFn: () =>
      updateBranding({
        productName,
        logoUrl: brandingQuery.data?.logoUrl ?? '/brand/logo.svg',
        primaryColor: brandingQuery.data?.primaryColor ?? '#2563eb',
        accentColor: brandingQuery.data?.accentColor ?? '#0ea5e9',
        supportEmail: brandingQuery.data?.supportEmail ?? 'support@neuralops.ai',
        customDomain: brandingQuery.data?.customDomain ?? '',
        updatedAt: brandingQuery.data?.updatedAt ?? new Date().toISOString(),
      }),
    onSuccess: () => {
      toast.success('Branding saved');
      void qc.invalidateQueries({ queryKey: ['branding'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const exportOpts = {
    onSuccess: (job: ExportJob) => {
      setLastExport(job);
      toast.success(`Export ${job.status}`);
    },
    onError: (e: unknown) => toast.error(getApiErrorMessage(e)),
  };

  const warehouseMut = useMutation({
    mutationFn: () => exportToWarehouse(exportDest),
    ...exportOpts,
  });
  const biMut = useMutation({
    mutationFn: () => exportToBI(exportDest),
    ...exportOpts,
  });
  const eventsMut = useMutation({
    mutationFn: () => exportEvents(exportDest),
    ...exportOpts,
  });

  useEffect(() => {
    if (abacQuery.data) setAbacEnabled(abacQuery.data.enabled);
  }, [abacQuery.data]);

  useEffect(() => {
    if (residencyQuery.data) setPrimaryRegion(residencyQuery.data.primaryRegion);
  }, [residencyQuery.data]);

  useEffect(() => {
    if (brandingQuery.data) setProductName(brandingQuery.data.productName);
  }, [brandingQuery.data]);

  return (
    <StitchPageShell
      title="Enterprise governance"
      subtitle="ABAC, residency, MSP white-label, and export connectors."
      actions={
        <DataExportMenu
          getData={() => {
            switch (tab) {
              case 'ABAC':
                return abacQuery.data ? [abacQuery.data] : [];
              case 'Residency':
                return residencyQuery.data ? [residencyQuery.data] : [];
              case 'Multi-region':
                return regionsQuery.data ? [regionsQuery.data] : [];
              case 'Branding':
                return brandingQuery.data ? [brandingQuery.data] : [];
              case 'MSP tenants':
                return mspQuery.data ?? [];
              case 'Exports':
                return lastExport ? [lastExport] : [];
              default:
                return [];
            }
          }}
          filenamePrefix={`governance-${tab.replace(/\s+/g, '-').toLowerCase()}`}
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
            data-testid={`gov-tab-${t.replace(/\s+/g, '-').toLowerCase()}`}
          >
            {t}
          </button>
        ))}
      </div>

      {tab === 'ABAC' && (
        <Card title="ABAC policy evaluator" data-testid="gov-abac-card">
          {abacQuery.isLoading && <LoadingState />}
          {abacQuery.data && (
            <>
              <label className="checkbox-row">
                <input type="checkbox" checked={abacEnabled} onChange={(e) => setAbacEnabled(e.target.checked)} />
                Enable ABAC (layered on RBAC)
              </label>
              <ul className="insight-list" style={{ marginTop: 12 }}>
                {abacQuery.data.rules.map((r) => (
                  <li key={r.id}>
                    <Badge variant={r.effect === 'allow' ? 'success' : 'critical'}>{r.effect}</Badge> {r.action} on{' '}
                    {r.resource} when <code>{r.condition}</code>
                  </li>
                ))}
              </ul>
              <Button data-testid="gov-abac-save" style={{ marginTop: 12 }} onClick={() => saveAbacMut.mutate()}>
                Save policies
              </Button>
            </>
          )}
        </Card>
      )}

      {tab === 'Residency' && (
        <Card title="Data residency" data-testid="gov-residency-card">
          {residencyQuery.data && (
            <>
              <Input label="Primary region" value={primaryRegion} onChange={(e) => setPrimaryRegion(e.target.value)} />
              <p className="muted" style={{ marginTop: 8 }}>
                Allowed: {residencyQuery.data.allowedRegions.join(', ')} · PII: {residencyQuery.data.piiStorageRegion}
                {residencyQuery.data.crossBorderDenied ? ' · cross-border denied' : ''}
              </p>
              <Button data-testid="gov-residency-save" style={{ marginTop: 12 }} onClick={() => saveResidencyMut.mutate()}>
                Save residency
              </Button>
            </>
          )}
        </Card>
      )}

      {tab === 'Multi-region' && (
        <Card title="Active-active regions" data-testid="gov-regions-card">
          {regionsQuery.isLoading && <LoadingState />}
          {regionsQuery.data && (
            <>
              <p>
                <Badge variant={regionsQuery.data.enabled ? 'healthy' : 'info'}>
                  {regionsQuery.data.enabled ? 'Multi-region enabled' : 'Single region'}
                </Badge>{' '}
                Local: <strong>{regionsQuery.data.localRegion}</strong>
                {regionsQuery.data.crossBorderOk ? ' · cross-border OK' : ' · cross-border restricted'}
              </p>
              <table className="data-table" style={{ marginTop: 12 }}>
                <thead>
                  <tr>
                    <th>Region</th>
                    <th>Gateway</th>
                    <th>Status</th>
                    <th>Primary</th>
                  </tr>
                </thead>
                <tbody>
                  {regionsQuery.data.peers.map((p) => (
                    <tr key={p.region}>
                      <td>{p.region}</td>
                      <td>
                        <code>{p.gatewayUrl}</code>
                      </td>
                      <td>
                        <Badge variant={p.status === 'healthy' ? 'healthy' : 'warning'}>{p.status}</Badge>
                      </td>
                      <td>{p.isPrimary ? 'Yes' : '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </>
          )}
        </Card>
      )}

      {tab === 'Branding' && (
        <Card title="White-label branding" data-testid="gov-branding-card">
          {brandingQuery.data && (
            <>
              <Input label="Product name" value={productName} onChange={(e) => setProductName(e.target.value)} />
              <p className="muted" style={{ marginTop: 8 }}>
                Domain: {brandingQuery.data.customDomain} · Colors: {brandingQuery.data.primaryColor} /{' '}
                {brandingQuery.data.accentColor}
              </p>
              <Button data-testid="gov-branding-save" style={{ marginTop: 12 }} onClick={() => saveBrandingMut.mutate()}>
                Save branding
              </Button>
            </>
          )}
        </Card>
      )}

      {tab === 'MSP tenants' && (
        <Card title="MSP tenant console" data-testid="gov-msp-card">
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginBottom: 16 }}>
            <Input label="Tenant name" value={mspName} onChange={(e) => setMspName(e.target.value)} />
            <Input label="Slug" value={mspSlug} onChange={(e) => setMspSlug(e.target.value)} placeholder="auto from name" />
            <Input label="Region" value={mspRegion} onChange={(e) => setMspRegion(e.target.value)} />
            <Button
              data-testid="gov-msp-create"
              disabled={!mspName.trim() || createMspMut.isPending}
              onClick={() => createMspMut.mutate()}
            >
              Register tenant
            </Button>
          </div>
          <table className="data-table">
            <thead>
              <tr>
                <th>Tenant</th>
                <th>Plan</th>
                <th>Status</th>
                <th>Users</th>
                <th>Region</th>
              </tr>
            </thead>
            <tbody>
              {(mspQuery.data ?? []).map((t) => (
                <tr key={t.id} data-testid={`msp-tenant-${t.slug}`}>
                  <td>{t.name}</td>
                  <td>{t.plan}</td>
                  <td>
                    <Badge variant={t.status === 'active' ? 'healthy' : 'warning'}>{t.status}</Badge>
                  </td>
                  <td>{t.userCount}</td>
                  <td>{t.region}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}

      {tab === 'Exports' && (
        <Card title="Export connectors" data-testid="gov-exports-card">
          <Input label="Destination URI" value={exportDest} onChange={(e) => setExportDest(e.target.value)} />
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginTop: 12 }}>
            <Button data-testid="gov-export-warehouse" onClick={() => warehouseMut.mutate()}>
              Warehouse export
            </Button>
            <Button data-testid="gov-export-bi" onClick={() => biMut.mutate()}>
              BI export
            </Button>
            <Button data-testid="gov-export-events" onClick={() => eventsMut.mutate()}>
              Events export
            </Button>
          </div>
          {lastExport && (
            <p style={{ marginTop: 12 }}>
              Job <code>{lastExport.id}</code> — {lastExport.status} · {lastExport.rowsExported.toLocaleString()} rows ·{' '}
              {lastExport.message}
            </p>
          )}
          <p className="muted" style={{ marginTop: 16 }}>
            Automate with CLI: <code>neuralops export -type warehouse -destination {exportDest}</code>
          </p>
        </Card>
      )}
    </StitchPageShell>
  );
}
