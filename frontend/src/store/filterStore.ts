import { create } from 'zustand';
import { subHours } from 'date-fns';

export type Environment = 'ALL' | 'PROD' | 'STAGING' | 'DEV';
export type TimeRangePreset = '1h' | '6h' | '24h' | '7d' | 'custom';

const MY_SERVICES = ['payment-api', 'upi-service', 'auth-service', 'ledger-service'];

interface FilterState {
  environment: Environment;
  timeRange: TimeRangePreset;
  customStart: Date | null;
  customEnd: Date | null;
  services: string[];
  severities: string[];
  host: string;
  pod: string;
  hasStackTrace: boolean;
  hasAIExplanation: boolean;
  myServicesOnly: boolean;
  setEnvironment: (env: Environment) => void;
  setTimeRange: (range: TimeRangePreset) => void;
  setCustomRange: (start: Date, end: Date) => void;
  setHost: (host: string) => void;
  setPod: (pod: string) => void;
  toggleService: (service: string) => void;
  toggleSeverity: (severity: string) => void;
  setHasStackTrace: (value: boolean) => void;
  setHasAIExplanation: (value: boolean) => void;
  setMyServicesOnly: (value: boolean) => void;
  setErrorsOnly: () => void;
  clearServices: () => void;
  getTimeBounds: () => { start: Date; end: Date };
}

function presetToHours(preset: TimeRangePreset): number {
  switch (preset) {
    case '1h':
      return 1;
    case '6h':
      return 6;
    case '24h':
      return 24;
    case '7d':
      return 168;
    default:
      return 24;
  }
}

export const useFilterStore = create<FilterState>((set, get) => ({
  environment: 'ALL',
  timeRange: '24h',
  customStart: null,
  customEnd: null,
  services: [],
  severities: [],
  host: '',
  pod: '',
  hasStackTrace: false,
  hasAIExplanation: false,
  myServicesOnly: false,

  setEnvironment: (environment) => set({ environment }),
  setTimeRange: (timeRange) => set({ timeRange }),
  setCustomRange: (customStart, customEnd) =>
    set({ customStart, customEnd, timeRange: 'custom' }),
  setHost: (host) => set({ host }),
  setPod: (pod) => set({ pod }),
  toggleService: (service) =>
    set((state) => ({
      services: state.services.includes(service)
        ? state.services.filter((s) => s !== service)
        : [...state.services, service],
    })),
  toggleSeverity: (severity) =>
    set((state) => ({
      severities: state.severities.includes(severity)
        ? state.severities.filter((s) => s !== severity)
        : [...state.severities, severity],
    })),
  setHasStackTrace: (hasStackTrace) => set({ hasStackTrace }),
  setHasAIExplanation: (hasAIExplanation) => set({ hasAIExplanation }),
  setMyServicesOnly: (myServicesOnly) =>
    set((state) => ({
      myServicesOnly,
      services: myServicesOnly ? [...MY_SERVICES] : state.services.filter((s) => !MY_SERVICES.includes(s)),
    })),
  setErrorsOnly: () => set({ severities: ['ERROR', 'FATAL'] }),
  clearServices: () => set({ services: [] }),

  getTimeBounds: () => {
    const state = get();
    const end = state.customEnd ?? new Date();
    const start =
      state.timeRange === 'custom' && state.customStart
        ? state.customStart
        : subHours(end, presetToHours(state.timeRange));
    return { start, end };
  },
}));

export { MY_SERVICES };
