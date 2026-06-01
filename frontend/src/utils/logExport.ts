import type { LogHit } from '../api/types';

function fieldValue(hit: LogHit, col: string): string {
  const record = hit as unknown as Record<string, unknown>;
  return String(record[col] ?? '');
}

function downloadBlob(content: string, filename: string, mime: string) {
  const blob = new Blob([content], { type: mime });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

export function exportLogsJson(hits: LogHit[], query: string) {
  const payload = {
    exportedAt: new Date().toISOString(),
    query,
    count: hits.length,
    hits,
  };
  downloadBlob(JSON.stringify(payload, null, 2), `logs-${Date.now()}.json`, 'application/json');
}

export function exportLogsCsv(hits: LogHit[]) {
  const headers = ['timestamp', 'severity', 'service', 'message', 'traceId', 'txnId', 'host', 'pod'];
  const rows = hits.map((h) =>
    headers
      .map((col) => {
        const val = fieldValue(h, col);
        return `"${val.replace(/"/g, '""')}"`;
      })
      .join(','),
  );
  downloadBlob([headers.join(','), ...rows].join('\n'), `logs-${Date.now()}.csv`, 'text/csv');
}

export function buildShareLink(
  query: string,
  start: string,
  end: string,
  extras?: { mode?: string; service?: string; traceId?: string },
): string {
  const params = new URLSearchParams({ q: query, from: start, to: end });
  if (extras?.mode) params.set('mode', extras.mode);
  if (extras?.service) params.set('service', extras.service);
  if (extras?.traceId) params.set('traceId', extras.traceId);
  return `${window.location.origin}/logs?${params.toString()}`;
}
