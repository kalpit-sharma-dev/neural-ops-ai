import { useMemo, useState, type ReactNode } from 'react';
import { Link } from '@tanstack/react-router';
import { logsSearch } from '../../utils/logsSearch';
import type { Span } from '../../api/observability';
import { Badge } from '../../components/ui/Badge';

interface TraceWaterfallProps {
  spans: Span[];
  traceTotalMs: number;
  traceId?: string;
  selectedSpanId?: string;
  onSelectSpan?: (span: Span) => void;
}

function buildTree(spans: Span[]) {
  const byId = new Map(spans.map((s) => [s.spanId, s]));
  const roots: Span[] = [];
  const children = new Map<string, Span[]>();
  for (const span of spans) {
    if (!span.parentId || !byId.has(span.parentId)) {
      roots.push(span);
    } else {
      const list = children.get(span.parentId) ?? [];
      list.push(span);
      children.set(span.parentId, list);
    }
  }
  return { roots, children };
}

function SpanRow({
  span,
  depth,
  traceTotalMs,
  selected,
  onSelect,
  children,
}: {
  span: Span;
  depth: number;
  traceTotalMs: number;
  selected: boolean;
  onSelect: () => void;
  children: ReactNode;
}) {
  const widthPct = Math.max(2, (span.durationMs / Math.max(traceTotalMs, 1)) * 100);
  const statusVariant = span.status === 'ERROR' ? 'critical' : span.status === 'OK' ? 'healthy' : 'warning';

  return (
    <>
      <div
        className={`trace-waterfall__row ${selected ? 'trace-waterfall__row--selected' : ''}`}
        onClick={onSelect}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => e.key === 'Enter' && onSelect()}
      >
        <div className="trace-waterfall__label" style={{ paddingLeft: depth * 16 + 8 }}>
          <Badge variant={statusVariant}>{span.service}</Badge>
          <span className="trace-waterfall__op">{span.operation}</span>
          <span className="muted">{span.durationMs}ms</span>
        </div>
        <div className="trace-waterfall__bar-track">
          <div
            className={`trace-waterfall__bar trace-waterfall__bar--${span.status.toLowerCase()}`}
            style={{ width: `${widthPct}%` }}
          />
        </div>
      </div>
      {children}
    </>
  );
}

function renderSpanTree(
  span: Span,
  depth: number,
  traceTotalMs: number,
  childrenMap: Map<string, Span[]>,
  selectedSpanId: string | undefined,
  onSelectSpan?: (span: Span) => void,
): ReactNode {
  const kids = childrenMap.get(span.spanId) ?? [];
  return (
    <SpanRow
      key={span.spanId}
      span={span}
      depth={depth}
      traceTotalMs={traceTotalMs}
      selected={selectedSpanId === span.spanId}
      onSelect={() => onSelectSpan?.(span)}
      children={kids.map((child) =>
        renderSpanTree(child, depth + 1, traceTotalMs, childrenMap, selectedSpanId, onSelectSpan),
      )}
    />
  );
}

export function TraceWaterfall({ spans, traceTotalMs, traceId, selectedSpanId, onSelectSpan }: TraceWaterfallProps) {
  const [internalSelected, setInternalSelected] = useState<string | undefined>();
  const selected = selectedSpanId ?? internalSelected;
  const { roots, children } = useMemo(() => buildTree(spans), [spans]);

  const handleSelect = (span: Span) => {
    setInternalSelected(span.spanId);
    onSelectSpan?.(span);
  };

  const selectedSpan = spans.find((s) => s.spanId === selected);

  return (
    <div className="trace-waterfall">
      <div className="trace-waterfall__header">
        <span>Span tree</span>
        <span className="muted">{spans.length} spans · {traceTotalMs}ms total</span>
      </div>
      <div className="trace-waterfall__body">
        {roots.map((root) => renderSpanTree(root, 0, traceTotalMs, children, selected, handleSelect))}
      </div>
      {selectedSpan && (
        <div className="trace-waterfall__detail ui-card">
          <p><strong>{selectedSpan.service}</strong> — {selectedSpan.operation}</p>
          <p className="muted">Span ID: {selectedSpan.spanId}</p>
          {selectedSpan.tags && (
            <pre className="code-block">{JSON.stringify(selectedSpan.tags, null, 2)}</pre>
          )}
          <Link
            to="/logs"
            search={logsSearch({
              service: selectedSpan.service,
              traceId: traceId ?? selectedSpan.traceId,
              from: new Date(new Date(selectedSpan.startTime).getTime() - 5000).toISOString(),
              to: new Date(new Date(selectedSpan.startTime).getTime() + selectedSpan.durationMs + 5000).toISOString(),
            })}
            className="trace-waterfall__logs-link"
          >
            View logs for span window ({selectedSpan.durationMs}ms) →
          </Link>
        </div>
      )}
    </div>
  );
}
