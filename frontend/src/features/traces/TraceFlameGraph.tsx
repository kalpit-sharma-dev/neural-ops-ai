import { useMemo, useState } from 'react';
import { Minus, Plus, RotateCcw } from 'lucide-react';
import type { Span } from '../../api/observability';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';

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
  scale,
  selected,
  onSelect,
}: {
  node: FlameNode;
  traceTotalMs: number;
  scale: number;
  selected: boolean;
  onSelect: (span: Span) => void;
}) {
  const widthPct = Math.max(1.5, (node.span.durationMs / Math.max(traceTotalMs, 1)) * 100 * scale);
  const hue = (node.depth * 37 + node.span.service.length * 11) % 360;
  const statusVariant = node.span.status === 'ERROR' ? 'critical' : 'healthy';

  return (
    <div className="trace-flame">
      <button
        type="button"
        className={`trace-flame__bar ${selected ? 'trace-flame__bar--selected' : ''}`}
        style={{
          width: `${Math.min(widthPct, 100)}%`,
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
              scale={scale}
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
  const [scale, setScale] = useState(1);
  const [internalSelected, setInternalSelected] = useState<string | undefined>();
  const selected = selectedSpanId ?? internalSelected;
  const roots = useMemo(() => buildFlameTree(spans), [spans]);

  if (roots.length === 0) {
    return <p className="muted">No spans to render</p>;
  }

  return (
    <div className="trace-flame-graph">
      <div className="trace-flame-graph__controls">
        <Button variant="ghost" size="sm" onClick={() => setScale((s) => Math.min(3, s + 0.25))} aria-label="Zoom in">
          <Plus size={14} />
        </Button>
        <Button variant="ghost" size="sm" onClick={() => setScale((s) => Math.max(0.5, s - 0.25))} aria-label="Zoom out">
          <Minus size={14} />
        </Button>
        <Button variant="ghost" size="sm" onClick={() => setScale(1)} aria-label="Reset zoom">
          <RotateCcw size={14} />
        </Button>
        <span className="muted">{Math.round(scale * 100)}%</span>
      </div>
      <div className="trace-flame-graph__canvas" style={{ transform: `scaleX(${scale})`, transformOrigin: 'left center' }}>
        {roots.map((root) => (
          <FlameBar
            key={root.span.spanId}
            node={root}
            traceTotalMs={traceTotalMs}
            scale={scale}
            selected={root.span.spanId === selected}
            onSelect={(s) => {
              setInternalSelected(s.spanId);
              onSelectSpan?.(s);
            }}
          />
        ))}
      </div>
    </div>
  );
}
