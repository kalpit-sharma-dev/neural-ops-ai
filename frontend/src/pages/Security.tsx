import { useMutation, useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { fetchAttacks, fetchVulnerabilities, fetchSecurityFindings, fetchSecurityPosture, exportSecurityToSIEM } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { DomainEmptyState } from '../components/ui/DomainEmptyState';
import { EmptyState } from '../components/ui/EmptyState';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { DataExportMenu } from '../components/ui/DataExportMenu';
import { StitchPageShell } from '../components/stitch';
import { useI18n } from '../i18n/I18nProvider';

export default function Security() {
  const { t } = useI18n();
  const vulnQuery = useQuery({ queryKey: ['security-vulns'], queryFn: fetchVulnerabilities });
  const attackQuery = useQuery({ queryKey: ['security-attacks'], queryFn: fetchAttacks });
  const findingsQuery = useQuery({ queryKey: ['security-findings'], queryFn: fetchSecurityFindings });
  const postureQuery = useQuery({ queryKey: ['security-posture'], queryFn: fetchSecurityPosture });
  const siemExportMut = useMutation({
    mutationFn: () => exportSecurityToSIEM('splunk', 'https://splunk.local/hec'),
  });

  const error = vulnQuery.error ?? attackQuery.error ?? findingsQuery.error ?? postureQuery.error;

  return (
    <StitchPageShell
      title="Application Security"
      subtitle="Runtime vulnerabilities and attack detection"
      actions={
        <DataExportMenu
          getData={() => [
            ...(vulnQuery.data ?? []).map((v) => ({ recordType: 'vulnerability', ...v })),
            ...(attackQuery.data ?? []).map((a) => ({ recordType: 'attack', ...a })),
            ...(findingsQuery.data ?? []).map((f) => ({ recordType: 'finding', ...f })),
          ]}
          filenamePrefix="security"
          disabled={
            !vulnQuery.data?.length && !attackQuery.data?.length && !findingsQuery.data?.length
          }
        />
      }
    >
      {(vulnQuery.isLoading || attackQuery.isLoading) && <LoadingState />}
      {error && (
        <ErrorState
          message={getApiErrorMessage(error)}
          onRetry={() => {
            void vulnQuery.refetch();
            void attackQuery.refetch();
          }}
        />
      )}
      {!error && (
        <div className="dashboard-row-2">
          <Card title="Vulnerabilities">
            {(vulnQuery.data ?? []).map((v) => (
              <div key={v.id} className="list-row">
                <Badge variant={v.severity === 'HIGH' ? 'critical' : 'warning'}>{v.severity}</Badge>
                <span>{v.cve}</span>
                <span>{v.service}</span>
              </div>
            ))}
            {vulnQuery.isFetched && (vulnQuery.data?.length ?? 0) === 0 && (
              <EmptyState
                title="No vulnerabilities"
                description="No runtime vulnerabilities detected for the current scope."
              />
            )}
          </Card>
          <Card title="Attacks">
            {(attackQuery.data ?? []).map((a) => (
              <Link key={a.id} to="/security/attacks/$id" params={{ id: a.id }} className="list-row" data-testid={`security-attack-link-${a.id}`}>
                <Badge variant={a.blocked ? 'healthy' : 'critical'}>{a.blocked ? 'Blocked' : 'Open'}</Badge>
                <span>{a.type}</span>
                <span>{a.sourceIp} → {a.service}</span>
              </Link>
            ))}
            {attackQuery.isFetched && (attackQuery.data?.length ?? 0) === 0 && (
              <DomainEmptyState domain="security" />
            )}
          </Card>
        </div>
      )}
      {!error && (
        <div className="dashboard-row-2">
          <Card
            title="Findings"
            action={<Button size="sm" variant="secondary" onClick={() => siemExportMut.mutate()} disabled={siemExportMut.isPending}>{t('page.security.exportSiem')}</Button>}
          >
            {(findingsQuery.data ?? []).map((f) => (
              <Link
                key={f.id}
                to="/security/findings/$id"
                params={{ id: f.id }}
                className="list-row"
                data-testid={`security-finding-link-${f.id}`}
              >
                <span>
                  <strong>{f.title}</strong>
                  <span className="muted" style={{ marginLeft: 8 }}>{f.category} · {f.service ?? f.asset ?? 'n/a'}</span>
                </span>
                <span style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                  <Badge variant={f.severity === 'HIGH' ? 'critical' : f.severity === 'MEDIUM' ? 'warning' : 'info'}>{f.severity}</Badge>
                  <Badge variant={f.status === 'OPEN' ? 'critical' : 'healthy'}>{f.status}</Badge>
                  {f.incidentId && <span className="muted">incident linked</span>}
                </span>
              </Link>
            ))}
          </Card>
          <Card title="Cloud posture">
            {(postureQuery.data ?? []).map((p) => (
              <div key={p.id} className="list-row">
                <span>
                  <strong>{p.name}</strong>
                  <span className="muted" style={{ marginLeft: 8 }}>{p.provider} · {p.resource}</span>
                </span>
                <Badge variant={p.status === 'pass' ? 'healthy' : p.status === 'warn' ? 'warning' : 'critical'}>{p.status}</Badge>
              </div>
            ))}
          </Card>
        </div>
      )}
    </StitchPageShell>
  );
}
