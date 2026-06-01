/** Normalized search params for `/logs` (TanStack Router requires all keys). */
export function logsSearch(partial: {
  service?: string;
  traceId?: string;
  from?: string;
  to?: string;
}) {
  return {
    service: partial.service,
    traceId: partial.traceId,
    from: partial.from,
    to: partial.to,
  };
}
