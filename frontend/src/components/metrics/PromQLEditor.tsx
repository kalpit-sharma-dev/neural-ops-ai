import { useEffect, useMemo, useState } from 'react';
import { fetchMetricCatalog } from '../../api/observability';
import { Textarea } from '../ui/Textarea';

const HISTORY_KEY = 'neuralops-promql-history';

interface PromQLEditorProps {
  value: string;
  onChange: (value: string) => void;
  onRun?: () => void;
  error?: string;
  label?: string;
}

export function PromQLEditor({ value, onChange, onRun, error, label = 'PromQL' }: PromQLEditorProps) {
  const [history, setHistory] = useState<string[]>(() => {
    try {
      return JSON.parse(localStorage.getItem(HISTORY_KEY) ?? '[]') as string[];
    } catch {
      return [];
    }
  });
  const [catalog, setCatalog] = useState<string[]>([]);

  useEffect(() => {
    void fetchMetricCatalog()
      .then((items) => setCatalog(items.map((m) => m.name)))
      .catch(() => undefined);
  }, []);

  const suggestions = useMemo(() => {
    const token = value.split(/\s+/).pop() ?? '';
    if (token.length < 2) return [];
    return [...catalog, ...history].filter((s) => s.includes(token)).slice(0, 8);
  }, [catalog, history, value]);

  const pushHistory = (q: string) => {
    const next = [q, ...history.filter((h) => h !== q)].slice(0, 20);
    setHistory(next);
    localStorage.setItem(HISTORY_KEY, JSON.stringify(next));
  };

  return (
    <div className="promql-editor">
      <Textarea
        label={label}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
            e.preventDefault();
            pushHistory(value);
            onRun?.();
          }
        }}
        error={error}
        hint="Ctrl+Enter to run · autocomplete from catalog"
        rows={4}
        className={error ? 'promql-editor--error' : ''}
        spellCheck={false}
      />
      {suggestions.length > 0 && (
        <ul className="promql-editor__suggestions" role="listbox">
          {suggestions.map((s) => (
            <li key={s}>
              <button type="button" onClick={() => onChange(s)}>
                {s}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
