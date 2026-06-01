import { useQuery } from '@tanstack/react-query';
import { fetchServiceFlow } from '../api/observability';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function ServiceFlow() {
  const { data, isLoading } = useQuery({ queryKey: ['service-flow'], queryFn: fetchServiceFlow });

  return (
    <div>
      <PageHeader title="Service Flow" subtitle="Request path latency and error rates between services" />
      {isLoading && <LoadingState />}
      <Card title="Call graph">
        <table className="data-table">
          <thead>
            <tr><th>Source</th><th>Target</th><th>Calls</th><th>Error %</th><th>P50</th><th>P95</th></tr>
          </thead>
          <tbody>
            {(data ?? []).map((e) => (
              <tr key={`${e.source}-${e.target}`}>
                <td>{e.source}</td>
                <td>{e.target}</td>
                <td>{e.callCount.toLocaleString()}</td>
                <td>{e.errorRate.toFixed(1)}%</td>
                <td>{e.p50Ms}ms</td>
                <td>{e.p95Ms}ms</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
