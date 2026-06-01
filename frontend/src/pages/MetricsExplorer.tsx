import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { fetchMetricCatalog, queryMetric } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Card } from '../components/ui/Card';
import { ErrorState, LoadingState, PageHeader } from '../components/ui/PageStates';
import { Select } from '../components/ui/Select';

export default function MetricsExplorer() {
  const [metric, setMetric] = useState('latency_p95');
  const [service, setService] = useState('payment-service');

  const catalogQuery = useQuery({ queryKey: ['metric-catalog'], queryFn: fetchMetricCatalog });
  const seriesQuery = useQuery({
    queryKey: ['metric-query', metric, service],
    queryFn: () => queryMetric(metric, service),
    enabled: !!metric,
  });

  const chartData = (seriesQuery.data?.points ?? []).map((p) => ({
    time: new Date(p.timestamp).toLocaleTimeString(),
    value: p.value,
  }));

  return (
    <div>
      <PageHeader title="Metrics Explorer" subtitle="Browse and query platform metrics" />
      <div style={{ display: 'flex', gap: 12, marginBottom: 24 }}>
        <Select value={metric} onChange={(e) => setMetric(e.target.value)}>
          {(catalogQuery.data ?? [{ name: 'latency_p95' }]).map((m) => (
            <option key={m.name} value={m.name}>{m.name}</option>
          ))}
        </Select>
        <Select value={service} onChange={(e) => setService(e.target.value)}>
          <option value="payment-service">payment-service</option>
          <option value="api-gateway">api-gateway</option>
          <option value="auth-service">auth-service</option>
        </Select>
      </div>
      {seriesQuery.isLoading && <LoadingState />}
      {seriesQuery.error && <ErrorState message={getApiErrorMessage(seriesQuery.error)} onRetry={() => seriesQuery.refetch()} />}
      {seriesQuery.data && (
        <Card title={`${metric} — ${service}`}>
          <ResponsiveContainer width="100%" height={280}>
            <LineChart data={chartData}>
              <XAxis dataKey="time" stroke="var(--text-muted)" fontSize={11} />
              <YAxis stroke="var(--text-muted)" fontSize={11} />
              <Tooltip contentStyle={{ background: 'var(--bg-elevated)', border: '1px solid var(--border-subtle)' }} />
              <Line type="monotone" dataKey="value" stroke="var(--accent-primary)" dot={false} strokeWidth={2} />
            </LineChart>
          </ResponsiveContainer>
        </Card>
      )}
    </div>
  );
}
