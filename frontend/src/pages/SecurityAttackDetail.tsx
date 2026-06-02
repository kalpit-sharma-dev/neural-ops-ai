import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from '@tanstack/react-router';
import { Activity, FileText, Globe, Server, ShieldAlert, ShieldCheck } from 'lucide-react';
import { fetchAttack } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { logsSearch } from '../utils/logsSearch';
import { Badge } from '../components/ui/Badge';
import { Card } from '../components/ui/Card';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';

export default function SecurityAttackDetail() {
  const { id = '' } = useParams({ strict: false });

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['security-attack', id],
    queryFn: () => fetchAttack(id),
    enabled: id.length > 0,
  });

  return (
    <StitchPageShell
      title={data?.type ?? 'Attack detail'}
      subtitle="Runtime attack event"
      breadcrumb={
        <nav className="settings-breadcrumb" aria-label="Breadcrumb">
          <Link to="/security" className="settings-breadcrumb__link">Application Security</Link>
          <span className="settings-breadcrumb__sep">/</span>
          <span className="settings-breadcrumb__current" aria-current="page">{data?.type ?? 'Attack'}</span>
        </nav>
      }
    >
      {isLoading && <LoadingState />}
      {error && <ErrorState message={getApiErrorMessage(error)} onRetry={() => refetch()} />}

      {data && (
        <>
          <Card
            title="Event"
            action={
              <Badge variant={data.blocked ? 'healthy' : 'critical'}>
                {data.blocked ? 'Blocked' : 'Open'}
              </Badge>
            }
          >
            <div className="attack-fields">
              <div className="attack-field">
                <span className="attack-field__label">Attack type</span>
                <span className="attack-field__value">{data.type}</span>
              </div>
              <div className="attack-field">
                <span className="attack-field__label">Source IP</span>
                <span className="attack-field__value">{data.sourceIp}</span>
              </div>
              <div className="attack-field">
                <span className="attack-field__label">Target service</span>
                <span className="attack-field__value">{data.service}</span>
              </div>
              <div className="attack-field">
                <span className="attack-field__label">Detected</span>
                <span className="attack-field__value">{new Date(data.detectedAt).toLocaleString()}</span>
              </div>
              <div className="attack-field">
                <span className="attack-field__label">Mitigation</span>
                <span className="attack-field__value" style={{ fontFamily: 'var(--font-body)' }}>
                  {data.blocked ? (
                    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6, color: 'var(--success)' }}>
                      <ShieldCheck size={15} aria-hidden /> Request blocked at the edge
                    </span>
                  ) : (
                    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6, color: 'var(--error)' }}>
                      <ShieldAlert size={15} aria-hidden /> Reached service — review impact
                    </span>
                  )}
                </span>
              </div>
            </div>
          </Card>

          <Card title="Investigate">
            <p className="muted" style={{ marginTop: 0 }}>
              Pivot into correlated telemetry for <strong>{data.service}</strong> around the detection window.
            </p>
            <div className="attack-actions">
              <Link
                to="/traces"
                search={{ service: data.service }}
                className="ui-button ui-button--secondary ui-button--sm"
              >
                <Activity size={14} aria-hidden /> View traces
              </Link>
              <Link
                to="/logs"
                search={logsSearch({ service: data.service })}
                className="ui-button ui-button--secondary ui-button--sm"
              >
                <FileText size={14} aria-hidden /> View logs
              </Link>
              <Link
                to="/entities/$type/$id"
                params={{ type: 'service', id: data.service }}
                className="ui-button ui-button--secondary ui-button--sm"
              >
                <Server size={14} aria-hidden /> Entity page
              </Link>
              <a
                href={`https://www.abuseipdb.com/check/${encodeURIComponent(data.sourceIp)}`}
                target="_blank"
                rel="noreferrer"
                className="ui-button ui-button--ghost ui-button--sm"
              >
                <Globe size={14} aria-hidden /> Lookup source IP
              </a>
            </div>
          </Card>
        </>
      )}
    </StitchPageShell>
  );
}
