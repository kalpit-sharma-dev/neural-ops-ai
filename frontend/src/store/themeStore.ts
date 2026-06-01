import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type ThemeMode = 'dark' | 'light';
export type ThemePreference = ThemeMode | 'system';

const STORAGE_KEY = 'neuralops-theme';

function systemTheme(): ThemeMode {
  if (typeof window === 'undefined') return 'dark';
  return window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
}

export function resolveTheme(preference: ThemePreference): ThemeMode {
  return preference === 'system' ? systemTheme() : preference;
}

export function applyTheme(mode: ThemeMode) {
  document.documentElement.dataset.theme = mode;
}

/** Call before React paint to avoid theme flash. */
export function initTheme(): ThemeMode {
  const stored = localStorage.getItem(STORAGE_KEY);
  let preference: ThemePreference = 'system';
  if (stored === 'light' || stored === 'dark' || stored === 'system') {
    preference = stored;
  } else {
    try {
      const parsed = JSON.parse(stored ?? '{}') as { state?: { preference?: ThemePreference } };
      if (parsed.state?.preference) preference = parsed.state.preference;
    } catch {
      /* use system */
    }
  }
  const mode = resolveTheme(preference);
  applyTheme(mode);
  return mode;
}

interface ThemeState {
  preference: ThemePreference;
  theme: ThemeMode;
  setPreference: (preference: ThemePreference) => void;
  cyclePreference: () => void;
}

let mediaCleanup: (() => void) | null = null;

function watchSystemTheme(set: (partial: Partial<ThemeState>) => void, getPreference: () => ThemePreference) {
  mediaCleanup?.();
  const mq = window.matchMedia('(prefers-color-scheme: light)');
  const handler = () => {
    if (getPreference() === 'system') {
      const mode = systemTheme();
      applyTheme(mode);
      set({ theme: mode });
    }
  };
  mq.addEventListener('change', handler);
  mediaCleanup = () => mq.removeEventListener('change', handler);
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set, get) => ({
      preference: 'system',
      theme: 'dark',
      setPreference: (preference) => {
        const mode = resolveTheme(preference);
        applyTheme(mode);
        set({ preference, theme: mode });
        watchSystemTheme((partial) => set(partial), () => get().preference);
      },
      cyclePreference: () => {
        const order: ThemePreference[] = ['light', 'dark', 'system'];
        const idx = order.indexOf(get().preference);
        const next = order[(idx + 1) % order.length];
        get().setPreference(next);
      },
    }),
    {
      name: STORAGE_KEY,
      partialize: (state) => ({ preference: state.preference }),
      onRehydrateStorage: () => (state) => {
        if (!state) return;
        const mode = resolveTheme(state.preference);
        applyTheme(mode);
        state.theme = mode;
        watchSystemTheme((partial) => useThemeStore.setState(partial), () => state.preference);
      },
    },
  ),
);
