/** Generic JSON/CSV/NDJSON download helpers for observability data exports. */

export function downloadBlob(content: string, filename: string, mime: string) {
  const blob = new Blob([content], { type: mime });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

export function stampFilename(prefix: string, ext: string): string {
  const d = new Date().toISOString().slice(0, 19).replace(/[:T]/g, '-');
  return `${prefix}-${d}.${ext}`;
}

export function exportJson(data: unknown, filenamePrefix: string, meta?: Record<string, unknown>) {
  const payload = meta ? { exportedAt: new Date().toISOString(), ...meta, data } : data;
  downloadBlob(JSON.stringify(payload, null, 2), stampFilename(filenamePrefix, 'json'), 'application/json');
}

export function exportNdjson(rows: unknown[], filenamePrefix: string) {
  const body = rows.map((r) => JSON.stringify(r)).join('\n');
  downloadBlob(body, stampFilename(filenamePrefix, 'ndjson'), 'application/x-ndjson');
}

function flattenRow(row: Record<string, unknown>, prefix = ''): Record<string, string> {
  const out: Record<string, string> = {};
  for (const [key, val] of Object.entries(row)) {
    const k = prefix ? `${prefix}.${key}` : key;
    if (val == null) out[k] = '';
    else if (typeof val === 'object' && !Array.isArray(val)) {
      Object.assign(out, flattenRow(val as Record<string, unknown>, k));
    } else if (Array.isArray(val)) {
      out[k] = val.join(';');
    } else {
      out[k] = String(val);
    }
  }
  return out;
}

export function exportCsv(rows: unknown[], filenamePrefix: string) {
  if (!rows.length) return;
  const flat = rows.map((r) => flattenRow(r as Record<string, unknown>));
  const headers = [...new Set(flat.flatMap((r) => Object.keys(r)))];
  const escape = (v: string) => `"${v.replace(/"/g, '""')}"`;
  const lines = [
    headers.join(','),
    ...flat.map((r) => headers.map((h) => escape(r[h] ?? '')).join(',')),
  ];
  downloadBlob(lines.join('\n'), stampFilename(filenamePrefix, 'csv'), 'text/csv');
}

/** Download remote report URL (FinOps generated reports). */
export async function downloadFromUrl(url: string, filename: string) {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`Download failed (${res.status})`);
  const text = await res.text();
  const ext = filename.includes('.') ? '' : url.endsWith('.csv') ? '.csv' : '.json';
  downloadBlob(text, filename.includes('.') ? filename : `${filename}${ext}`, res.headers.get('content-type') ?? 'application/octet-stream');
}
