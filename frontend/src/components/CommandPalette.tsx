import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import * as Dialog from '@radix-ui/react-dialog';
import { Command, Plus, Search, X } from 'lucide-react';
import { ALL_NAV_ITEMS } from '../lib/navConfig';
import { filterNavByFeature } from '../lib/featureFlags';
import { useCommandPaletteStore } from '../store/commandPaletteStore';

interface CommandItem {
  id: string;
  label: string;
  hint?: string;
  path?: string;
  action?: () => void;
  keywords?: string[];
}

const ACTION_COMMANDS: Omit<CommandItem, 'action'>[] = [
  { id: 'new-dashboard', label: 'New dashboard', path: '/dashboards', keywords: ['create'] },
  { id: 'new-slo', label: 'Create SLO', path: '/slos', keywords: ['create'] },
  { id: 'new-workflow', label: 'Create workflow', path: '/workflows/editor', keywords: ['create'] },
  { id: 'new-notebook', label: 'New notebook', path: '/notebooks', keywords: ['create'] },
  { id: 'new-alert', label: 'Create alert rule', path: '/alerts', keywords: ['create'] },
];

export function CommandPalette() {
  const navigate = useNavigate();
  const open = useCommandPaletteStore((s) => s.open);
  const setOpen = useCommandPaletteStore((s) => s.setOpen);
  const [query, setQuery] = useState('');

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault();
        useCommandPaletteStore.getState().toggle();
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, []);

  const commands = useMemo<CommandItem[]>(() => {
    const nav = filterNavByFeature(ALL_NAV_ITEMS).map((item) => ({
      id: item.id,
      label: item.label,
      path: item.to,
      keywords: item.keywords,
      hint: item.to,
    }));
    return [...nav, ...ACTION_COMMANDS];
  }, []);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return commands;
    return commands.filter(
      (item) =>
        item.label.toLowerCase().includes(q) ||
        item.path?.includes(q) ||
        item.keywords?.some((k) => k.includes(q)),
    );
  }, [commands, query]);

  const run = (item: CommandItem) => {
    if (item.action) item.action();
    else if (item.path) navigate({ to: item.path });
    setOpen(false);
    setQuery('');
  };

  return (
    <Dialog.Root open={open} onOpenChange={setOpen}>
      <Dialog.Portal>
        <Dialog.Overlay className="command-palette-overlay" />
        <Dialog.Content className="command-palette" aria-label="Command palette" role="dialog">
          <div className="command-palette__header">
            <Search size={16} aria-hidden />
            <input
              autoFocus
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Jump to page or action…"
              aria-label="Search commands"
            />
            <button type="button" className="icon-btn" onClick={() => setOpen(false)} aria-label="Close palette">
              <X size={16} />
            </button>
          </div>
          <ul className="command-palette__list" role="listbox">
            {filtered.map((item) => (
              <li key={item.id} role="option">
                <button type="button" onClick={() => run(item)}>
                  {item.id.startsWith('new-') ? <Plus size={14} /> : <Command size={14} />}
                  <span>{item.label}</span>
                  <span className="muted">{item.hint ?? item.path}</span>
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
