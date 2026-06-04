import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { fetchK8sClusters, fetchK8sDeployments, fetchK8sNamespaces, fetchK8sPods } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { logsSearch } from '../utils/logsSearch';
import { Badge } from '../components/ui/Badge';
import { Card } from '../components/ui/Card';
import { DomainEmptyState } from '../components/ui/DomainEmptyState';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';
import { DataExportMenu } from '../components/ui/DataExportMenu';

type SortKey = 'name' | 'namespace' | 'readyReplicas' | 'replicas' | 'status' | 'node';
type SortDir = 'asc' | 'desc';

function sortRows<T extends Record<string, unknown>>(rows: T[], key: SortKey, dir: SortDir): T[] {
  return [...rows].sort((a, b) => {
    const av = a[key];
    const bv = b[key];
    if (typeof av === 'number' && typeof bv === 'number') return dir === 'asc' ? av - bv : bv - av;
    return dir === 'asc'
      ? String(av).localeCompare(String(bv))
      : String(bv).localeCompare(String(av));
  });
}

function SortHeader({
  label,
  sortKey,
  activeKey,
  dir,
  onSort,
}: {
  label: string;
  sortKey: SortKey;
  activeKey: SortKey;
  dir: SortDir;
  onSort: (k: SortKey) => void;
}) {
  const active = activeKey === sortKey;
  return (
    <th>
      <button type="button" className="table-sort-btn" onClick={() => onSort(sortKey)}>
        {label} {active ? (dir === 'asc' ? '↑' : '↓') : ''}
      </button>
    </th>
  );
}

export default function Kubernetes() {
  const [namespace, setNamespace] = useState('');
  const [depSort, setDepSort] = useState<{ key: SortKey; dir: SortDir }>({ key: 'name', dir: 'asc' });
  const [podSort, setPodSort] = useState<{ key: SortKey; dir: SortDir }>({ key: 'name', dir: 'asc' });

  const clustersQuery = useQuery({ queryKey: ['k8s-clusters'], queryFn: fetchK8sClusters });
  const nsQuery = useQuery({ queryKey: ['k8s-namespaces'], queryFn: fetchK8sNamespaces });
  const podsQuery = useQuery({ queryKey: ['k8s-pods', namespace], queryFn: () => fetchK8sPods(namespace || undefined) });
  const depQuery = useQuery({
    queryKey: ['k8s-deployments', namespace],
    queryFn: () => fetchK8sDeployments(namespace || undefined),
  });

  const toggleDepSort = (key: SortKey) => {
    setDepSort((s) => (s.key === key ? { key, dir: s.dir === 'asc' ? 'desc' : 'asc' } : { key, dir: 'asc' }));
  };
  const togglePodSort = (key: SortKey) => {
    setPodSort((s) => (s.key === key ? { key, dir: s.dir === 'asc' ? 'desc' : 'asc' } : { key, dir: 'asc' }));
  };

  const deployments = useMemo(
    () => sortRows(depQuery.data ?? [], depSort.key, depSort.dir),
    [depQuery.data, depSort],
  );
  const pods = useMemo(
    () => sortRows(podsQuery.data ?? [], podSort.key, podSort.dir),
    [podsQuery.data, podSort],
  );

  return (
    <StitchPageShell
      title="Kubernetes"
      subtitle="Live cluster inventory via Kubernetes API (in-cluster) or Prometheus fallback"
      actions={
        <>
          <DataExportMenu
            getData={() => [
              ...(clustersQuery.data ?? []).map((c) => ({ recordType: 'cluster', ...c })),
              ...deployments.map((d) => ({ recordType: 'deployment', ...d })),
              ...pods.map((p) => ({ recordType: 'pod', ...p })),
            ]}
            filenamePrefix="kubernetes"
            disabled={
              !(clustersQuery.data?.length || deployments.length || pods.length)
            }
          />
          <Link to="/infrastructure">← Infrastructure</Link>
        </>
      }
    >
      {clustersQuery.isLoading && <LoadingState />}
      {clustersQuery.error && <ErrorState message={getApiErrorMessage(clustersQuery.error)} onRetry={() => clustersQuery.refetch()} />}
      {!clustersQuery.error && clustersQuery.isFetched && (clustersQuery.data?.length ?? 0) === 0 && (
        <DomainEmptyState domain="kubernetes" />
      )}
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
          <thead>
            <tr>
              <SortHeader label="Name" sortKey="name" activeKey={depSort.key} dir={depSort.dir} onSort={toggleDepSort} />
              <SortHeader label="Namespace" sortKey="namespace" activeKey={depSort.key} dir={depSort.dir} onSort={toggleDepSort} />
              <SortHeader label="Ready" sortKey="readyReplicas" activeKey={depSort.key} dir={depSort.dir} onSort={toggleDepSort} />
              <SortHeader label="Replicas" sortKey="replicas" activeKey={depSort.key} dir={depSort.dir} onSort={toggleDepSort} />
            </tr>
          </thead>
          <tbody>
            {deployments.map((d) => (
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
          <thead>
            <tr>
              <SortHeader label="Name" sortKey="name" activeKey={podSort.key} dir={podSort.dir} onSort={togglePodSort} />
              <SortHeader label="Namespace" sortKey="namespace" activeKey={podSort.key} dir={podSort.dir} onSort={togglePodSort} />
              <th>Node</th>
              <SortHeader label="Status" sortKey="status" activeKey={podSort.key} dir={podSort.dir} onSort={togglePodSort} />
              <th>Logs</th>
            </tr>
          </thead>
          <tbody>
            {pods.map((p) => (
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
    </StitchPageShell>
  );
}
