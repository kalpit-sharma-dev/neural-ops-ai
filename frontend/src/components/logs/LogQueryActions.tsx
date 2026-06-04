import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { Bell, BarChart3, ChevronDown, Copy, Download, Filter, Link2, MinusCircle } from 'lucide-react';
import toast from 'react-hot-toast';
import { createAlertRule } from '../../api/alerts';
import { createLogMetricRule } from '../../api/observability';
import { getApiErrorMessage } from '../../api/client';
import type { LogHit, LogSearchRequest } from '../../api/types';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Modal } from '../ui/Modal';
import { Select } from '../ui/Select';
import {
  buildShareLink,
  exportLogsCsv,
  exportLogsJson,
  exportLogsNdjson,
  exportLogsTxt,
  fetchAllLogsForExport,
} from '../../utils/logExport';

interface LogQueryActionsProps {
  query: string;
  mode: 'text' | 'regex' | 'ai';
  hits: LogHit[];
  start: Date;
  end: Date;
  searchParams: LogSearchRequest;
  filterExpression: string;
  onExcludeQuery?: (term: string) => void;
  disabled?: boolean;
}

export function LogQueryActions({
  query,
  mode,
  hits,
  start,
  end,
  searchParams,
  filterExpression,
  onExcludeQuery,
  disabled,
}: LogQueryActionsProps) {
  const [downloadOpen, setDownloadOpen] = useState(false);
  const [actionsOpen, setActionsOpen] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [exportProgress, setExportProgress] = useState('');
  const [alertOpen, setAlertOpen] = useState(false);
  const [metricOpen, setMetricOpen] = useState(false);

  const rangeMeta = { start: start.toISOString(), end: end.toISOString() };

  const shareLink = () => {
    const link = buildShareLink(query, rangeMeta.start, rangeMeta.end, { mode });
    void navigator.clipboard.writeText(link);
    toast.success('Share link copied');
  };

  const copyFilter = () => {
    void navigator.clipboard.writeText(filterExpression || query);
    toast.success('Filter copied');
  };

  const exportPage = (format: 'json' | 'csv' | 'ndjson' | 'txt') => {
    if (!hits.length) return;
    switch (format) {
      case 'json':
        exportLogsJson(hits, query, rangeMeta);
        break;
      case 'csv':
        exportLogsCsv(hits, rangeMeta);
        break;
      case 'ndjson':
        exportLogsNdjson(hits, rangeMeta);
        break;
      case 'txt':
        exportLogsTxt(hits, rangeMeta);
        break;
    }
    toast.success(`Downloaded ${hits.length} logs (${format.toUpperCase()})`);
    setDownloadOpen(false);
  };

  const exportFullRange = async (format: 'json' | 'csv' | 'ndjson' | 'txt') => {
    if (!query || mode === 'ai') {
      toast.error('Full export requires a text or regex query');
      return;
    }
    setExporting(true);
    setExportProgress('Fetching logs…');
    try {
      const all = await fetchAllLogsForExport(searchParams, {
        onProgress: (loaded, total) => setExportProgress(`${loaded.toLocaleString()} / ${total.toLocaleString()}`),
      });
      if (!all.length) {
        toast.error('No logs matched in this time range');
        return;
      }
      switch (format) {
        case 'json':
          exportLogsJson(all, query, rangeMeta);
          break;
        case 'csv':
          exportLogsCsv(all, rangeMeta);
          break;
        case 'ndjson':
          exportLogsNdjson(all, rangeMeta);
          break;
        case 'txt':
          exportLogsTxt(all, rangeMeta);
          break;
      }
      toast.success(`Downloaded ${all.length.toLocaleString()} logs for selected period`);
      setDownloadOpen(false);
    } catch (e) {
      toast.error(getApiErrorMessage(e));
    } finally {
      setExporting(false);
      setExportProgress('');
    }
  };

  return (
    <>
      <div className="log-query-actions">
        <div className="log-query-actions__dropdown">
          <Button
            variant="secondary"
            size="sm"
            disabled={disabled || !query}
            onClick={() => {
              setDownloadOpen((o) => !o);
              setActionsOpen(false);
            }}
          >
            <Download size={14} /> Download
            <ChevronDown size={12} />
          </Button>
          {downloadOpen && (
            <div className="log-query-actions__menu" role="menu">
              <span className="log-query-actions__menu-heading">Current page ({hits.length})</span>
              {(['json', 'csv', 'ndjson', 'txt'] as const).map((fmt) => (
                <button
                  key={`page-${fmt}`}
                  type="button"
                  role="menuitem"
                  disabled={!hits.length}
                  onClick={() => exportPage(fmt)}
                >
                  {fmt.toUpperCase()}
                </button>
              ))}
              <span className="log-query-actions__menu-heading">Full time range</span>
              {exporting && <span className="log-query-actions__progress">{exportProgress}</span>}
              {(['json', 'csv', 'ndjson', 'txt'] as const).map((fmt) => (
                <button
                  key={`full-${fmt}`}
                  type="button"
                  role="menuitem"
                  disabled={exporting || mode === 'ai'}
                  onClick={() => void exportFullRange(fmt)}
                >
                  {fmt.toUpperCase()} (all matching)
                </button>
              ))}
            </div>
          )}
        </div>

        <div className="log-query-actions__dropdown">
          <Button
            variant="ghost"
            size="sm"
            disabled={disabled || !query}
            onClick={() => {
              setActionsOpen((o) => !o);
              setDownloadOpen(false);
            }}
          >
            Actions
            <ChevronDown size={12} />
          </Button>
          {actionsOpen && (
            <div className="log-query-actions__menu" role="menu">
              <button type="button" role="menuitem" onClick={() => { setAlertOpen(true); setActionsOpen(false); }}>
                <Bell size={14} /> Create alert from query
              </button>
              <button type="button" role="menuitem" onClick={() => { setMetricOpen(true); setActionsOpen(false); }}>
                <BarChart3 size={14} /> Create log-based metric
              </button>
              <button type="button" role="menuitem" onClick={() => { copyFilter(); setActionsOpen(false); }}>
                <Copy size={14} /> Copy filter
              </button>
              <button type="button" role="menuitem" onClick={() => { shareLink(); setActionsOpen(false); }}>
                <Link2 size={14} /> Copy link to query
              </button>
              {query && onExcludeQuery && (
                <button
                  type="button"
                  role="menuitem"
                  onClick={() => {
                    onExcludeQuery(query);
                    setActionsOpen(false);
                    toast.success('Added exclusion to query');
                  }}
                >
                  <MinusCircle size={14} /> Exclude matching logs
                </button>
              )}
              <button
                type="button"
                role="menuitem"
                onClick={() => {
                  void navigator.clipboard.writeText(
                    `Time: ${start.toLocaleString()} → ${end.toLocaleString()}\n${filterExpression}`,
                  );
                  setActionsOpen(false);
                  toast.success('Query + time range copied');
                }}
              >
                <Filter size={14} /> Copy query with time range
              </button>
            </div>
          )}
        </div>
      </div>

      <CreateLogAlertModal
        open={alertOpen}
        onOpenChange={setAlertOpen}
        query={query}
        filterExpression={filterExpression}
      />
      <CreateLogMetricModal open={metricOpen} onOpenChange={setMetricOpen} query={query} />
    </>
  );
}

function CreateLogAlertModal({
  open,
  onOpenChange,
  query,
  filterExpression,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  query: string;
  filterExpression: string;
}) {
  const [name, setName] = useState('');
  const [severity, setSeverity] = useState('P2');
  const [servicePattern, setServicePattern] = useState('*');

  const createMut = useMutation({
    mutationFn: () =>
      createAlertRule({
        name: name.trim(),
        source: 'CUSTOM',
        servicePattern: servicePattern.trim() || '*',
        severity,
        enabled: true,
        labels: {
          signal: 'logs',
          query: query.trim(),
          filter: filterExpression,
        },
      }),
    onSuccess: () => {
      toast.success('Log alert rule created');
      onOpenChange(false);
      setName('');
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  return (
    <Modal open={open} onOpenChange={onOpenChange} title="Create alert from log query">
      <p className="muted log-query-modal__hint">
        Fires when logs matching this filter appear. Review rules under Alerts → Rules.
      </p>
      <label className="log-query-modal__field">
        <span>Rule name</span>
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="High error rate in payment-api" />
      </label>
      <label className="log-query-modal__field">
        <span>Service pattern</span>
        <Input value={servicePattern} onChange={(e) => setServicePattern(e.target.value)} placeholder="payment-api or *" />
      </label>
      <label className="log-query-modal__field">
        <span>Severity</span>
        <Select value={severity} onChange={(e) => setSeverity(e.target.value)}>
          <option value="P1">P1 — Critical</option>
          <option value="P2">P2 — High</option>
          <option value="P3">P3 — Medium</option>
          <option value="P4">P4 — Low</option>
        </Select>
      </label>
      <label className="log-query-modal__field">
        <span>Log filter</span>
        <textarea className="log-query-modal__textarea" readOnly value={filterExpression || query} rows={4} />
      </label>
      <div className="log-query-modal__actions">
        <Button variant="ghost" onClick={() => onOpenChange(false)}>Cancel</Button>
        <Button variant="primary" disabled={!name.trim() || createMut.isPending} onClick={() => createMut.mutate()}>
          Create alert rule
        </Button>
      </div>
    </Modal>
  );
}

function CreateLogMetricModal({
  open,
  onOpenChange,
  query,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  query: string;
}) {
  const [name, setName] = useState('');
  const [pattern, setPattern] = useState(query);
  const [service, setService] = useState('');

  const createMut = useMutation({
    mutationFn: () =>
      createLogMetricRule({
        name: name.trim(),
        pattern: pattern.trim(),
        service: service.trim() || undefined,
        enabled: true,
      }),
    onSuccess: () => {
      toast.success('Log-based metric created');
      onOpenChange(false);
      setName('');
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  return (
    <Modal open={open} onOpenChange={onOpenChange} title="Create log-based metric">
      <p className="muted log-query-modal__hint">
        Counts log lines matching a pattern — similar to GCP log-based metrics.
      </p>
      <label className="log-query-modal__field">
        <span>Metric name</span>
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="payment_timeout_count" />
      </label>
      <label className="log-query-modal__field">
        <span>Match pattern (regex or substring)</span>
        <Input value={pattern} onChange={(e) => setPattern(e.target.value)} placeholder="timeout|deadline exceeded" />
      </label>
      <label className="log-query-modal__field">
        <span>Service (optional)</span>
        <Input value={service} onChange={(e) => setService(e.target.value)} placeholder="payment-api" />
      </label>
      <div className="log-query-modal__actions">
        <Button variant="ghost" onClick={() => onOpenChange(false)}>Cancel</Button>
        <Button
          variant="primary"
          disabled={!name.trim() || !pattern.trim() || createMut.isPending}
          onClick={() => createMut.mutate()}
        >
          Create metric
        </Button>
      </div>
    </Modal>
  );
}
