import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import toast from 'react-hot-toast';
import { Link } from '@tanstack/react-router';
import { createAPIKey, fetchAPIKeys } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Modal } from '../components/ui/Modal';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function SettingsApiKeys() {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({ queryKey: ['admin-api-keys'], queryFn: fetchAPIKeys });
  const [name, setName] = useState('');
  const [secret, setSecret] = useState<string | null>(null);

  const createMut = useMutation({
    mutationFn: () => createAPIKey({ name, role: 'developer' }),
    onSuccess: (result) => {
      toast.success('API key created — copy the secret now');
      setSecret(result.secret);
      setName('');
      void queryClient.invalidateQueries({ queryKey: ['admin-api-keys'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  return (
    <div>
      <PageHeader
        title="API Keys"
        subtitle="Programmatic access tokens"
        actions={
          <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <input placeholder="Key name" value={name} onChange={(e) => setName(e.target.value)} />
            <Button variant="primary" disabled={!name} onClick={() => createMut.mutate()}>
              Create key
            </Button>
            <Link to="/settings">← Settings</Link>
          </div>
        }
      />
      {isLoading && <LoadingState />}
      <Card title="Keys">
        <table className="data-table">
          <thead><tr><th>Name</th><th>Prefix</th><th>Scopes</th><th>Created</th></tr></thead>
          <tbody>
            {(data ?? []).map((k) => (
              <tr key={k.id}>
                <td>{k.name}</td>
                <td><code>{k.prefix}…</code></td>
                <td>{k.scopes.join(', ')}</td>
                <td>{new Date(k.createdAt).toLocaleDateString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>

      <Modal open={secret != null} onOpenChange={(open) => !open && setSecret(null)} title="API key secret">
        <p className="muted">This secret is shown once. Store it securely.</p>
        <code style={{ display: 'block', padding: 12, wordBreak: 'break-all' }}>{secret}</code>
      </Modal>
    </div>
  );
}
