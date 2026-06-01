import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate } from '@tanstack/react-router';
import { fetchServiceFlow } from '../api/observability';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function ServiceFlow() {
  const navigate = useNavigate();
  const { data, isLoading } = useQuery({ queryKey: ['service-flow'], queryFn: fetchServiceFlow });

  const nodes = useMemo(() => {
    const set = new Set<string>();
    for (const e of data ?? []) {
      set.add(e.source);
      set.add(e.target);
    }
    return [...set].sort();
  }, [data]);

  return (
    <div>
      <PageHeader title="Service Flow" subtitle="Request path latency and error rates between services" />
      {isLoading && <LoadingState />}

      {nodes.length > 0 && (
        <Card title="Services (click to open entity)">
          <div className="service-flow-nodes">
            {nodes.map((svc) => (
              <button
                key={svc}
                type="button"
                className="pill service-flow-node"
                onClick={() => void navigate({ to: '/entities/$type/$id', params: { type: 'service', id: svc } })}
              >
                {svc}
              </button>
            ))}
          </div>
        </Card>
      )}

      <Card title="Call graph" style={{ marginTop: 16 }}>
        <table className="data-table">
          <thead>
            <tr><th>Source</th><th>Target</th><th>Calls</th><th>Error %</th><th>P50</th><th>P95</th><th /></tr>
          </thead>
          <tbody>
            {(data ?? []).map((e) => (
              <tr key={`${e.source}-${e.target}`}>
                <td>
                  <Link to="/entities/$type/$id" params={{ type: 'service', id: e.source }}>{e.source}</Link>
                </td>
                <td>
                  <Link to="/entities/$type/$id" params={{ type: 'service', id: e.target }}>{e.target}</Link>
                </td>
                <td>{e.callCount.toLocaleString()}</td>
                <td>{e.errorRate.toFixed(1)}%</td>
                <td>{e.p50Ms}ms</td>
                <td>{e.p95Ms}ms</td>
                <td>
                  <Link to="/traces" search={{ service: e.source }}>Traces</Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
