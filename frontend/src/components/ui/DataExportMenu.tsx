import { useState } from 'react';
import { ChevronDown, Download } from 'lucide-react';
import toast from 'react-hot-toast';
import { exportCsv, exportJson, exportNdjson } from '../../utils/dataExport';
import { Button } from './Button';

export type ExportFormat = 'json' | 'csv' | 'ndjson';

interface DataExportMenuProps {
  /** Static data or lazy resolver (e.g. after refetch). */
  getData: () => unknown[] | null | undefined;
  filenamePrefix: string;
  meta?: Record<string, unknown>;
  formats?: ExportFormat[];
  label?: string;
  disabled?: boolean;
  size?: 'sm' | 'md';
}

export function DataExportMenu({
  getData,
  filenamePrefix,
  meta,
  formats = ['json', 'csv', 'ndjson'],
  label = 'Download',
  disabled,
  size = 'sm',
}: DataExportMenuProps) {
  const [open, setOpen] = useState(false);

  const run = (format: ExportFormat) => {
    const rows = getData();
    if (!rows?.length) {
      toast.error('No data to export');
      return;
    }
    switch (format) {
      case 'json':
        exportJson(rows, filenamePrefix, meta);
        break;
      case 'csv':
        exportCsv(rows, filenamePrefix);
        break;
      case 'ndjson':
        exportNdjson(rows, filenamePrefix);
        break;
    }
    toast.success(`Exported ${rows.length} row(s) as ${format.toUpperCase()}`);
    setOpen(false);
  };

  return (
    <div className="data-export-menu">
      <Button
        variant="secondary"
        size={size}
        disabled={disabled}
        onClick={() => setOpen((o) => !o)}
        aria-haspopup="menu"
        aria-expanded={open}
      >
        <Download size={14} aria-hidden />
        {label}
        <ChevronDown size={12} aria-hidden />
      </Button>
      {open && (
        <div className="data-export-menu__dropdown" role="menu">
          {formats.map((fmt) => (
            <button key={fmt} type="button" role="menuitem" onClick={() => run(fmt)}>
              {fmt.toUpperCase()}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
