import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { fetchTenantPolicies, updateTenantPolicies } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/PageStates';
import { StitchPageShell, SettingsBreadcrumb } from '../components/stitch';

export default function SettingsPolicies() {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({ queryKey: ['tenant-policies'], queryFn: fetchTenantPolicies });

  const saveMut = useMutation({
    mutationFn: updateTenantPolicies,
    onSuccess: () => {
      toast.success('Tenant policies saved');
      void queryClient.invalidateQueries({ queryKey: ['tenant-policies'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  if (isLoading || !data) return <LoadingState />;

  return (
    <StitchPageShell
      title="Tenant policies"
      subtitle="Retention and ingestion limits"
      breadcrumb={<SettingsBreadcrumb page="Tenant policies" />}
    >
      <Card title="Data governance">
        <form
          className="form-stack"
          onSubmit={(e) => {
            e.preventDefault();
            const fd = new FormData(e.currentTarget);
            saveMut.mutate({
              logRetentionDays: Number(fd.get('logRetentionDays')),
              ingestionRateLimit: Number(fd.get('ingestionRateLimit')),
            });
          }}
        >
          <label>Log retention (days)<input name="logRetentionDays" type="number" defaultValue={data.logRetentionDays} /></label>
          <label>Ingestion rate limit (events/min)<input name="ingestionRateLimit" type="number" defaultValue={data.ingestionRateLimit} /></label>
          <Button type="submit" variant="primary">Save</Button>
        </form>
      </Card>
    </StitchPageShell>
  );
}
