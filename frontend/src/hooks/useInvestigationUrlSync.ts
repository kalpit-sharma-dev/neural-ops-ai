import { useEffect } from 'react';
import { useRouterState } from '@tanstack/react-router';
import { useInvestigationStore } from '../store/investigationStore';

/** Keep investigation chips in sync with URL search params. */
export function useInvestigationUrlSync() {
  const search = useRouterState({ select: (s) => s.location.search }) as Record<string, unknown>;
  const syncFromSearch = useInvestigationStore((s) => s.syncFromSearch);

  useEffect(() => {
    syncFromSearch(search);
  }, [search, syncFromSearch]);
}
