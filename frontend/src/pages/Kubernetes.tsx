import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { fetchK8sClusters, fetchK8sDeployments, fetchK8sNamespaces, fetchK8sPods } from '../api/observability';
import { logsSearch } from '../utils/logsSearch';
import { Badge } from '../components/ui/Badge';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function Kubernetes() {
  const [namespace, setNamespace] = useState('');
  const clustersQuery = useQuery({ queryKey: ['k8s-clusters'], queryFn: fetchK8sClusters });
  const nsQuery = useQuery({ queryKey: ['k8s-namespaces'], queryFn: fetchK8sNamespaces });
  const podsQuery = useQuery({ queryKey: ['k8s-pods', namespace], queryFn: () => fetchK8sPods(namespace || undefined) });
  const depQuery = useQuery({
    queryKey: ['k8s-deployments', namespace],
    queryFn: () => fetchK8sDeployments(namespace || undefined),
  });

  return (
    <div>
      <PageHeader
        title="Kubernetes"
        subtitle="Live cluster inventory via Kubernetes API (in-cluster) or Prometheus fallback"
        actions={<Link to="/infrastructure">← Infrastructure</Link>}
      />
      {clustersQuery.isLoading && <LoadingState />}
      {(clustersQuery.data ?? []).map((c) => (
        <Card key={c.id} title={c.name}>
          <p>{c.nodes} nodes · {c.pods} pods · <Badge variant="healthy">{c.health}</Badge></p>
        </Card>
      ))}
      <Card title="Namespaces" style={{ marginTop: 16 }}>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginBottom: 12 }}>
          <button type="button" className={`tab-bar__item ${namespace === '' ? 'tab-bar__item--active' : ''}`} onClick={() => setNamespace('')}>All</button>
          {(nsQuery.data ?? []).map((n) => (
            <button key={n.id} type="button" className={`tab-bar__item ${namespace === n.name ? 'tab-bar__item--active' : ''}`} onClick={() => setNamespace(n.name)}>
              {n.name} ({n.podCount})
            </button>
          ))}
        </div>
      </Card>
      <Card title="Deployments" style={{ marginTop: 16 }}>
        <table className="data-table">
          <thead><tr><th>Name</th><th>Namespace</th><th>Ready</th><th>Replicas</th></tr></thead>
          <tbody>
            {(depQuery.data ?? []).map((d) => (
              <tr key={d.id}>
                <td>{d.name}</td>
                <td>{d.namespace}</td>
                <td>{d.readyReplicas}/{d.replicas}</td>
                <td>{d.replicas}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
      <Card title="Pods" style={{ marginTop: 16 }}>
        <table className="data-table">
          <thead><tr><th>Name</th><th>Namespace</th><th>Node</th><th>Status</th><th>Logs</th></tr></thead>
          <tbody>
            {(podsQuery.data ?? []).map((p) => (
              <tr key={p.id}>
                <td>{p.name}</td>
                <td>{p.namespace}</td>
                <td>{p.node}</td>
                <td>{p.status}</td>
                <td>
                  <Link to="/logs" search={logsSearch({ service: p.name })}>Logs</Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
