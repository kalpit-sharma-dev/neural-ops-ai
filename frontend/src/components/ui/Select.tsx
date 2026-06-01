import { SelectHTMLAttributes, ReactNode } from 'react';
import './ui.css';

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
  children: ReactNode;
}

export function Select({ label, children, className = '', ...props }: SelectProps) {
  return (
    <label className={`ui-select ${className}`.trim()}>
      {label && <span className="ui-select__label">{label}</span>}
      <select className="ui-select__control" {...props}>
        {children}
      </select>
    </label>
  );
}
