import { useCallback } from 'react';
import { useFilterStore } from '../store/filterStore';

export interface ResolvedTimeBounds {
  start: Date;
  end: Date;
  startIso: string;
  endIso: string;
}

/**
 * Bridges the global time-range filter into data queries.
 *
 * `key` is a stable cache key suitable for React Query `queryKey`s: presets are
 * keyed by their name (so the "now" edge can drift without churning the key /
 * causing refetch loops), while custom ranges are keyed by their exact bounds.
 * `resolve()` returns the concrete bounds at call time so the actual request
 * still uses an up-to-date window for relative presets.
 */
export function useTimeBounds(): { key: string; resolve: () => ResolvedTimeBounds } {
  const timeRange = useFilterStore((s) => s.timeRange);
  const customStart = useFilterStore((s) => s.customStart);
  const customEnd = useFilterStore((s) => s.customEnd);
  const getTimeBounds = useFilterStore((s) => s.getTimeBounds);

  const key =
    timeRange === 'custom'
      ? `custom:${customStart?.toISOString() ?? ''}:${customEnd?.toISOString() ?? ''}`
      : timeRange;

  const resolve = useCallback((): ResolvedTimeBounds => {
    const { start, end } = getTimeBounds();
    return { start, end, startIso: start.toISOString(), endIso: end.toISOString() };
    // `key` intentionally participates so a custom-range change yields a fresh closure.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [getTimeBounds, key]);

  return { key, resolve };
}
