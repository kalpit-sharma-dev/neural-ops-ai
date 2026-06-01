import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { fetchIntegrations } from '../api/observability';
import { Badge } from '../components/ui/Badge';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

const EXTENSIONS = [
  { id: 'ext-grafana', name: 'Grafana panels', type: 'visualization', status: 'installed' },
  { id: 'ext-jira', name: 'Jira incident sync', type: 'integration', status: 'available' },
  { id: 'ext-pagerduty', name: 'PagerDuty on-call', type: 'integration', status: 'installed' },
];

export default function Marketplace() {
  const { data, isLoading } = useQuery({ queryKey: ['integrations'], queryFn: fetchIntegrations });

  return (
    <div>
      <PageHeader title="Marketplace" subtitle="Extensions and integration catalog" actions={<Link to="/integrations">Manage connections →</Link>} />
      {isLoading && <LoadingState />}
      <div className="settings-grid">
        {EXTENSIONS.map((ext) => (
          <Card key={ext.id} title={ext.name}>
            <Badge variant={ext.status === 'installed' ? 'healthy' : 'info'}>{ext.status}</Badge>
            <p className="muted">{ext.type}</p>
          </Card>
        ))}
        {(data ?? []).map((i) => (
          <Card key={i.id} title={i.name}>
            <Badge variant={i.connected ? 'healthy' : 'info'}>{i.connected ? 'Connected' : 'Available'}</Badge>
            <p className="muted">{i.type}</p>
            <Link to="/integrations">Configure</Link>
          </Card>
        ))}
      </div>
    </div>
  );
}
