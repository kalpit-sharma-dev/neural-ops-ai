import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { fetchAttacks, fetchVulnerabilities } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Badge } from '../components/ui/Badge';
import { Card } from '../components/ui/Card';
import { DomainEmptyState } from '../components/ui/DomainEmptyState';
import { EmptyState } from '../components/ui/EmptyState';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';

export default function Security() {
  const vulnQuery = useQuery({ queryKey: ['security-vulns'], queryFn: fetchVulnerabilities });
  const attackQuery = useQuery({ queryKey: ['security-attacks'], queryFn: fetchAttacks });

  const error = vulnQuery.error ?? attackQuery.error;

  return (
    <StitchPageShell title="Application Security" subtitle="Runtime vulnerabilities and attack detection">
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
    </StitchPageShell>
  );
}
