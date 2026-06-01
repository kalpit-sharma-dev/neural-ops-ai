import { CSSProperties, ReactNode } from 'react';
import './ui.css';

interface CardProps {
  title?: string;
  action?: ReactNode;
  children: ReactNode;
  className?: string;
  hover?: boolean;
  style?: CSSProperties;
}

export function Card({ title, action, children, className = '', hover = false, style }: CardProps) {
  return (
    <article className={`ui-card ${hover ? 'ui-card--hover' : ''} ${className}`.trim()} style={style}>
      {(title || action) && (
        <header className="ui-card__header">
          {title && <h3 className="ui-card__title">{title}</h3>}
          {action}
        </header>
      )}
      <div className="ui-card__body">{children}</div>
    </article>
  );
}
