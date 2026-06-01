import * as Dialog from '@radix-ui/react-dialog';
import { X } from 'lucide-react';
import { useFilterStore, type Environment, type TimeRangePreset } from '../../store/filterStore';
import { Select } from '../ui/Select';
import { Button } from '../ui/Button';

interface MobileFilterSheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function MobileFilterSheet({ open, onOpenChange }: MobileFilterSheetProps) {
  const environment = useFilterStore((s) => s.environment);
  const timeRange = useFilterStore((s) => s.timeRange);
  const setEnvironment = useFilterStore((s) => s.setEnvironment);
  const setTimeRange = useFilterStore((s) => s.setTimeRange);

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
            <option value="1h">Last 1h</option>
            <option value="6h">Last 6h</option>
            <option value="24h">Last 24h</option>
            <option value="7d">Last 7d</option>
          </Select>
          <Button variant="primary" onClick={() => onOpenChange(false)}>
            Apply
          </Button>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
