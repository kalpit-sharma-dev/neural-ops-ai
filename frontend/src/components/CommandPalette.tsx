import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import * as Dialog from '@radix-ui/react-dialog';
import { Command, Search, X } from 'lucide-react';

interface CommandItem {
  id: string;
  label: string;
  hint?: string;
  path: string;
  keywords?: string[];
}

const COMMANDS: CommandItem[] = [
  { id: 'dashboard', label: 'Command Center', path: '/', keywords: ['home', 'overview'] },
  { id: 'logs', label: 'Log Explorer', path: '/logs', keywords: ['search', 'logs'] },
  { id: 'incidents', label: 'Incidents', path: '/incidents', keywords: ['p1', 'outage'] },
  { id: 'traces', label: 'Trace Explorer', path: '/traces' },
  { id: 'trace-compare', label: 'Compare Traces', path: '/traces/compare', keywords: ['apm', 'diff'] },
  { id: 'service-flow', label: 'Service Flow', path: '/service-flow' },
  { id: 'metrics', label: 'Metrics Explorer', path: '/metrics' },
  { id: 'dashboards', label: 'Dashboards', path: '/dashboards' },
  { id: 'service-map', label: 'Service Map', path: '/service-map', keywords: ['topology'] },
  { id: 'infrastructure', label: 'Infrastructure', path: '/infrastructure' },
  { id: 'kubernetes', label: 'Kubernetes', path: '/kubernetes' },
  { id: 'databases', label: 'Databases', path: '/databases' },
  { id: 'slos', label: 'SLOs', path: '/slos' },
  { id: 'rum', label: 'RUM', path: '/rum' },
  { id: 'synthetic', label: 'Synthetic', path: '/synthetic' },
  { id: 'workflows', label: 'Workflows', path: '/workflows' },
  { id: 'notebooks', label: 'Notebooks', path: '/notebooks' },
  { id: 'security', label: 'Security', path: '/security' },
  { id: 'ai-chat', label: 'AI Assistant', path: '/ai-chat', keywords: ['chat', 'copilot'] },
  { id: 'transactions', label: 'Transaction Journey', path: '/transactions', keywords: ['upi'] },
  { id: 'anomalies', label: 'Anomaly Detection', path: '/anomalies' },
  { id: 'alerts', label: 'Alerts', path: '/alerts' },
  { id: 'settings', label: 'Settings', path: '/settings' },
];

export function CommandPalette() {
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault();
        setOpen((value) => !value);
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, []);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return COMMANDS;
    return COMMANDS.filter(
      (item) =>
        item.label.toLowerCase().includes(q) ||
        item.path.includes(q) ||
        item.keywords?.some((k) => k.includes(q)),
    );
  }, [query]);

  const run = (item: CommandItem) => {
    navigate({ to: item.path });
    setOpen(false);
    setQuery('');
  };

  return (
    <Dialog.Root open={open} onOpenChange={setOpen}>
      <Dialog.Portal>
        <Dialog.Overlay className="command-palette-overlay" />
        <Dialog.Content className="command-palette" aria-label="Command palette">
          <div className="command-palette__header">
            <Search size={16} />
            <input
              autoFocus
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Jump to page…"
              aria-label="Search commands"
            />
            <button type="button" className="icon-btn" onClick={() => setOpen(false)} aria-label="Close">
              <X size={16} />
            </button>
          </div>
          <ul className="command-palette__list">
            {filtered.map((item) => (
              <li key={item.id}>
                <button type="button" onClick={() => run(item)}>
                  <Command size={14} />
                  <span>{item.label}</span>
                  <span className="muted">{item.path}</span>
                </button>
              </li>
            ))}
            {filtered.length === 0 && <li className="muted command-palette__empty">No matches</li>}
          </ul>
          <p className="command-palette__hint muted">Ctrl+K · Esc to close</p>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
