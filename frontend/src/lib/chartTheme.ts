import type { CSSProperties } from 'react';

export function chartAxisProps() {
  return {
    stroke: 'var(--text-muted)',
    fontSize: 11,
    tickLine: false,
  };
}

export function chartGridProps() {
  return {
    strokeDasharray: '3 3',
    stroke: 'var(--border-subtle)',
  };
}

export function chartTooltipStyle(): CSSProperties {
  return {
    background: 'var(--bg-elevated)',
    border: '1px solid var(--border-subtle)',
    borderRadius: 'var(--radius-md)',
    color: 'var(--text-primary)',
    fontSize: 12,
  };
}

export function chartCartesianDefaults() {
  return {
    axis: chartAxisProps(),
    grid: chartGridProps(),
    tooltipStyle: chartTooltipStyle(),
  };
}
