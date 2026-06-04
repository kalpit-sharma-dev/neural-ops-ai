import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type TableDensity = 'comfortable' | 'compact';
export type RefreshIntervalMs = 0 | 30_000 | 60_000 | 300_000;

const REFRESH_OPTIONS: { value: RefreshIntervalMs; label: string }[] = [
  { value: 0, label: 'Off' },
  { value: 30_000, label: '30s' },
  { value: 60_000, label: '1m' },
  { value: 300_000, label: '5m' },
];

export { REFRESH_OPTIONS };

function applyDensity(density: TableDensity) {
  document.documentElement.dataset.density = density;
}

interface UiPreferencesState {
  density: TableDensity;
  refreshIntervalMs: RefreshIntervalMs;
  refreshGeneration: number;
  lastRefreshedAt: number;
  setDensity: (density: TableDensity) => void;
  cycleDensity: () => void;
  setRefreshIntervalMs: (ms: RefreshIntervalMs) => void;
  tickRefresh: () => void;
}

export const useUiPreferencesStore = create<UiPreferencesState>()(
  persist(
    (set, get) => ({
      density: 'comfortable',
      refreshIntervalMs: 0,
      refreshGeneration: 0,
      lastRefreshedAt: Date.now(),
      setDensity: (density) => {
        applyDensity(density);
        set({ density });
      },
      cycleDensity: () => {
        const next = get().density === 'comfortable' ? 'compact' : 'comfortable';
        get().setDensity(next);
      },
      setRefreshIntervalMs: (refreshIntervalMs) => set({ refreshIntervalMs }),
      tickRefresh: () =>
        set((s) => ({
          refreshGeneration: s.refreshGeneration + 1,
          lastRefreshedAt: Date.now(),
        })),
    }),
    {
      name: 'neuralops.ui-preferences',
      partialize: (s) => ({ density: s.density, refreshIntervalMs: s.refreshIntervalMs }),
      onRehydrateStorage: () => (state) => {
        if (state) applyDensity(state.density);
      },
    },
  ),
);
