import { ReactNode } from 'react';
import './ui.css';

type BadgeVariant =
  | 'critical'
  | 'p1'
  | 'p2'
  | 'p3'
  | 'p4'
  | 'healthy'
  | 'degraded'
  | 'down'
  | 'info'
  | 'success'
  | 'warning'
  | 'error';

interface BadgeProps {
  variant?: BadgeVariant;
  children: ReactNode;
  className?: string;
}

export function Badge({ variant = 'info', children, className = '' }: BadgeProps) {
  return <span className={`ui-badge ui-badge--${variant} ${className}`.trim()}>{children}</span>;
}
