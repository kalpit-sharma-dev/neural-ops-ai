import type { ReactNode } from 'react';
import { useRouterState } from '@tanstack/react-router';
import { ExternalLink } from 'lucide-react';
import { PageHeader } from '../ui/PageStates';
import { getStitchMeta, stitchPreviewUrl } from '../../lib/stitchPageMeta';

interface StitchPageShellProps {
  title: string;
  subtitle?: string;
  /** Header right-aligned actions. */
  actions?: ReactNode;
  /** Optional breadcrumb rendered above the header. */
  breadcrumb?: ReactNode;
  /** Optional filter row rendered below the header. */
  filters?: ReactNode;
  children: ReactNode;
}

/**
 * Standard page chrome aligned with the NeuralOps Stitch designs:
 * optional breadcrumb, page header (title/subtitle/actions), optional filter
 * row, then page content.
 *
 * In dev builds it also surfaces a subtle link to the canonical Stitch screen
 * for the current route (resolved from stitchPageMeta) to aid visual QA.
 */
export function StitchPageShell({
  title,
  subtitle,
  actions,
  breadcrumb,
  filters,
  children,
}: StitchPageShellProps) {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const meta = getStitchMeta(pathname);
  const showDesignRef = import.meta.env.DEV && meta?.screenId;

  const headerActions =
    showDesignRef || actions ? (
      <div className="stitch-page__actions">
        {actions}
        {showDesignRef && (
          <a
            className="stitch-page__design-ref"
            href={stitchPreviewUrl(meta!.screenId!)}
            target="_blank"
            rel="noreferrer"
            title={`Open Stitch design: ${meta!.title}`}
          >
            <ExternalLink size={13} aria-hidden /> Design
          </a>
        )}
      </div>
    ) : undefined;

  return (
    <div className="stitch-page">
      {breadcrumb}
      <PageHeader title={title} subtitle={subtitle} actions={headerActions} />
      {filters && <div className="stitch-page__filters">{filters}</div>}
      {children}
    </div>
  );
}
