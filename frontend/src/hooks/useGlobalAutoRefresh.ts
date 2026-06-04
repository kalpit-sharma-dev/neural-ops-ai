import { useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useUiPreferencesStore } from '../store/uiPreferencesStore';

/** Global auto-refresh tick — bumps generation and invalidates active queries. */
export function useGlobalAutoRefresh() {
  const intervalMs = useUiPreferencesStore((s) => s.refreshIntervalMs);
  const tickRefresh = useUiPreferencesStore((s) => s.tickRefresh);
  const queryClient = useQueryClient();

  useEffect(() => {
    if (!intervalMs) return;
    const id = window.setInterval(() => {
      tickRefresh();
      void queryClient.invalidateQueries();
    }, intervalMs);
    return () => window.clearInterval(id);
  }, [intervalMs, tickRefresh, queryClient]);
}

export function formatLastUpdated(ts: number): string {
  const sec = Math.max(0, Math.floor((Date.now() - ts) / 1000));
  if (sec < 5) return 'just now';
  if (sec < 60) return `${sec}s ago`;
  const min = Math.floor(sec / 60);
  if (min < 60) return `${min}m ago`;
  return `${Math.floor(min / 60)}h ago`;
}
