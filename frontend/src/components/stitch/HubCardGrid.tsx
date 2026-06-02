import type { ReactNode } from 'react';
import { Link } from '@tanstack/react-router';
import { ArrowRight } from 'lucide-react';

export function HubCardGrid({ children }: { children: ReactNode }) {
  return <div className="hub-card-grid">{children}</div>;
}

interface HubCardProps {
  title: string;
  description: string;
  to: string;
  /** Optional leading icon. */
  icon?: ReactNode;
  /** Optional CTA label (defaults to "Open"). */
  cta?: string;
}

/**
 * Navigation card used by hub screens (Settings overview, Marketplace).
 * Title + one-line description + an "Open →" affordance.
 */
export function HubCard({ title, description, to, icon, cta = 'Open' }: HubCardProps) {
  return (
    <Link to={to} className="hub-card">
      <div className="hub-card__head">
        {icon && <span className="hub-card__icon">{icon}</span>}
        <h3 className="hub-card__title">{title}</h3>
      </div>
      <p className="hub-card__desc">{description}</p>
      <span className="hub-card__cta">
        {cta} <ArrowRight size={14} aria-hidden />
      </span>
    </Link>
  );
}
