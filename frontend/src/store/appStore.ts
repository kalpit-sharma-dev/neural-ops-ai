import { create } from 'zustand';

interface AppState {
  lastHealthChecks: Record<string, string>;
  setLastHealthCheck: (service: string, timestamp: string) => void;
}

export const useAppStore = create<AppState>((set) => ({
  lastHealthChecks: {},
  setLastHealthCheck: (service, timestamp) =>
    set((state) => ({
      lastHealthChecks: { ...state.lastHealthChecks, [service]: timestamp },
    })),
}));
