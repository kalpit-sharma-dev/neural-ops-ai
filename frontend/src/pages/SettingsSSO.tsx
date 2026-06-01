import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import toast from 'react-hot-toast';
import { fetchSSOConfig, updateSSOConfig } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function SettingsSSO() {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({ queryKey: ['admin-sso'], queryFn: fetchSSOConfig });

  const saveMut = useMutation({
    mutationFn: updateSSOConfig,
    onSuccess: () => {
      toast.success('SSO configuration saved — gateway will reload OIDC on next login');
      void queryClient.invalidateQueries({ queryKey: ['admin-sso'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  if (isLoading || !data) return <LoadingState />;

  return (
    <div>
      <PageHeader title="SSO / IdP" subtitle="OpenID Connect provider settings (live reload)" actions={<Link to="/settings">← Settings</Link>} />
      <Card title="Identity provider">
        <form
          className="form-stack"
          onSubmit={(e) => {
            e.preventDefault();
            const fd = new FormData(e.currentTarget);
            saveMut.mutate({
              provider: String(fd.get('provider') ?? 'oidc'),
              metadataUrl: String(fd.get('metadataUrl') ?? ''),
              clientId: String(fd.get('clientId') ?? ''),
              issuerUrl: String(fd.get('issuerUrl') ?? ''),
              clientSecret: String(fd.get('clientSecret') ?? ''),
            });
          }}
        >
          <label>Provider<input name="provider" defaultValue={data.provider} /></label>
          <label>Metadata URL<input name="metadataUrl" defaultValue={data.metadataUrl ?? ''} /></label>
          <label>Client ID<input name="clientId" defaultValue={data.clientId ?? ''} /></label>
          <label>Issuer URL<input name="issuerUrl" defaultValue={data.issuerUrl ?? ''} /></label>
          <label>
            Client secret {data.clientSecretSet && <span className="muted">(leave blank to keep existing)</span>}
            <input name="clientSecret" type="password" placeholder={data.clientSecretSet ? '••••••••' : ''} />
          </label>
          <Button type="submit" variant="primary">Save & reload</Button>
        </form>
      </Card>
    </div>
  );
}
