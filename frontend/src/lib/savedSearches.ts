export interface SavedSearch {
  id: string;
  name: string;
  query: string;
  mode: 'text' | 'regex' | 'ai';
  createdAt: string;
}

const STORAGE_KEY = 'neuralops-saved-searches';

function storageKey(tenantId: string, userId: string) {
  return `${STORAGE_KEY}:${tenantId}:${userId}`;
}

export function listSavedSearches(tenantId: string, userId: string): SavedSearch[] {
  try {
    const raw = localStorage.getItem(storageKey(tenantId, userId));
    if (!raw) return [];
    const parsed = JSON.parse(raw) as SavedSearch[];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

export function saveSearch(
  tenantId: string,
  userId: string,
  entry: Omit<SavedSearch, 'id' | 'createdAt'> & { id?: string },
): SavedSearch {
  const list = listSavedSearches(tenantId, userId);
  const saved: SavedSearch = {
    id: entry.id ?? crypto.randomUUID(),
    name: entry.name,
    query: entry.query,
    mode: entry.mode,
    createdAt: new Date().toISOString(),
  };
  const idx = list.findIndex((s) => s.id === saved.id);
  const next = idx >= 0 ? list.map((s, i) => (i === idx ? saved : s)) : [saved, ...list].slice(0, 50);
  localStorage.setItem(storageKey(tenantId, userId), JSON.stringify(next));
  return saved;
}

export function deleteSavedSearch(tenantId: string, userId: string, id: string) {
  const next = listSavedSearches(tenantId, userId).filter((s) => s.id !== id);
  localStorage.setItem(storageKey(tenantId, userId), JSON.stringify(next));
}
