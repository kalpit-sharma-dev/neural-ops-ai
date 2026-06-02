import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { fetchSSOConfig, updateSSOConfig } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { LoadingState } from '../components/ui/PageStates';
import { StitchPageShell, SettingsBreadcrumb } from '../components/stitch';

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
    <StitchPageShell
      title="SSO / IdP"
      subtitle="OpenID Connect provider settings (live reload)"
      breadcrumb={<SettingsBreadcrumb page="SSO / IdP" />}
    >
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
          <Input name="provider" label="Provider" defaultValue={data.provider} />
          <Input name="metadataUrl" label="Metadata URL" defaultValue={data.metadataUrl ?? ''} />
          <Input name="clientId" label="Client ID" defaultValue={data.clientId ?? ''} />
          <Input name="issuerUrl" label="Issuer URL" defaultValue={data.issuerUrl ?? ''} />
          <Input
            name="clientSecret"
            label="Client secret"
            type="password"
            hint={data.clientSecretSet ? 'Leave blank to keep existing' : undefined}
            placeholder={data.clientSecretSet ? '••••••••' : ''}
          />
          <Button type="submit" variant="primary">Save & reload</Button>
        </form>
      </Card>
    </StitchPageShell>
  );
}
