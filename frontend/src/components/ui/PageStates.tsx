import { ReactNode } from 'react';
import { AlertCircle } from 'lucide-react';
import { useI18n } from '../../i18n/I18nProvider';
import './ui.css';

interface ErrorStateProps {
  message: string;
  onRetry?: () => void;
}

export function ErrorState({ message, onRetry }: ErrorStateProps) {
  const { t } = useI18n();
  return (
    <div className="ui-error">
      <AlertCircle size={20} />
      <p>{message}</p>
      {onRetry && (
        <button type="button" className="ui-error__retry" onClick={onRetry}>
          {t('common.retry')}
        </button>
      )}
    </div>
  );
}

export function LoadingState({ label }: { label?: string }) {
  const { t } = useI18n();
  return (
    <div className="ui-loading" role="status">
      <span className="ui-loading__spinner" />
      {label ?? t('common.loading')}
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
