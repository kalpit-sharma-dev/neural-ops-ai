import { useMemo, useState, type ReactNode } from 'react';
import { Link } from '@tanstack/react-router';
import { ChevronDown, ChevronRight, Copy } from 'lucide-react';
import toast from 'react-hot-toast';
import { logsSearch } from '../../utils/logsSearch';
import type { Span } from '../../api/observability';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';

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
  collapsed,
  hasChildren,
  onToggleCollapse,
  onSelect,
  onCopyId,
  children,
}: {
  span: Span;
  depth: number;
  traceTotalMs: number;
  selected: boolean;
  collapsed: boolean;
  hasChildren: boolean;
  onToggleCollapse: () => void;
  onSelect: () => void;
  onCopyId: () => void;
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
          {hasChildren ? (
            <button
              type="button"
              className="trace-waterfall__toggle"
              aria-label={collapsed ? 'Expand' : 'Collapse'}
              onClick={(e) => {
                e.stopPropagation();
                onToggleCollapse();
              }}
            >
              {collapsed ? <ChevronRight size={14} /> : <ChevronDown size={14} />}
            </button>
          ) : (
            <span className="trace-waterfall__toggle-spacer" />
          )}
          <Badge variant={statusVariant}>{span.service}</Badge>
          <span className="trace-waterfall__op">{span.operation}</span>
          <span className="muted">{span.durationMs}ms</span>
          <Button
            variant="ghost"
            size="sm"
            onClick={(e) => {
              e.stopPropagation();
              onCopyId();
            }}
            aria-label="Copy span ID"
          >
            <Copy size={12} />
          </Button>
        </div>
        <div className="trace-waterfall__bar-track">
          <div
            className={`trace-waterfall__bar trace-waterfall__bar--${span.status.toLowerCase()}`}
            style={{ width: `${widthPct}%` }}
          />
        </div>
      </div>
      {!collapsed && children}
    </>
  );
}

function renderSpanTree(
  span: Span,
  depth: number,
  traceTotalMs: number,
  childrenMap: Map<string, Span[]>,
  collapsedIds: Set<string>,
  selectedSpanId: string | undefined,
  onToggleCollapse: (id: string) => void,
  onSelectSpan?: (span: Span) => void,
  onCopyId?: (id: string) => void,
): ReactNode {
  const kids = childrenMap.get(span.spanId) ?? [];
  const collapsed = collapsedIds.has(span.spanId);
  return (
    <SpanRow
      key={span.spanId}
      span={span}
      depth={depth}
      traceTotalMs={traceTotalMs}
      selected={selectedSpanId === span.spanId}
      collapsed={collapsed}
      hasChildren={kids.length > 0}
      onToggleCollapse={() => onToggleCollapse(span.spanId)}
      onSelect={() => onSelectSpan?.(span)}
      onCopyId={() => onCopyId?.(span.spanId)}
      children={
        !collapsed
          ? kids.map((child) =>
              renderSpanTree(
                child,
                depth + 1,
                traceTotalMs,
                childrenMap,
                collapsedIds,
                selectedSpanId,
                onToggleCollapse,
                onSelectSpan,
                onCopyId,
              ),
            )
          : null
      }
    />
  );
}

export function TraceWaterfall({ spans, traceTotalMs, traceId, selectedSpanId, onSelectSpan }: TraceWaterfallProps) {
  const [internalSelected, setInternalSelected] = useState<string | undefined>();
  const [collapsedIds, setCollapsedIds] = useState<Set<string>>(new Set());
  const selected = selectedSpanId ?? internalSelected;
  const { roots, children } = useMemo(() => buildTree(spans), [spans]);

  const handleSelect = (span: Span) => {
    setInternalSelected(span.spanId);
    onSelectSpan?.(span);
  };

  const toggleCollapse = (id: string) => {
    setCollapsedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const copySpanId = async (id: string) => {
    await navigator.clipboard.writeText(id);
    toast.success('Span ID copied');
  };

  const selectedSpan = spans.find((s) => s.spanId === selected);
  const effectiveTraceId = traceId ?? selectedSpan?.traceId;

  return (
    <div className="trace-waterfall">
      <div className="trace-waterfall__header">
        <span>Span tree</span>
        <span className="muted">{spans.length} spans · {traceTotalMs}ms total</span>
        <Button variant="ghost" size="sm" onClick={() => setCollapsedIds(new Set(spans.map((s) => s.spanId)))}>
          Collapse all
        </Button>
        <Button variant="ghost" size="sm" onClick={() => setCollapsedIds(new Set())}>
          Expand all
        </Button>
      </div>
      <div className="trace-waterfall__body">
        {roots.map((root) =>
          renderSpanTree(
            root,
            0,
            traceTotalMs,
            children,
            collapsedIds,
            selected,
            toggleCollapse,
            handleSelect,
            copySpanId,
          ),
        )}
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
              traceId: effectiveTraceId,
              from: new Date(new Date(selectedSpan.startTime).getTime() - 5000).toISOString(),
              to: new Date(new Date(selectedSpan.startTime).getTime() + selectedSpan.durationMs + 5000).toISOString(),
            })}
            className="trace-waterfall__logs-link"
          >
            View logs (traceId={effectiveTraceId?.slice(0, 12)}…) →
          </Link>
        </div>
      )}
    </div>
  );
}
