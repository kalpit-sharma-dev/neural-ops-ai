import { useEffect, useRef, useState } from 'react';
import { Calendar, ChevronLeft, ChevronRight } from 'lucide-react';
import { useFilterStore, type TimeRangePreset } from '../../store/filterStore';
import { TIME_RANGE_GROUPS, presetShortLabel } from '../../lib/timeRangePresets';

const pad = (n: number) => String(n).padStart(2, '0');

/** Format a Date into the value expected by <input type="datetime-local"> (local time). */
export function toLocalInputValue(date: Date): string {
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(
    date.getMinutes(),
  )}`;
}

const compact = (d: Date) =>
  d.toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });

function timezoneAbbr(): string {
  try {
    const parts = new Intl.DateTimeFormat(undefined, { timeZoneName: 'short' }).formatToParts(new Date());
    return parts.find((p) => p.type === 'timeZoneName')?.value ?? 'local';
  } catch {
    return 'local';
  }
}

export function formatRangeLabel(
  timeRange: TimeRangePreset,
  customStart: Date | null,
  customEnd: Date | null,
): string {
  if (timeRange === 'custom' && customStart && customEnd) {
    return `${compact(customStart)} → ${compact(customEnd)}`;
  }
  return presetShortLabel(timeRange);
}

export function TimeRangePicker() {
  const timeRange = useFilterStore((s) => s.timeRange);
  const customStart = useFilterStore((s) => s.customStart);
  const customEnd = useFilterStore((s) => s.customEnd);
  const setTimeRange = useFilterStore((s) => s.setTimeRange);
  const setCustomRange = useFilterStore((s) => s.setCustomRange);
  const shiftTimeRange = useFilterStore((s) => s.shiftTimeRange);
  const snapTimeRangeToNow = useFilterStore((s) => s.snapTimeRangeToNow);
  const getTimeBounds = useFilterStore((s) => s.getTimeBounds);

  const [open, setOpen] = useState(false);
  const [startInput, setStartInput] = useState('');
  const [endInput, setEndInput] = useState('');
  const [error, setError] = useState('');
  const ref = useRef<HTMLDivElement>(null);
  const tz = timezoneAbbr();

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
      <div className="time-range-picker__cluster">
        <button
          type="button"
          className="time-range-picker__nudge"
          aria-label="Shift time range earlier"
          onClick={() => shiftTimeRange(-1)}
        >
          <ChevronLeft size={14} />
        </button>
        <button
          type="button"
          className="time-range-picker__trigger"
          onClick={() => setOpen((o) => !o)}
          aria-haspopup="dialog"
          aria-expanded={open}
        >
          <Calendar size={14} aria-hidden />
          <span>{formatRangeLabel(timeRange, customStart, customEnd)}</span>
          <span className="time-range-picker__tz">{tz}</span>
        </button>
        <button
          type="button"
          className="time-range-picker__nudge"
          aria-label="Shift time range later"
          onClick={() => shiftTimeRange(1)}
        >
          <ChevronRight size={14} />
        </button>
      </div>

      {open && (
        <div className="time-range-popover time-range-popover--wide" role="dialog" aria-label="Select time range">
          <div className="time-range-popover__toolbar">
            <button type="button" className="ui-button ui-button--ghost ui-button--sm" onClick={() => snapTimeRangeToNow()}>
              Snap to now
            </button>
          </div>

          {TIME_RANGE_GROUPS.map((group) => (
            <div key={group.id} className="time-range-popover__group">
              <span className="time-range-popover__heading">{group.label}</span>
              <div className="time-range-popover__presets">
                {group.presets.map((p) => (
                  <button
                    key={p.id}
                    type="button"
                    className={`time-range-option ${timeRange === p.id ? 'time-range-option--active' : ''}`}
                    onClick={() => choosePreset(p.id)}
                  >
                    {p.label}
                  </button>
                ))}
              </div>
            </div>
          ))}

          <div className="time-range-popover__custom">
            <span className="time-range-popover__heading">Custom range ({tz})</span>
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
