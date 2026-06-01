export type LogColumnId = 'severity' | 'time' | 'service' | 'message' | 'trace';

export interface LogColumnPref {
  id: LogColumnId;
  width: number;
  visible: boolean;
}

const DEFAULT: LogColumnPref[] = [
  { id: 'severity', width: 48, visible: true },
  { id: 'time', width: 88, visible: true },
  { id: 'service', width: 120, visible: true },
  { id: 'message', width: 1, visible: true },
  { id: 'trace', width: 96, visible: true },
];

const KEY = 'neuralops-log-columns';

export function loadLogColumns(): LogColumnPref[] {
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return DEFAULT;
    const parsed = JSON.parse(raw) as LogColumnPref[];
    return Array.isArray(parsed) ? parsed : DEFAULT;
  } catch {
    return DEFAULT;
  }
}

export function saveLogColumns(cols: LogColumnPref[]) {
  localStorage.setItem(KEY, JSON.stringify(cols));
}
