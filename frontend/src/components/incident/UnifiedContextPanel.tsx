import { Link } from '@tanstack/react-router';
import { useQuery } from '@tanstack/react-query';
import {
  fetchIncidentUnifiedContext,
  type SecurityFinding,
  type SuggestedQuery,
} from '../../api/observability';
import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { LoadingState } from '../ui/PageStates';

interface UnifiedContextPanelProps {
  incidentId: string;
  primaryService?: string;
  affectedServices?: string[];
}

function severityVariant(sev: string) {
  const s = sev.toUpperCase();
  if (s === 'CRITICAL' || s === 'HIGH') return 'critical' as const;
  if (s === 'MEDIUM') return 'warning' as const;
  return 'info' as const;
}

export function UnifiedContextPanel({ incidentId, primaryService, affectedServices }: UnifiedContextPanelProps) {
  const service = primaryService ?? affectedServices?.[0] ?? '';
  const ctxQuery = useQuery({
    queryKey: ['incident-unified-context', incidentId, service],
    queryFn: () =>
      fetchIncidentUnifiedContext(incidentId, {
        service,
        services: affectedServices?.join(','),
      }),
    enabled: !!incidentId,
  });

  if (ctxQuery.isLoading) return <LoadingState label="Loading unified context…" />;
  const ctx = ctxQuery.data;
  if (!ctx) return null;

  const links = [
    { label: 'Logs', href: ctx.deepLinks.logs },
    { label: 'Traces', href: ctx.deepLinks.traces },
    { label: 'Metrics', href: ctx.deepLinks.metrics },
    { label: 'Security', href: ctx.deepLinks.security },
    { label: 'Service map', href: ctx.deepLinks.serviceMap },
    { label: 'AI assistant', href: ctx.deepLinks.aiChat },
  ];

  return (
    <Card title="Unified context">
      <p className="muted" style={{ marginBottom: 12 }}>
        Cross-signal triage for <strong>{ctx.primaryService}</strong> — one workflow vs switching between APM, logs, and
        security tools.
      </p>
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginBottom: 16 }}>
        {links.map((l) => (
          <a key={l.label} href={l.href}>
            <Button size="sm" variant="secondary" type="button">
              {l.label}
            </Button>
          </a>
        ))}
      </div>
      {ctx.suggestedQueries.length > 0 && (
        <>
          <p className="muted">Suggested investigations</p>
          <ul className="insight-list">
            {ctx.suggestedQueries.map((q: SuggestedQuery) => (
              <li key={q.href}>
                <a href={q.href}>{q.label}</a> <Badge variant="info">{q.kind}</Badge>
              </li>
            ))}
          </ul>
        </>
      )}
      {ctx.securityFindings.length > 0 && (
        <>
          <p className="muted" style={{ marginTop: 12 }}>
            Correlated security findings
          </p>
          <ul className="insight-list">
            {ctx.securityFindings.map((f: SecurityFinding) => (
              <li key={f.id}>
                <Link to="/security/findings/$id" params={{ id: f.id }}>
                  {f.title}
                </Link>{' '}
                <Badge variant={severityVariant(f.severity)}>{f.severity}</Badge>
              </li>
            ))}
          </ul>
        </>
      )}
    </Card>
  );
}
