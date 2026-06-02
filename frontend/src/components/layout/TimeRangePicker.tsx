import { useEffect, useRef, useState } from 'react';
import { Calendar } from 'lucide-react';
import { useFilterStore, type TimeRangePreset } from '../../store/filterStore';

const PRESETS: { value: Exclude<TimeRangePreset, 'custom'>; label: string }[] = [
  { value: '1h', label: 'Last 1 hour' },
  { value: '6h', label: 'Last 6 hours' },
  { value: '24h', label: 'Last 24 hours' },
  { value: '7d', label: 'Last 7 days' },
];

const PRESET_SHORT: Record<TimeRangePreset, string> = {
  '1h': 'Last 1h',
  '6h': 'Last 6h',
  '24h': 'Last 24h',
  '7d': 'Last 7d',
  custom: 'Custom',
};

const pad = (n: number) => String(n).padStart(2, '0');

/** Format a Date into the value expected by <input type="datetime-local"> (local time). */
export function toLocalInputValue(date: Date): string {
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(
    date.getMinutes(),
  )}`;
}

const compact = (d: Date) =>
  d.toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });

export function formatRangeLabel(
  timeRange: TimeRangePreset,
  customStart: Date | null,
  customEnd: Date | null,
): string {
  if (timeRange === 'custom' && customStart && customEnd) {
    return `${compact(customStart)} → ${compact(customEnd)}`;
  }
  return PRESET_SHORT[timeRange] ?? 'Last 24h';
}

export function TimeRangePicker() {
  const timeRange = useFilterStore((s) => s.timeRange);
  const customStart = useFilterStore((s) => s.customStart);
  const customEnd = useFilterStore((s) => s.customEnd);
  const setTimeRange = useFilterStore((s) => s.setTimeRange);
  const setCustomRange = useFilterStore((s) => s.setCustomRange);
  const getTimeBounds = useFilterStore((s) => s.getTimeBounds);

  const [open, setOpen] = useState(false);
  const [startInput, setStartInput] = useState('');
  const [endInput, setEndInput] = useState('');
  const [error, setError] = useState('');
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const bounds = getTimeBounds();
    setStartInput(toLocalInputValue(customStart ?? bounds.start));
    setEndInput(toLocalInputValue(customEnd ?? bounds.end));
    setError('');
  }, [open, customStart, customEnd, getTimeBounds]);

  useEffect(() => {
    if (!open) return;
    const onPointerDown = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false);
    };
    document.addEventListener('mousedown', onPointerDown);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onPointerDown);
      document.removeEventListener('keydown', onKey);
    };
  }, [open]);

  const choosePreset = (preset: TimeRangePreset) => {
    setTimeRange(preset);
    setOpen(false);
  };

  const applyCustom = () => {
    const start = new Date(startInput);
    const end = new Date(endInput);
    if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
      setError('Enter a valid start and end time.');
      return;
    }
    if (start >= end) {
      setError('Start must be before end.');
      return;
    }
    if (end.getTime() > Date.now() + 60_000) {
      setError('End time cannot be in the future.');
      return;
    }
    setCustomRange(start, end);
    setOpen(false);
  };

  return (
    <div className="time-range-picker" ref={ref}>
      <button
        type="button"
        className="time-range-picker__trigger"
        onClick={() => setOpen((o) => !o)}
        aria-haspopup="dialog"
        aria-expanded={open}
      >
        <Calendar size={14} aria-hidden />
        <span>{formatRangeLabel(timeRange, customStart, customEnd)}</span>
      </button>

      {open && (
        <div className="time-range-popover" role="dialog" aria-label="Select time range">
          <div className="time-range-popover__presets">
            {PRESETS.map((p) => (
              <button
                key={p.value}
                type="button"
                className={`time-range-option ${timeRange === p.value ? 'time-range-option--active' : ''}`}
                onClick={() => choosePreset(p.value)}
              >
                {p.label}
              </button>
            ))}
          </div>

          <div className="time-range-popover__custom">
            <span className="time-range-popover__heading">Custom range</span>
            <label className="time-range-field">
              <span>Start</span>
              <input
                type="datetime-local"
                value={startInput}
                max={endInput || undefined}
                onChange={(e) => setStartInput(e.target.value)}
              />
            </label>
            <label className="time-range-field">
              <span>End</span>
              <input
                type="datetime-local"
                value={endInput}
                min={startInput || undefined}
                onChange={(e) => setEndInput(e.target.value)}
              />
            </label>
            {error && <p className="time-range-error">{error}</p>}
            <button type="button" className="ui-button ui-button--primary ui-button--sm" onClick={applyCustom}>
              Apply custom range
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
