import { forwardRef, type InputHTMLAttributes } from 'react';
import './ui.css';

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  hint?: string;
  error?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(function Input(
  { label, hint, error, className = '', id, ...props },
  ref,
) {
  const inputId = id ?? (label ? `input-${label.replace(/\s+/g, '-').toLowerCase()}` : undefined);
  return (
    <div className={`ui-field ${error ? 'ui-field--error' : ''} ${className}`.trim()}>
      {label && (
        <label className="ui-field__label" htmlFor={inputId}>
          {label}
        </label>
      )}
      <input ref={ref} id={inputId} className="ui-input" aria-invalid={Boolean(error)} aria-describedby={hint ? `${inputId}-hint` : undefined} {...props} />
      {hint && !error && (
        <p className="ui-field__hint" id={`${inputId}-hint`}>
          {hint}
        </p>
      )}
      {error && <p className="ui-field__error">{error}</p>}
    </div>
  );
});
