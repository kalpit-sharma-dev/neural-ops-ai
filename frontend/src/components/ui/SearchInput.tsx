import { InputHTMLAttributes } from 'react';
import { Search } from 'lucide-react';
import './ui.css';

interface SearchInputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
  shortcut?: string;
}

export function SearchInput({ shortcut = '⌘K', className = '', ...props }: SearchInputProps) {
  return (
    <div className={`ui-search ${className}`.trim()}>
      <Search size={16} className="ui-search__icon" />
      <input type="search" className="ui-search__input" {...props} />
      {shortcut && <kbd className="ui-search__kbd">{shortcut}</kbd>}
    </div>
  );
}
