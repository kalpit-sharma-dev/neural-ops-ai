import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { fetchAdminUsers } from '../api/observability';
import { Badge } from '../components/ui/Badge';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function SettingsUsers() {
  const { data, isLoading } = useQuery({ queryKey: ['admin-users'], queryFn: fetchAdminUsers });

  return (
    <div>
      <PageHeader title="Users" subtitle="Tenant user management" actions={<Link to="/settings">← Settings</Link>} />
      {isLoading && <LoadingState />}
      <Card title="Users">
        <table className="data-table">
          <thead><tr><th>Email</th><th>Role</th><th>Status</th></tr></thead>
          <tbody>
            {(data ?? []).map((u) => (
              <tr key={u.id}>
                <td>{u.email}</td>
                <td><Badge variant="info">{u.role}</Badge></td>
                <td>{u.active ? 'Active' : 'Disabled'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
