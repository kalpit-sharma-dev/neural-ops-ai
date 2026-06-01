import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { fetchDBStatements, fetchDatabases } from '../api/observability';
import { Card } from '../components/ui/Card';
import { Badge } from '../components/ui/Badge';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

type StmtSort = 'avgMs' | 'calls' | 'totalMs';

export default function Databases() {
  const [selected, setSelected] = useState('');
  const [sortKey, setSortKey] = useState<StmtSort>('avgMs');
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');

  const dbQuery = useQuery({ queryKey: ['databases'], queryFn: fetchDatabases });
  const stmtQuery = useQuery({
    queryKey: ['db-statements', selected],
    queryFn: () => fetchDBStatements(selected),
    enabled: !!selected,
  });

  const sortedStatements = useMemo(() => {
    const rows = stmtQuery.data ?? [];
    return [...rows].sort((a, b) => {
      const av = a[sortKey];
      const bv = b[sortKey];
      return sortDir === 'asc' ? av - bv : bv - av;
    });
  }, [stmtQuery.data, sortKey, sortDir]);

  const toggleSort = (key: StmtSort) => {
    setSortKey((k) => {
      if (k === key) {
        setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
        return key;
      }
      setSortDir('desc');
      return key;
    });
  };

  return (
    <div>
      <PageHeader title="Databases" subtitle="Query performance and connection pools" />
      {dbQuery.isLoading && <LoadingState />}
      <div className="dashboard-row-2">
        {(dbQuery.data ?? []).map((db) => (
          <Card key={db.id} title={db.name}>
            <Badge variant={db.status === 'healthy' ? 'healthy' : 'warning'}>{db.engine}</Badge>
            <p>QPS {db.qps} · Slow {db.slowQueries} · Conn {db.connections}</p>
            <button type="button" className="pill" onClick={() => setSelected(db.id)}>View slow queries</button>
          </Card>
        ))}
      </div>
      {selected && stmtQuery.data && (
        <Card title="Slow queries (sortable)">
          <table className="data-table">
            <thead>
              <tr>
                <th>Query</th>
                <th>
                  <button type="button" className="table-sort-btn" onClick={() => toggleSort('calls')}>
                    Calls {sortKey === 'calls' ? (sortDir === 'asc' ? '↑' : '↓') : ''}
                  </button>
                </th>
                <th>
                  <button type="button" className="table-sort-btn" onClick={() => toggleSort('avgMs')}>
                    Avg ms (p95 proxy) {sortKey === 'avgMs' ? (sortDir === 'asc' ? '↑' : '↓') : ''}
                  </button>
                </th>
                <th>Trace</th>
              </tr>
            </thead>
            <tbody>
              {sortedStatements.map((s, i) => {
                const traceId = `trace-db-${selected}-${i}`;
                return (
                  <tr key={i}>
                    <td><code>{s.query.length > 80 ? `${s.query.slice(0, 80)}…` : s.query}</code></td>
                    <td>{s.calls}</td>
                    <td>{s.avgMs}</td>
                    <td>
                      <Link to="/traces/$traceId" params={{ traceId }} search={{ service: selected }}>
                        View trace
                      </Link>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </Card>
      )}
    </div>
  );
}
