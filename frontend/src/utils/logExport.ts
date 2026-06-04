import { searchLogs } from '../api/search';
import type { LogHit, LogSearchRequest } from '../api/types';
import { downloadBlob } from './dataExport';

function fieldValue(hit: LogHit, col: string): string {
  const record = hit as unknown as Record<string, unknown>;
  return String(record[col] ?? '');
}

function exportStamp(start: string, end: string): string {
  const s = start.slice(0, 10);
  const e = end.slice(0, 10);
  return s === e ? s : `${s}_to_${e}`;
}

export function exportLogsJson(hits: LogHit[], query: string, range?: { start: string; end: string }) {
  const payload = {
    exportedAt: new Date().toISOString(),
    query,
    timeRange: range,
    count: hits.length,
    hits,
  };
  const stamp = range ? exportStamp(range.start, range.end) : String(Date.now());
  downloadBlob(JSON.stringify(payload, null, 2), `logs-${stamp}.json`, 'application/json');
}

export function exportLogsCsv(hits: LogHit[], range?: { start: string; end: string }) {
  const headers = ['timestamp', 'severity', 'service', 'message', 'traceId', 'txnId', 'host', 'pod'];
  const rows = hits.map((h) =>
    headers
      .map((col) => {
        const val = fieldValue(h, col);
        return `"${val.replace(/"/g, '""')}"`;
      })
      .join(','),
  );
  const stamp = range ? exportStamp(range.start, range.end) : String(Date.now());
  downloadBlob([headers.join(','), ...rows].join('\n'), `logs-${stamp}.csv`, 'text/csv');
}

export function exportLogsNdjson(hits: LogHit[], range?: { start: string; end: string }) {
  const body = hits.map((h) => JSON.stringify(h)).join('\n');
  const stamp = range ? exportStamp(range.start, range.end) : String(Date.now());
  downloadBlob(body, `logs-${stamp}.ndjson`, 'application/x-ndjson');
}

export function exportLogsTxt(hits: LogHit[], range?: { start: string; end: string }) {
  const lines = hits.map(
    (h) => `${h.timestamp}\t${h.severity}\t${h.service}\t${h.message.replace(/\n/g, ' ')}`,
  );
  const stamp = range ? exportStamp(range.start, range.end) : String(Date.now());
  downloadBlob(lines.join('\n'), `logs-${stamp}.txt`, 'text/plain');
}

/** Paginate through search API to export all logs matching the query in the time window. */
export async function fetchAllLogsForExport(
  params: LogSearchRequest,
  options?: { maxRows?: number; pageSize?: number; onProgress?: (loaded: number, total: number) => void },
): Promise<LogHit[]> {
  const maxRows = options?.maxRows ?? 10_000;
  const pageSize = Math.min(options?.pageSize ?? 1000, maxRows);
  const all: LogHit[] = [];
  let searchAfter: unknown[] | undefined;

  while (all.length < maxRows) {
    const batch = await searchLogs({
      ...params,
      size: Math.min(pageSize, maxRows - all.length),
      searchAfter,
    });
    if (!batch.hits.length) break;
    all.push(...batch.hits);
    options?.onProgress?.(all.length, batch.total);
    if (!batch.searchAfter || batch.hits.length < pageSize) break;
    searchAfter = batch.searchAfter;
  }

  return all;
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

/** GCP-style filter string from structured filters + free-text query. */
export function buildLogFilterExpression(parts: {
  query?: string;
  services?: string[];
  severities?: string[];
  host?: string;
  pod?: string;
  environment?: string;
}): string {
  const clauses: string[] = [];
  if (parts.query?.trim()) clauses.push(parts.query.trim());
  parts.services?.forEach((s) => clauses.push(`service="${s}"`));
  parts.severities?.forEach((s) => clauses.push(`severity="${s}"`));
  if (parts.host) clauses.push(`host:"${parts.host}"`);
  if (parts.pod) clauses.push(`pod:"${parts.pod}"`);
  if (parts.environment && parts.environment !== 'ALL') {
    clauses.push(`labels.environment="${parts.environment}"`);
  }
  return clauses.join('\n');
}
