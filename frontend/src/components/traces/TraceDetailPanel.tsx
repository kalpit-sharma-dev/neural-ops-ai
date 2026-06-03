import { Link } from '@tanstack/react-router';
import { ExternalLink } from 'lucide-react';
import type { TraceSummary } from '../../api/observability';
import { TraceDetailContent } from '../../features/traces/TraceDetailContent';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';

interface TraceDetailPanelProps {
  trace: TraceSummary;
  onClose: () => void;
}

export function TraceDetailPanel({ trace, onClose }: TraceDetailPanelProps) {
  return (
    <aside className="trace-detail-panel">
      <div className="trace-detail-panel__header">
        <div>
          <h3>Trace detail</h3>
          <p className="muted trace-detail-panel__id" title={trace.traceId}>
            {trace.traceId}
          </p>
        </div>
        <div className="trace-detail-panel__actions">
          <Link
            to="/traces/$traceId"
            params={{ traceId: trace.traceId }}
            className="trace-detail-panel__open-link"
          >
            <ExternalLink size={14} aria-hidden />
            Full page
          </Link>
          <Button variant="ghost" size="sm" onClick={onClose}>
            Close
          </Button>
        </div>
      </div>

      <div className="trace-detail-meta">
        <Badge variant={trace.status === 'ERROR' ? 'critical' : 'healthy'}>{trace.status}</Badge>
        <Badge variant="info">{trace.service}</Badge>
        <span className="muted">{trace.operation}</span>
        <span className="muted">
          {trace.durationMs}ms · {trace.spanCount} spans
        </span>
      </div>

      <TraceDetailContent traceId={trace.traceId} showProfile={false} />
    </aside>
  );
}
