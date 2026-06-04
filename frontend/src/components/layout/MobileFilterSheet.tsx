import { useEffect, useState } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { X } from 'lucide-react';
import { useFilterStore, type Environment, type TimeRangePreset } from '../../store/filterStore';
import { TIME_RANGE_GROUPS } from '../../lib/timeRangePresets';
import { Select } from '../ui/Select';
import { Button } from '../ui/Button';
import { toLocalInputValue } from './TimeRangePicker';

interface MobileFilterSheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function MobileFilterSheet({ open, onOpenChange }: MobileFilterSheetProps) {
  const environment = useFilterStore((s) => s.environment);
  const timeRange = useFilterStore((s) => s.timeRange);
  const customStart = useFilterStore((s) => s.customStart);
  const customEnd = useFilterStore((s) => s.customEnd);
  const setEnvironment = useFilterStore((s) => s.setEnvironment);
  const setTimeRange = useFilterStore((s) => s.setTimeRange);
  const setCustomRange = useFilterStore((s) => s.setCustomRange);
  const getTimeBounds = useFilterStore((s) => s.getTimeBounds);

  const [startInput, setStartInput] = useState('');
  const [endInput, setEndInput] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    if (!open) return;
    const bounds = getTimeBounds();
    setStartInput(toLocalInputValue(customStart ?? bounds.start));
    setEndInput(toLocalInputValue(customEnd ?? bounds.end));
    setError('');
  }, [open, customStart, customEnd, getTimeBounds]);

  const handleApply = () => {
    if (timeRange === 'custom') {
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
    }
    onOpenChange(false);
  };

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="mobile-sheet-overlay" />
        <Dialog.Content className="mobile-sheet" aria-label="Global filters">
          <div className="mobile-sheet__header">
            <Dialog.Title>Filters</Dialog.Title>
            <Dialog.Close asChild>
              <button type="button" className="icon-btn" aria-label="Close filters">
                <X size={18} />
              </button>
            </Dialog.Close>
          </div>
          <Select label="Environment" value={environment} onChange={(e) => setEnvironment(e.target.value as Environment)}>
            <option value="ALL">ALL</option>
            <option value="PROD">PROD</option>
            <option value="STAGING">STAGING</option>
            <option value="DEV">DEV</option>
          </Select>
          <Select label="Time range" value={timeRange} onChange={(e) => setTimeRange(e.target.value as TimeRangePreset)}>
            {TIME_RANGE_GROUPS.flatMap((g) => g.presets).map((p) => (
              <option key={p.id} value={p.id}>
                {p.label}
              </option>
            ))}
            <option value="custom">Custom range</option>
          </Select>
          {timeRange === 'custom' && (
            <div className="time-range-popover__custom" style={{ borderTop: 'none', paddingTop: 0 }}>
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
            </div>
          )}
          <Button variant="primary" onClick={handleApply}>
            Apply
          </Button>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
