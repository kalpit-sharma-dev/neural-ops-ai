import type { CSSProperties, ReactNode } from 'react';

interface KpiRowProps {
  children: ReactNode;
  /** Number of columns on wide viewports (default 4). */
  columns?: number;
  className?: string;
}

/**
 * Responsive grid for KPI / metric cards. Mirrors the KPI rows in the Stitch
 * Command Center and other dashboards.
 */
export function KpiRow({ children, columns = 4, className }: KpiRowProps) {
  return (
    <div
      className={`kpi-row ${className ?? ''}`.trim()}
      style={{ '--kpi-cols': columns } as CSSProperties}
    >
      {children}
    </div>
  );
}
