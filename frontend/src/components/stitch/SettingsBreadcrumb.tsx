import { Link } from '@tanstack/react-router';
import { ChevronRight } from 'lucide-react';

interface SettingsBreadcrumbProps {
  /** Current leaf page label, e.g. "Users". */
  page: string;
  /** Optional root label/route (defaults to Settings hub). */
  rootLabel?: string;
  rootTo?: string;
}

/**
 * @deprecated Use global `PageNavigationBar` in Layout (back + full route trail).
 * Kept for optional per-page overrides via StitchPageShell `breadcrumb` prop.
 */
export function SettingsBreadcrumb({
  page,
  rootLabel = 'Settings',
  rootTo = '/settings',
}: SettingsBreadcrumbProps) {
  return (
    <nav className="settings-breadcrumb" aria-label="Breadcrumb">
      <Link to={rootTo} className="settings-breadcrumb__link">
        {rootLabel}
      </Link>
      <ChevronRight size={14} aria-hidden className="settings-breadcrumb__sep" />
      <span className="settings-breadcrumb__current" aria-current="page">
        {page}
      </span>
    </nav>
  );
}
