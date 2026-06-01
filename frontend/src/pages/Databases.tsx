import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { fetchDBStatements, fetchDatabases } from '../api/observability';
import { Card } from '../components/ui/Card';
import { Badge } from '../components/ui/Badge';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function Databases() {
  const [selected, setSelected] = useState('');
  const dbQuery = useQuery({ queryKey: ['databases'], queryFn: fetchDatabases });
  const stmtQuery = useQuery({
    queryKey: ['db-statements', selected],
    queryFn: () => fetchDBStatements(selected),
    enabled: !!selected,
  });

  return (
    <div>
      <PageHeader title="Databases" subtitle="Query performance and connection pools" />
      {dbQuery.isLoading && <LoadingState />}
      <div className="dashboard-row-2">
        {(dbQuery.data ?? []).map((db) => (
          <Card key={db.id} title={db.name}>
            <Badge variant={db.status === 'healthy' ? 'healthy' : 'warning'}>{db.engine}</Badge>
            <p>QPS {db.qps} · Slow {db.slowQueries} · Conn {db.connections}</p>
            <button type="button" onClick={() => setSelected(db.id)}>View statements</button>
          </Card>
        ))}
      </div>
      {selected && stmtQuery.data && (
        <Card title="Top statements">
          <table className="data-table">
            <thead><tr><th>Query</th><th>Calls</th><th>Avg ms</th></tr></thead>
            <tbody>
              {stmtQuery.data.map((s, i) => (
                <tr key={i}><td><code>{s.query.slice(0, 80)}…</code></td><td>{s.calls}</td><td>{s.avgMs}</td></tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}
    </div>
  );
}
