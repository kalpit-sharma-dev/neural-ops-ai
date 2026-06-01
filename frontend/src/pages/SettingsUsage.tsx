import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { fetchUsage } from '../api/observability';
import { MetricCard } from '../components/ui/MetricCard';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function SettingsUsage() {
  const { data, isLoading } = useQuery({ queryKey: ['admin-usage'], queryFn: fetchUsage });

  return (
    <div>
      <PageHeader title="Usage" subtitle="Tenant consumption metrics" actions={<Link to="/settings">← Settings</Link>} />
      {isLoading && <LoadingState />}
      {data && (
        <div className="dashboard-row-4">
          <MetricCard label="Logs ingested" value={`${data.logsIngestedGb} GB`} />
          <MetricCard label="Traces" value={data.tracesIngested.toLocaleString()} />
          <MetricCard label="AI tokens" value={data.aiTokensUsed.toLocaleString()} />
          <MetricCard label="Active users" value={String(data.activeUsers)} />
        </div>
      )}
    </div>
  );
}
