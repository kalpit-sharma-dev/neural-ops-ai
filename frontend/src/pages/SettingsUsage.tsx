import { useQuery } from '@tanstack/react-query';
import { fetchUsage } from '../api/observability';
import { MetricCard } from '../components/ui/MetricCard';
import { LoadingState } from '../components/ui/PageStates';
import { StitchPageShell, SettingsBreadcrumb, KpiRow } from '../components/stitch';

export default function SettingsUsage() {
  const { data, isLoading } = useQuery({ queryKey: ['admin-usage'], queryFn: fetchUsage });

  return (
    <StitchPageShell
      title="Usage"
      subtitle="Tenant consumption metrics"
      breadcrumb={<SettingsBreadcrumb page="Usage" />}
    >
      {isLoading && <LoadingState />}
      {data && (
        <KpiRow>
          <div data-testid="usage-logs-ingested">
            <MetricCard label="Logs ingested" value={`${data.logsIngestedGb} GB`} />
          </div>
          <div data-testid="usage-traces-ingested">
            <MetricCard label="Traces" value={data.tracesIngested.toLocaleString()} />
          </div>
          <div data-testid="usage-ai-tokens">
            <MetricCard label="AI tokens" value={data.aiTokensUsed.toLocaleString()} />
          </div>
          <div data-testid="usage-active-users">
            <MetricCard label="Active users" value={String(data.activeUsers)} />
          </div>
        </KpiRow>
      )}
    </StitchPageShell>
  );
}
