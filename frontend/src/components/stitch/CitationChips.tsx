import { Link } from '@tanstack/react-router';
import { FileText } from 'lucide-react';
import type { ChatSource } from '../../api/chat';

type SourceLink = {
  to: string;
  params?: Record<string, string>;
  search?: Record<string, string>;
};

/** Resolve an in-app deep link for an AI citation source. */
export function citationLink(source: ChatSource): SourceLink | null {
  const msg = source.message ?? '';
  const traceMatch = msg.match(/\b(trace-[a-z0-9-]+)\b/i);
  if (traceMatch) return { to: '/traces/$traceId', params: { traceId: traceMatch[1] } };
  const incMatch = msg.match(/\b(inc-[a-z0-9-]+)\b/i);
  if (incMatch) return { to: '/incidents/$id', params: { id: incMatch[1] } };
  if (source.service) return { to: '/logs', search: { service: source.service } };
  return null;
}

/**
 * Citation list for AI Assistant answers — each source links to the relevant
 * trace, incident, or filtered logs when a target can be inferred.
 */
export function CitationChips({ sources }: { sources: ChatSource[] }) {
  if (!sources.length) return null;

  return (
    <details className="chat-sources citation-chips" open>
      <summary>
        <FileText size={13} aria-hidden /> Sources ({sources.length})
      </summary>
      <ul>
        {sources.map((source, idx) => {
          const link = citationLink(source);
          return (
            <li key={`${source.service}-${idx}`}>
              {link ? (
                <Link to={link.to} params={link.params} search={link.search}>
                  <strong>{source.service}</strong> · {source.severity}
                </Link>
              ) : (
                <span>
                  <strong>{source.service}</strong> · {source.severity}
                </span>
              )}
              <p className="muted">{source.message}</p>
            </li>
          );
        })}
      </ul>
    </details>
  );
}
