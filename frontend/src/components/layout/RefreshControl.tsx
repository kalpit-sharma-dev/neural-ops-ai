import { RefreshCw } from 'lucide-react';
import { useQueryClient } from '@tanstack/react-query';
import { REFRESH_OPTIONS, useUiPreferencesStore } from '../../store/uiPreferencesStore';
import { formatLastUpdated } from '../../hooks/useGlobalAutoRefresh';

export function RefreshControl() {
  const intervalMs = useUiPreferencesStore((s) => s.refreshIntervalMs);
  const lastRefreshedAt = useUiPreferencesStore((s) => s.lastRefreshedAt);
  const setRefreshIntervalMs = useUiPreferencesStore((s) => s.setRefreshIntervalMs);
  const tickRefresh = useUiPreferencesStore((s) => s.tickRefresh);
  const queryClient = useQueryClient();

  const manualRefresh = () => {
    tickRefresh();
    void queryClient.invalidateQueries();
  };

  return (
    <div className="refresh-control">
      <button
        type="button"
        className="refresh-control__btn icon-btn"
        aria-label="Refresh data"
        onClick={manualRefresh}
      >
        <RefreshCw size={16} />
      </button>
      <select
        className="refresh-control__select"
        aria-label="Auto-refresh interval"
        value={intervalMs}
        onChange={(e) => setRefreshIntervalMs(Number(e.target.value) as typeof intervalMs)}
      >
        {REFRESH_OPTIONS.map((o) => (
          <option key={o.value} value={o.value}>
            {o.value === 0 ? 'Auto: Off' : `Auto: ${o.label}`}
          </option>
        ))}
      </select>
      <span className="refresh-control__updated muted" title="Last data refresh">
        {formatLastUpdated(lastRefreshedAt)}
      </span>
    </div>
  );
}
