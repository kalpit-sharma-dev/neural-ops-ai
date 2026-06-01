import { ReactNode } from 'react';
import { AlertCircle } from 'lucide-react';
import './ui.css';

interface ErrorStateProps {
  message: string;
  onRetry?: () => void;
}

export function ErrorState({ message, onRetry }: ErrorStateProps) {
  return (
    <div className="ui-error">
      <AlertCircle size={20} />
      <p>{message}</p>
      {onRetry && (
        <button type="button" className="ui-error__retry" onClick={onRetry}>
          Retry
        </button>
      )}
    </div>
  );
}

export function LoadingState({ label = 'Loading…' }: { label?: string }) {
  return (
    <div className="ui-loading" role="status">
      <span className="ui-loading__spinner" />
      {label}
    </div>
  );
}

export function PageHeader({ title, subtitle, actions }: { title: string; subtitle?: string; actions?: ReactNode }) {
  return (
    <header className="page-header">
      <div>
        <h1>{title}</h1>
        {subtitle && <p>{subtitle}</p>}
      </div>
      {actions}
    </header>
  );
}
