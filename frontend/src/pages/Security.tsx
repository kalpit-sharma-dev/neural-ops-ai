import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { fetchAttacks, fetchVulnerabilities } from '../api/observability';
import { Badge } from '../components/ui/Badge';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function Security() {
  const vulnQuery = useQuery({ queryKey: ['security-vulns'], queryFn: fetchVulnerabilities });
  const attackQuery = useQuery({ queryKey: ['security-attacks'], queryFn: fetchAttacks });

  return (
    <div>
      <PageHeader title="Application Security" subtitle="Runtime vulnerabilities and attack detection" />
      {(vulnQuery.isLoading || attackQuery.isLoading) && <LoadingState />}
      <div className="dashboard-row-2">
        <Card title="Vulnerabilities">
          {(vulnQuery.data ?? []).map((v) => (
            <div key={v.id} className="list-row">
              <Badge variant={v.severity === 'HIGH' ? 'critical' : 'warning'}>{v.severity}</Badge>
              <span>{v.cve}</span>
              <span>{v.service}</span>
            </div>
          ))}
        </Card>
        <Card title="Attacks">
          {(attackQuery.data ?? []).map((a) => (
            <Link key={a.id} to="/security/attacks/$id" params={{ id: a.id }} className="list-row">
              <Badge variant={a.blocked ? 'healthy' : 'critical'}>{a.blocked ? 'Blocked' : 'Open'}</Badge>
              <span>{a.type}</span>
              <span>{a.sourceIp} → {a.service}</span>
            </Link>
          ))}
        </Card>
      </div>
    </div>
  );
}
