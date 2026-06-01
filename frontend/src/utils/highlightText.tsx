import type { ReactNode } from 'react';

export function highlightMatches(text: string, query: string): ReactNode {
  if (!query.trim() || query.length < 2) return text;

  const escaped = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const parts = text.split(new RegExp(`(${escaped})`, 'gi'));

  return parts.map((part, i) =>
    part.toLowerCase() === query.toLowerCase() ? (
      <mark key={i} className="search-highlight">
        {part}
      </mark>
    ) : (
      part
    ),
  );
}

export function hasStackTrace(message: string): boolean {
  return /\n\s+at\s/.test(message) || /Exception|Error:/.test(message);
}
