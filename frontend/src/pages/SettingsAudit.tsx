import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { fetchAuditLog } from '../api/observability';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function SettingsAudit() {
  const { data, isLoading } = useQuery({ queryKey: ['admin-audit'], queryFn: fetchAuditLog });

  return (
    <div>
      <PageHeader title="Audit Log" subtitle="Configuration and security events" actions={<Link to="/settings">← Settings</Link>} />
      {isLoading && <LoadingState />}
      <Card title="Recent events">
        <table className="data-table">
          <thead><tr><th>Time</th><th>User</th><th>Action</th><th>Resource</th><th>Detail</th></tr></thead>
          <tbody>
            {(data ?? []).map((e) => (
              <tr key={e.id}>
                <td>{new Date(e.timestamp).toLocaleString()}</td>
                <td>{e.userId}</td>
                <td>{e.action}</td>
                <td>{e.resource}</td>
                <td>{e.detail}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
