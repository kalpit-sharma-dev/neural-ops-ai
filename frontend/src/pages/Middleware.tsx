import { useQuery } from '@tanstack/react-query';
import { fetchKafkaLag } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Card } from '../components/ui/Card';
import { EmptyState } from '../components/ui/EmptyState';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';

export default function Middleware() {
  const { data, isLoading, error, refetch, isFetched } = useQuery({ queryKey: ['kafka-lag'], queryFn: fetchKafkaLag });

  return (
    <StitchPageShell title="Middleware" subtitle="Kafka consumer lag and queue depth">
      {isLoading && <LoadingState />}
      {error && <ErrorState message={getApiErrorMessage(error)} onRetry={() => refetch()} />}
      {!error && isFetched && (data?.length ?? 0) === 0 && (
        <EmptyState title="No consumer lag data" description="Connect a Kafka collector to surface consumer lag and queue depth." />
      )}
      {!error && (data?.length ?? 0) > 0 && (
        <Card title="Kafka lag">
          <table className="data-table">
            <thead><tr><th>Topic</th><th>Consumer group</th><th>Partition</th><th>Lag</th></tr></thead>
            <tbody>
              {(data ?? []).map((r, i) => (
                <tr key={i}>
                  <td>{r.topic}</td>
                  <td>{r.consumerGroup}</td>
                  <td>{r.partition}</td>
                  <td>{r.lag}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}
    </StitchPageShell>
  );
}
