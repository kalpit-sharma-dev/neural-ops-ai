import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { UserPlus } from 'lucide-react';
import {
  createAdminUser,
  fetchAdminUsers,
  updateAdminUser,
  USER_ROLES,
  type AdminUser,
} from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function roleBadgeVariant(role: string): 'critical' | 'info' | 'healthy' {
  const r = role.toUpperCase();
  if (r === 'ADMIN') return 'critical';
  if (r === 'READONLY') return 'healthy';
  return 'info';
}

export default function SettingsUsers() {
  const queryClient = useQueryClient();
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['admin-users'],
    queryFn: fetchAdminUsers,
  });

  const [email, setEmail] = useState('');
  const [role, setRole] = useState<string>('READONLY');
  const [emailError, setEmailError] = useState('');

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['admin-users'] });

  const inviteMut = useMutation({
    mutationFn: () => createAdminUser({ email: email.trim(), role }),
    onSuccess: () => {
      toast.success(`Invited ${email.trim()}`);
      setEmail('');
      setRole('READONLY');
      void invalidate();
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const roleMut = useMutation({
    mutationFn: ({ id, nextRole }: { id: string; nextRole: string }) =>
      updateAdminUser(id, { role: nextRole }),
    onSuccess: () => {
      toast.success('Role updated');
      void invalidate();
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const activeMut = useMutation({
    mutationFn: ({ id, active }: { id: string; active: boolean }) => updateAdminUser(id, { active }),
    onSuccess: (_d, vars) => {
      toast.success(vars.active ? 'User activated' : 'User deactivated');
      void invalidate();
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const submitInvite = () => {
    const trimmed = email.trim();
    if (!EMAIL_RE.test(trimmed)) {
      setEmailError('Enter a valid email address.');
      return;
    }
    setEmailError('');
    inviteMut.mutate();
  };

  const users = data ?? [];
  const mutating = roleMut.isPending || activeMut.isPending;

  return (
    <StitchPageShell
      title="Users"
      subtitle="Invite teammates and manage roles and access"
    >
      <Card title="Invite user">
        <div className="users-invite">
          <Input
            label="Email"
            type="email"
            placeholder="teammate@company.com"
            value={email}
            error={emailError}
            data-testid="users-invite-email"
            onChange={(e) => setEmail(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') submitInvite();
            }}
          />
          <label className="ui-select users-invite__role">
            <span className="ui-select__label">Role</span>
            <select
              className="ui-select__control"
              value={role}
              onChange={(e) => setRole(e.target.value)}
              aria-label="Invite role"
            >
              {USER_ROLES.map((r) => (
                <option key={r} value={r}>
                  {r}
                </option>
              ))}
            </select>
          </label>
          <Button
            variant="primary"
            onClick={submitInvite}
            disabled={inviteMut.isPending || !email.trim()}
            data-testid="users-invite-btn"
          >
            <UserPlus size={15} aria-hidden /> Invite
          </Button>
        </div>
      </Card>

      {isLoading && <LoadingState />}
      {error && <ErrorState message={getApiErrorMessage(error)} onRetry={() => refetch()} />}

      {!isLoading && !error && (
        <Card title={`Users (${users.length})`}>
          <table className="data-table">
            <thead>
              <tr>
                <th>Email</th>
                <th>Role</th>
                <th>Status</th>
                <th aria-label="Actions" />
              </tr>
            </thead>
            <tbody>
              {users.map((u: AdminUser) => (
                <tr key={u.id} className={u.active ? '' : 'data-table__row--muted'} data-testid={`users-row-${u.id}`}>
                  <td>{u.email}</td>
                  <td>
                    <div className="users-row-role">
                      <Badge variant={roleBadgeVariant(u.role)}>{u.role.toUpperCase()}</Badge>
                      <select
                        className="ui-select__control users-row-role__select"
                        value={u.role.toUpperCase()}
                        disabled={mutating}
                        onChange={(e) => roleMut.mutate({ id: u.id, nextRole: e.target.value })}
                        aria-label={`Role for ${u.email}`}
                      >
                        {USER_ROLES.map((r) => (
                          <option key={r} value={r}>
                            {r}
                          </option>
                        ))}
                      </select>
                    </div>
                  </td>
                  <td>
                    <Badge variant={u.active ? 'healthy' : 'critical'}>{u.active ? 'Active' : 'Disabled'}</Badge>
                  </td>
                  <td style={{ textAlign: 'right' }}>
                    <Button
                      variant={u.active ? 'ghost' : 'secondary'}
                      size="sm"
                      disabled={mutating}
                      onClick={() => activeMut.mutate({ id: u.id, active: !u.active })}
                    >
                      {u.active ? 'Deactivate' : 'Activate'}
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}
    </StitchPageShell>
  );
}
