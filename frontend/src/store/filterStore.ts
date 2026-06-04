import { create } from 'zustand';
import { endOfDay, startOfDay, subDays, subHours, subMinutes, subSeconds } from 'date-fns';
import { TIME_RANGE_PRESET_IDS } from '../lib/timeRangePresets';

export type Environment = 'ALL' | 'PROD' | 'STAGING' | 'DEV';

export type TimeRangePreset =
  | 'today'
  | 'yesterday'
  | '15s'
  | '30s'
  | '1m'
  | '5m'
  | '10m'
  | '15m'
  | '30m'
  | '45m'
  | '1h'
  | '3h'
  | '6h'
  | '12h'
  | '24h'
  | '1d'
  | '2d'
  | '3d'
  | '5d'
  | '7d'
  | '14d'
  | '30d'
  | '90d'
  | 'custom';

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
  shiftTimeRange: (direction: -1 | 1) => void;
  snapTimeRangeToNow: () => void;
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

export function presetStartFromEnd(end: Date, preset: TimeRangePreset): Date {
  switch (preset) {
    case '15s':
      return subSeconds(end, 15);
    case '30s':
      return subSeconds(end, 30);
    case '1m':
      return subMinutes(end, 1);
    case '5m':
      return subMinutes(end, 5);
    case '10m':
      return subMinutes(end, 10);
    case '15m':
      return subMinutes(end, 15);
    case '30m':
      return subMinutes(end, 30);
    case '45m':
      return subMinutes(end, 45);
    case '1h':
      return subHours(end, 1);
    case '3h':
      return subHours(end, 3);
    case '6h':
      return subHours(end, 6);
    case '12h':
      return subHours(end, 12);
    case '24h':
      return subHours(end, 24);
    case '1d':
      return subDays(end, 1);
    case '2d':
      return subDays(end, 2);
    case '3d':
      return subDays(end, 3);
    case '5d':
      return subDays(end, 5);
    case '7d':
      return subDays(end, 7);
    case '14d':
      return subDays(end, 14);
    case '30d':
      return subDays(end, 30);
    case '90d':
      return subDays(end, 90);
    default:
      return subHours(end, 24);
  }
}

export function isValidTimeRangePreset(value: string): value is TimeRangePreset {
  return TIME_RANGE_PRESET_IDS.has(value);
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
  setTimeRange: (timeRange) => set({ timeRange, customStart: null, customEnd: null }),
  setCustomRange: (customStart, customEnd) =>
    set({ customStart, customEnd, timeRange: 'custom' }),
  shiftTimeRange: (direction) => {
    const { start, end } = get().getTimeBounds();
    const width = end.getTime() - start.getTime();
    const delta = direction * width;
    const newEnd = new Date(Math.min(Date.now(), end.getTime() + delta));
    const newStart = new Date(start.getTime() + delta);
    set({ customStart: newStart, customEnd: newEnd, timeRange: 'custom' });
  },
  snapTimeRangeToNow: () => {
    const state = get();
    const end = new Date();
    if (state.timeRange === 'custom' && state.customStart) {
      const width = (state.customEnd ?? end).getTime() - state.customStart.getTime();
      set({ customEnd: end, customStart: new Date(end.getTime() - width), timeRange: 'custom' });
      return;
    }
    if (state.timeRange === 'today' || state.timeRange === 'yesterday') {
      return;
    }
    set({ customStart: null, customEnd: null });
  },
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
    const now = new Date();

    if (state.timeRange === 'custom' && state.customStart) {
      return { start: state.customStart, end: state.customEnd ?? now };
    }
    if (state.timeRange === 'today') {
      return { start: startOfDay(now), end: now };
    }
    if (state.timeRange === 'yesterday') {
      const day = subDays(now, 1);
      return { start: startOfDay(day), end: endOfDay(day) };
    }

    const end = now;
    const start = presetStartFromEnd(end, state.timeRange);
    return { start, end };
  },
}));

export { MY_SERVICES };
