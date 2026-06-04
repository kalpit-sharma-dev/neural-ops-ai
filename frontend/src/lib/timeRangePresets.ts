import type { TimeRangePreset } from '../store/filterStore';

export interface TimeRangePresetDef {
  id: TimeRangePreset;
  label: string;
  short: string;
}

export interface TimeRangePresetGroup {
  id: string;
  label: string;
  presets: TimeRangePresetDef[];
}

/** Single source of truth for global time range presets. */
export const TIME_RANGE_GROUPS: TimeRangePresetGroup[] = [
  {
    id: 'calendar',
    label: 'Calendar',
    presets: [
      { id: 'today', label: 'Today', short: 'Today' },
      { id: 'yesterday', label: 'Yesterday', short: 'Yesterday' },
    ],
  },
  {
    id: 'seconds',
    label: 'Seconds',
    presets: [
      { id: '15s', label: 'Last 15 seconds', short: '15s' },
      { id: '30s', label: 'Last 30 seconds', short: '30s' },
    ],
  },
  {
    id: 'minutes',
    label: 'Minutes',
    presets: [
      { id: '1m', label: 'Last 1 minute', short: '1m' },
      { id: '5m', label: 'Last 5 minutes', short: '5m' },
      { id: '10m', label: 'Last 10 minutes', short: '10m' },
      { id: '15m', label: 'Last 15 minutes', short: '15m' },
      { id: '30m', label: 'Last 30 minutes', short: '30m' },
      { id: '45m', label: 'Last 45 minutes', short: '45m' },
    ],
  },
  {
    id: 'hours',
    label: 'Hours',
    presets: [
      { id: '1h', label: 'Last 1 hour', short: '1h' },
      { id: '3h', label: 'Last 3 hours', short: '3h' },
      { id: '6h', label: 'Last 6 hours', short: '6h' },
      { id: '12h', label: 'Last 12 hours', short: '12h' },
      { id: '24h', label: 'Last 24 hours', short: '24h' },
    ],
  },
  {
    id: 'days',
    label: 'Days',
    presets: [
      { id: '1d', label: 'Last 1 day', short: '1d' },
      { id: '2d', label: 'Last 2 days', short: '2d' },
      { id: '7d', label: 'Last 7 days', short: '7d' },
      { id: '14d', label: 'Last 14 days', short: '14d' },
      { id: '30d', label: 'Last 30 days', short: '30d' },
    ],
  },
  {
    id: 'extended',
    label: 'Extended',
    presets: [
      { id: '3d', label: 'Last 3 days', short: '3d' },
      { id: '5d', label: 'Last 5 days', short: '5d' },
      { id: '90d', label: 'Last 90 days', short: '90d' },
    ],
  },
];

export const ALL_TIME_RANGE_PRESETS: TimeRangePresetDef[] = TIME_RANGE_GROUPS.flatMap((g) => g.presets);

export const TIME_RANGE_PRESET_IDS = new Set<string>([
  ...ALL_TIME_RANGE_PRESETS.map((p) => p.id),
  'custom',
]);

export const TIME_RANGE_SHORT: Record<string, string> = Object.fromEntries(
  ALL_TIME_RANGE_PRESETS.map((p) => [p.id, p.short]),
);

export function presetLabel(id: TimeRangePreset): string {
  if (id === 'custom') return 'Custom';
  return ALL_TIME_RANGE_PRESETS.find((p) => p.id === id)?.label ?? id;
}

export function presetShortLabel(id: TimeRangePreset): string {
  if (id === 'custom') return 'Custom';
  return TIME_RANGE_SHORT[id] ?? id;
}

/** Sidebar quick picks (compact). */
export const LOG_SIDEBAR_TIME_PRESETS: TimeRangePreset[] = [
  '15m',
  '1h',
  '6h',
  'today',
  '1d',
  '7d',
  '30d',
];
