import { useMemo } from 'react';
import type { Span } from '../../api/observability';
import { Badge } from '../../components/ui/Badge';

interface FlameNode {
  span: Span;
  children: FlameNode[];
  depth: number;
}

interface TraceFlameGraphProps {
  spans: Span[];
  traceTotalMs: number;
  selectedSpanId?: string;
  onSelectSpan?: (span: Span) => void;
}

function buildFlameTree(spans: Span[]): FlameNode[] {
  const byId = new Map(spans.map((s) => [s.spanId, s]));
  const children = new Map<string, Span[]>();
  const roots: Span[] = [];

  for (const span of spans) {
    if (!span.parentId || !byId.has(span.parentId)) {
      roots.push(span);
    } else {
      const list = children.get(span.parentId) ?? [];
      list.push(span);
      children.set(span.parentId, list);
    }
  }

  function toNode(span: Span, depth: number): FlameNode {
    return {
      span,
      depth,
      children: (children.get(span.spanId) ?? []).map((c) => toNode(c, depth + 1)),
    };
  }

  return roots.map((r) => toNode(r, 0));
}

function FlameBar({
  node,
  traceTotalMs,
  selected,
  onSelect,
}: {
  node: FlameNode;
  traceTotalMs: number;
  selected: boolean;
  onSelect: (span: Span) => void;
}) {
  const widthPct = Math.max(1.5, (node.span.durationMs / Math.max(traceTotalMs, 1)) * 100);
  const hue = (node.depth * 37 + node.span.service.length * 11) % 360;
  const statusVariant = node.span.status === 'ERROR' ? 'critical' : 'healthy';

  return (
    <div className="trace-flame">
      <button
        type="button"
        className={`trace-flame__bar ${selected ? 'trace-flame__bar--selected' : ''}`}
        style={{
          width: `${widthPct}%`,
          background: `hsl(${hue} 55% 42%)`,
        }}
        onClick={() => onSelect(node.span)}
        title={`${node.span.service} · ${node.span.operation} · ${node.span.durationMs}ms`}
      >
        <Badge variant={statusVariant}>{node.span.service}</Badge>
        <span className="trace-flame__label">{node.span.operation}</span>
      </button>
      {node.children.length > 0 && (
        <div className="trace-flame__children">
          {node.children.map((child) => (
            <FlameBar
              key={child.span.spanId}
              node={child}
              traceTotalMs={traceTotalMs}
              selected={selected}
              onSelect={onSelect}
            />
          ))}
        </div>
      )}
    </div>
  );
}

export function TraceFlameGraph({ spans, traceTotalMs, selectedSpanId, onSelectSpan }: TraceFlameGraphProps) {
  const roots = useMemo(() => buildFlameTree(spans), [spans]);

  if (roots.length === 0) {
    return <p className="muted">No spans to render</p>;
  }

  return (
    <div className="trace-flame-graph">
      {roots.map((root) => (
        <FlameBar
          key={root.span.spanId}
          node={root}
          traceTotalMs={traceTotalMs}
          selected={root.span.spanId === selectedSpanId}
          onSelect={(s) => onSelectSpan?.(s)}
        />
      ))}
    </div>
  );
}
