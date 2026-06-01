import type { ReactNode } from 'react';

export function highlightMatches(text: string, query: string, useRegex = false): ReactNode {
  if (!query.trim() || query.length < 2) return text;

  let pattern: RegExp;
  try {
    pattern = useRegex ? new RegExp(`(${query})`, 'gi') : new RegExp(`(${query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
  } catch {
    return text;
  }

  const parts = text.split(pattern);

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
