import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { formatDistanceToNow } from 'date-fns';
import { ChevronDown, ChevronRight, Copy, MessageSquare, Sparkles } from 'lucide-react';
import toast from 'react-hot-toast';
import { fetchTransaction, searchLogs, searchTrace } from '../../api/search';
import { fetchIncidents } from '../../api/incidents';
import type { LogHit } from '../../api/types';
import { hasStackTrace } from '../../utils/highlightText';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { CodeBlock } from '../ui/CodeBlock';
import { LoadingState } from '../ui/PageStates';
import { formatLogTimestamp } from './LogFiltersSidebar';

function severityVariant(sev: string) {
  if (sev === 'FATAL' || sev === 'CRITICAL') return 'critical';
  if (sev === 'ERROR') return 'error';
  if (sev === 'WARN') return 'warning';
  return 'info';
}

function StackTraceSection({ message }: { message: string }) {
  const [open, setOpen] = useState(true);
  const lines = message.split('\n').filter((l) => l.trim().startsWith('at ') || /Exception|Error:/.test(l));
  if (lines.length === 0) return null;

  const hotspot = lines.find((l) => !l.includes('java.lang') && !l.includes('runtime'));

  return (
    <section className="log-detail-section">
      <button type="button" className="log-detail-section__toggle" onClick={() => setOpen((v) => !v)}>
        {open ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
        Stack Trace Analysis
        <Badge variant="warning">Recurring?</Badge>
      </button>
      {open && (
        <div className="stack-trace">
          {lines.map((line, i) => (
            <div key={i} className={`stack-trace__line ${line === hotspot ? 'stack-trace__line--hotspot' : ''}`}>
              {line.trim()}
            </div>
          ))}
        </div>
      )}
    </section>
  );
}

interface LogDetailPanelProps {
  log: LogHit;
  onClose: () => void;
}

export function LogDetailPanel({ log, onClose }: LogDetailPanelProps) {
  const [rawOpen, setRawOpen] = useState(false);

  const traceQuery = useQuery({
    queryKey: ['related-trace', log.traceId],
    queryFn: () => searchTrace(log.traceId!),
    enabled: !!log.traceId,
  });

  const txnQuery = useQuery({
    queryKey: ['related-txn', log.txnId],
    queryFn: () => fetchTransaction(log.txnId!),
    enabled: !!log.txnId,
  });

  const similarQuery = useQuery({
    queryKey: ['similar-logs', log.service, log.message.slice(0, 40)],
    queryFn: () =>
      searchLogs({
        query: log.message.split('\n')[0].slice(0, 80),
        service: log.service,
        severity: log.severity,
        size: 5,
      }),
    enabled: log.message.length > 10,
  });

  const incidentsQuery = useQuery({
    queryKey: ['open-incidents'],
    queryFn: () => fetchIncidents({ status: 'OPEN' }),
  });

  const linkedIncident = incidentsQuery.data?.find((inc) =>
    inc.title.toLowerCase().includes(log.service.toLowerCase()),
  );

  const copyTrace = () => {
    if (log.traceId) {
      void navigator.clipboard.writeText(log.traceId);
      toast.success('Trace ID copied');
    }
  };

  return (
    <aside className="log-detail-panel">
      <div className="log-detail-panel__header">
        <h3>Log Detail</h3>
        <Button variant="ghost" size="sm" onClick={onClose}>
          Close
        </Button>
      </div>

      <div className="log-detail-meta">
        <Badge variant={severityVariant(log.severity) as 'error'}>{log.severity}</Badge>
        <Badge variant="info">{log.service}</Badge>
        <span className="muted">{formatLogTimestamp(log.timestamp)}</span>
      </div>

      <p className="log-detail-message">{log.message}</p>

      {(log.plainEnglish || log.classification) && (
        <section className="log-detail-section">
          <h4>
            <Sparkles size={14} /> AI Explanation
          </h4>
          <p>{log.plainEnglish ?? log.classification}</p>
          <ul className="insight-list">
            <li>Check recent deployments on {log.service}</li>
            <li>Review correlated traces and downstream dependencies</li>
            <li>Validate circuit breaker and retry configuration</li>
          </ul>
        </section>
      )}

      {hasStackTrace(log.message) && <StackTraceSection message={log.message} />}

      <section className="log-detail-section">
        <h4>Related logs</h4>
        {log.traceId && (
          <div className="related-group">
            <strong>Same trace ({log.traceId.slice(0, 12)}…)</strong>
            {traceQuery.isLoading && <LoadingState label="Loading trace…" />}
            {(traceQuery.data?.hits ?? []).slice(0, 5).map((hit) => (
              <div key={hit.id} className="related-log-row">
                <Badge variant="info">{hit.service}</Badge>
                <span className="log-row__message">{hit.message.slice(0, 80)}</span>
              </div>
            ))}
          </div>
        )}
        {log.txnId && (
          <div className="related-group">
            <strong>Transaction {log.txnId.slice(0, 12)}…</strong>
            {txnQuery.isLoading && <LoadingState label="Loading transaction…" />}
            {txnQuery.data?.hops.map((hop) => (
              <div key={hop.spanId} className="related-log-row">
                <Badge variant={hop.status === 'ERROR' ? 'error' : 'info'}>{hop.serviceName}</Badge>
                <span>{hop.latencyMs}ms · {hop.status}</span>
              </div>
            ))}
          </div>
        )}
        <div className="related-group">
          <strong>Similar errors</strong>
          {(similarQuery.data?.hits ?? [])
            .filter((h) => h.id !== log.id)
            .slice(0, 3)
            .map((hit) => (
              <div key={hit.id} className="related-log-row">
                <span className="muted">{formatDistanceToNow(new Date(hit.timestamp), { addSuffix: true })}</span>
                <span className="log-row__message">{hit.message.slice(0, 60)}</span>
              </div>
            ))}
        </div>
      </section>

      {linkedIncident && (
        <section className="log-detail-section">
          <h4>Linked incident</h4>
          <Link to="/incidents/$id" params={{ id: linkedIncident.id }} className="linked-incident">
            {linkedIncident.title}
          </Link>
        </section>
      )}

      <div className="log-detail-actions">
        {log.traceId && (
          <Button variant="secondary" size="sm" onClick={copyTrace}>
            <Copy size={14} /> Copy Trace
          </Button>
        )}
        {log.traceId && (
          <Link to="/traces/$traceId" params={{ traceId: log.traceId }}>
            <Button variant="secondary" size="sm">Open waterfall trace</Button>
          </Link>
        )}
        <Link to="/ai-chat" search={{ context: log.message.slice(0, 200) }}>
          <Button variant="primary" size="sm">
            <MessageSquare size={14} /> Open in AI Chat
          </Button>
        </Link>
      </div>

      <section className="log-detail-section">
        <button type="button" className="log-detail-section__toggle" onClick={() => setRawOpen((v) => !v)}>
          {rawOpen ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
          Raw JSON
        </button>
        {rawOpen && <CodeBlock code={JSON.stringify(log, null, 2)} language="json" />}
      </section>
    </aside>
  );
}
