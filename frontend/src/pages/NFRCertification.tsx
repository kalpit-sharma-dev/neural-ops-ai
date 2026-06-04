import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import {
  fetchNFRA11y,
  fetchNFRBenchmarks,
  fetchNFRCertification,
  fetchNFRLocales,
  fetchNFRReliability,
  runNFRBenchmark,
  runNFRSlaCertification,
} from '../api/observability';
import { DataExportMenu } from '../components/ui/DataExportMenu';
import { StitchPageShell } from '../components/stitch';
import { Card } from '../components/ui/Card';
import { Badge } from '../components/ui/Badge';
import { LoadingState } from '../components/ui/PageStates';
import { Select } from '../components/ui/Select';
import { Button } from '../components/ui/Button';
import { useI18n } from '../i18n/I18nProvider';
import type { LocaleCode } from '../i18n/messages';

function passBadge(pass: boolean, t: (k: string) => string) {
  return <Badge variant={pass ? 'success' : 'critical'}>{pass ? t('nfr.pass') : t('nfr.fail')}</Badge>;
}

export default function NFRCertification() {
  const { locale, setLocale, t } = useI18n();
  const qc = useQueryClient();
  const benchmarksQuery = useQuery({ queryKey: ['nfr-benchmarks'], queryFn: fetchNFRBenchmarks });
  const reliabilityQuery = useQuery({ queryKey: ['nfr-reliability'], queryFn: fetchNFRReliability });
  const a11yQuery = useQuery({ queryKey: ['nfr-a11y'], queryFn: fetchNFRA11y });
  const localesQuery = useQuery({ queryKey: ['nfr-locales'], queryFn: fetchNFRLocales });
  const certQuery = useQuery({ queryKey: ['nfr-cert'], queryFn: fetchNFRCertification });

  const slaRunMut = useMutation({
    mutationFn: runNFRSlaCertification,
    onSuccess: (res) => {
      toast.success(res.passed ? 'SLA certification passed' : 'SLA certification failed');
      void qc.invalidateQueries({ queryKey: ['nfr-cert'] });
    },
  });

  const benchRunMut = useMutation({
    mutationFn: runNFRBenchmark,
    onSuccess: (res) => {
      toast.success(`Benchmark ${res.passed ? 'passed' : 'failed'} · p95 ${res.p95Ms}ms`);
      void qc.invalidateQueries({ queryKey: ['nfr-benchmarks'] });
    },
  });

  const loading =
    benchmarksQuery.isLoading ||
    reliabilityQuery.isLoading ||
    a11yQuery.isLoading ||
    localesQuery.isLoading ||
    certQuery.isLoading;

  return (
    <StitchPageShell
      title={t('nfr.title')}
      subtitle={t('nfr.subtitle')}
      actions={
        certQuery.data ? (
          <DataExportMenu
            getData={() => [
              {
                certification: certQuery.data,
                benchmarks: benchmarksQuery.data,
                reliability: reliabilityQuery.data,
                a11y: a11yQuery.data,
              },
            ]}
            filenamePrefix="nfr-report"
            formats={['json']}
            label="Download report"
          />
        ) : null
      }
    >
      <div style={{ maxWidth: 220, marginBottom: 16 }} data-testid="nfr-locale-switcher">
        <Select label={t('nfr.locale')} value={locale} onChange={(e) => setLocale(e.target.value as LocaleCode)}>
          <option value="en">English</option>
          <option value="es">Español</option>
          <option value="de">Deutsch</option>
        </Select>
      </div>

      {loading && <LoadingState />}

      {certQuery.data && (
        <Card title={t('nfr.certification')} data-testid="nfr-cert-card">
          <p>
            Version {certQuery.data.version} · {passBadge(certQuery.data.overallPass, t)}
          </p>
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginTop: 12 }}>
            <Button
              variant="secondary"
              size="sm"
              data-testid="nfr-run-sla"
              disabled={slaRunMut.isPending}
              onClick={() => slaRunMut.mutate()}
            >
              {t('nfr.runSla')}
            </Button>
            <Button
              variant="secondary"
              size="sm"
              data-testid="nfr-run-benchmark"
              disabled={benchRunMut.isPending}
              onClick={() => benchRunMut.mutate()}
            >
              {t('nfr.runBench')}
            </Button>
          </div>
          <ul className="insight-list" style={{ marginTop: 8 }}>
            {certQuery.data.evidenceUris.map((uri) => (
              <li key={uri}>
                <code>{uri}</code>
              </li>
            ))}
          </ul>
        </Card>
      )}

      <Card title={t('nfr.benchmarks')} style={{ marginTop: 24 }} data-testid="nfr-benchmarks-card">
        <table className="data-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Target</th>
              <th>Actual</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {(benchmarksQuery.data ?? []).map((b) => (
              <tr key={b.id}>
                <td>{b.name}</td>
                <td>{b.target}</td>
                <td>
                  {b.actual} {b.unit}
                </td>
                <td>{passBadge(b.pass, t)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>

      <Card title={t('nfr.reliability')} style={{ marginTop: 24 }} data-testid="nfr-reliability-card">
        {(reliabilityQuery.data ?? []).map((d) => (
          <div key={d.id} style={{ marginBottom: 12 }}>
            <strong>{d.name}</strong> ({d.type}) · {d.region} — RTO {d.rtoSeconds}s / SLO {d.rtoSloSeconds}s{' '}
            {passBadge(d.pass, t)}
          </div>
        ))}
      </Card>

      {a11yQuery.data && (
        <Card title={t('nfr.accessibility')} style={{ marginTop: 24 }} data-testid="nfr-a11y-card">
          <p>
            {a11yQuery.data.standard} {a11yQuery.data.level} · {a11yQuery.data.pagesAudited} pages · critical{' '}
            {a11yQuery.data.violationsCritical} · serious {a11yQuery.data.violationsSerious} ·{' '}
            {passBadge(a11yQuery.data.pass, t)}
          </p>
        </Card>
      )}

      <Card title={t('nfr.i18n')} style={{ marginTop: 24 }} data-testid="nfr-i18n-card">
        <table className="data-table">
          <thead>
            <tr>
              <th>Locale</th>
              <th>Coverage</th>
              <th>Enabled</th>
            </tr>
          </thead>
          <tbody>
            {(localesQuery.data ?? []).map((l) => (
              <tr key={l.code}>
                <td>
                  {l.name} ({l.code})
                </td>
                <td>{l.coveragePct}%</td>
                <td>{l.enabled ? passBadge(true, t) : <Badge variant="warning">Off</Badge>}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </StitchPageShell>
  );
}
