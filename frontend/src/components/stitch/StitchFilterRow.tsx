import type { ReactNode } from 'react';

interface StitchFilterRowProps {
  children: ReactNode;
  /** Right-aligned actions (e.g. refresh, export). */
  actions?: ReactNode;
  className?: string;
}

/**
 * Horizontal filter row used across explorer/list screens — filters on the
 * left, optional actions pinned to the right. Matches the Stitch filter bars.
 */
export function StitchFilterRow({ children, actions, className }: StitchFilterRowProps) {
  return (
    <div className={`stitch-filter-row ${className ?? ''}`.trim()}>
      <div className="stitch-filter-row__filters">{children}</div>
      {actions && <div className="stitch-filter-row__actions">{actions}</div>}
    </div>
  );
}
