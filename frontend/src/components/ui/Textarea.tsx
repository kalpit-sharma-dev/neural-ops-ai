import { forwardRef, type TextareaHTMLAttributes } from 'react';
import './ui.css';

export interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
  hint?: string;
  error?: string;
}

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(function Textarea(
  { label, hint, error, className = '', id, ...props },
  ref,
) {
  const inputId = id ?? (label ? `textarea-${label.replace(/\s+/g, '-').toLowerCase()}` : undefined);
  return (
    <div className={`ui-field ${error ? 'ui-field--error' : ''} ${className}`.trim()}>
      {label && (
        <label className="ui-field__label" htmlFor={inputId}>
          {label}
        </label>
      )}
      <textarea ref={ref} id={inputId} className="ui-textarea" aria-invalid={Boolean(error)} {...props} />
      {hint && !error && <p className="ui-field__hint">{hint}</p>}
      {error && <p className="ui-field__error">{error}</p>}
    </div>
  );
});
