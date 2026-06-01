import { useEffect, useRef } from 'react';
import { useRouter } from '@tanstack/react-router';
import { useFilterStore, type Environment, type TimeRangePreset } from '../store/filterStore';

const VALID_ENV = new Set(['ALL', 'PROD', 'STAGING', 'DEV']);
const VALID_RANGE = new Set(['1h', '6h', '24h', '7d', 'custom']);

/** Sync global env + time range with URL search params. */
export function useFilterUrlSync() {
  const router = useRouter();
  const environment = useFilterStore((s) => s.environment);
  const timeRange = useFilterStore((s) => s.timeRange);
  const setEnvironment = useFilterStore((s) => s.setEnvironment);
  const setTimeRange = useFilterStore((s) => s.setTimeRange);
  const hydrated = useRef(false);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const env = params.get('env');
    const range = params.get('range');
    if (env && VALID_ENV.has(env)) setEnvironment(env as Environment);
    if (range && VALID_RANGE.has(range)) setTimeRange(range as TimeRangePreset);
    hydrated.current = true;
  }, [setEnvironment, setTimeRange]);

  useEffect(() => {
    if (!hydrated.current) return;
    const params = new URLSearchParams(window.location.search);
    if (environment !== 'ALL') params.set('env', environment);
    else params.delete('env');
    if (timeRange !== '24h') params.set('range', timeRange);
    else params.delete('range');
    const qs = params.toString();
    const next = `${window.location.pathname}${qs ? `?${qs}` : ''}${window.location.hash}`;
    if (next !== `${window.location.pathname}${window.location.search}${window.location.hash}`) {
      void router.history.replace(next);
    }
  }, [environment, timeRange, router]);
}
