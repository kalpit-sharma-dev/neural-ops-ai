import { useMemo, useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate } from '@tanstack/react-router';
import { formatDistanceToNow } from 'date-fns';
import { fetchIncidents } from '../api/incidents';
import type { Incident } from '../api/types';
import { getApiErrorMessage } from '../api/client';
import { Badge } from '../components/ui/Badge';
import { DomainEmptyState } from '../components/ui/DomainEmptyState';
import { Input } from '../components/ui/Input';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';
import { useListKeyboardNav } from '../hooks/useListKeyboardNav';

function severityBadge(sev: string) {
  const map: Record<string, 'p1' | 'p2' | 'p3' | 'p4'> = {
    P1: 'p1',
    P2: 'p2',
    P3: 'p3',
    P4: 'p4',
  };
  return map[sev] ?? 'p4';
}

export default function Incidents() {
  const navigate = useNavigate();
  const searchRef = useRef<HTMLInputElement>(null);
  const [query, setQuery] = useState('');

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['incidents'],
    queryFn: () => fetchIncidents({ size: 100 }),
    refetchInterval: 30_000,
  });

  const filtered = useMemo(() => {
    const list = data ?? [];
    const q = query.trim().toLowerCase();
    if (!q) return list;
    return list.filter(
      (inc) =>
        inc.title.toLowerCase().includes(q) ||
        inc.status.toLowerCase().includes(q) ||
        inc.affectedServices?.some((s) => s.toLowerCase().includes(q)),
    );
  }, [data, query]);

  const { activeIndex, setActiveIndex } = useListKeyboardNav(filtered, {
    searchInputRef: searchRef,
    onSelect: (inc: Incident) => {
      void navigate({ to: '/incidents/$id', params: { id: inc.id } });
    },
  });

  return (
    <StitchPageShell title="Incidents" subtitle="Active and recent operational incidents">
      <div style={{ marginBottom: 16, maxWidth: 400 }}>
        <Input
          ref={searchRef}
          placeholder="Filter incidents…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          hint="Press / to focus · j/k navigate · Enter to open"
        />
      </div>

      {isLoading && <LoadingState />}
      {error && <ErrorState message={getApiErrorMessage(error)} onRetry={() => refetch()} />}

      {data && filtered.length === 0 && !isLoading && (
        <DomainEmptyState domain="incidents" />
      )}

      {data && filtered.length > 0 && (
        <div className="ui-card">
          <table className="incident-table">
            <thead>
              <tr>
                <th>Severity</th>
                <th>Title</th>
                <th>Status</th>
                <th>Services</th>
                <th>Started</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((inc, idx) => (
                <tr
                  key={inc.id}
                  className={idx === activeIndex ? 'incident-table__row--active' : ''}
                  onMouseEnter={() => setActiveIndex(idx)}
                >
                  <td>
                    <Badge variant={severityBadge(inc.severity)}>{inc.severity}</Badge>
                  </td>
                  <td>
                    <Link to="/incidents/$id" params={{ id: inc.id }}>{inc.title}</Link>
                  </td>
                  <td>{inc.status}</td>
                  <td>{inc.affectedServices?.join(', ') ?? '—'}</td>
                  <td className="muted">
                    {formatDistanceToNow(new Date(inc.startTime), { addSuffix: true })}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </StitchPageShell>
  );
}
