import { useQuery } from '@tanstack/react-query';
import { fetchKafkaLag } from '../api/observability';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function Middleware() {
  const { data, isLoading } = useQuery({ queryKey: ['kafka-lag'], queryFn: fetchKafkaLag });

  return (
    <div>
      <PageHeader title="Middleware" subtitle="Kafka consumer lag and queue depth" />
      {isLoading && <LoadingState />}
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
    </div>
  );
}
