import { ReactNode } from 'react';
import './ui.css';

interface FormFieldProps {
  label: string;
  hint?: string;
  error?: string;
  htmlFor?: string;
  children: ReactNode;
  className?: string;
}

/** Wraps Select or custom controls with consistent label spacing. */
export function FormField({ label, hint, error, htmlFor, children, className = '' }: FormFieldProps) {
  return (
    <div className={`ui-field ${error ? 'ui-field--error' : ''} ${className}`.trim()}>
      <label className="ui-field__label" htmlFor={htmlFor}>
        {label}
      </label>
      {children}
      {hint && !error && <p className="ui-field__hint">{hint}</p>}
      {error && <p className="ui-field__error">{error}</p>}
    </div>
  );
}
