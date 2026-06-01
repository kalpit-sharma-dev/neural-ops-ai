import { useCallback, useEffect, useMemo, useRef, useState, type MouseEvent } from 'react';
import { Link, useSearch } from '@tanstack/react-router';
import { useQuery } from '@tanstack/react-query';
import { useVirtualizer } from '@tanstack/react-virtual';
import { AnimatePresence, motion } from 'framer-motion';
import { format, formatDistanceToNow } from 'date-fns';
import { Bookmark, Copy, Download, Link2, Sparkles, Trash2 } from 'lucide-react';
import toast from 'react-hot-toast';
import { aiSearch, searchLogs } from '../api/search';
import { fetchDashboardOverview } from '../api/dashboard';
import { getApiErrorMessage } from '../api/client';
import type { LogHit } from '../api/types';
import { LogDetailPanel } from '../components/logs/LogDetailPanel';
import { LogFiltersSidebar } from '../components/logs/LogFiltersSidebar';
import { LogHistogram } from '../components/logs/LogHistogram';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { SearchInput } from '../components/ui/SearchInput';
import { DomainEmptyState } from '../components/ui/DomainEmptyState';
import { ErrorState, LoadingState, PageHeader } from '../components/ui/PageStates';
import { useFilterStore } from '../store/filterStore';
import { useAuthStore } from '../store/authStore';
import { highlightMatches, hasStackTrace } from '../utils/highlightText';
import { buildShareLink, exportLogsCsv, exportLogsJson } from '../utils/logExport';
import { deleteSavedSearch, listSavedSearches, saveSearch, type SavedSearch } from '../lib/savedSearches';
import { loadLogColumns, saveLogColumns, type LogColumnPref } from '../lib/logColumnPrefs';

type SearchMode = 'text' | 'regex' | 'ai';

function severityVariant(sev: string) {
  if (sev === 'FATAL' || sev === 'CRITICAL') return 'critical';
  if (sev === 'ERROR') return 'error';
  if (sev === 'WARN') return 'warning';
  return 'info';
}

function applyClientFilters(hits: LogHit[], filters: ReturnType<typeof useFilterStore.getState>): LogHit[] {
  return hits.filter((hit) => {
    if (filters.services.length > 0 && !filters.services.includes(hit.service)) return false;
    if (filters.severities.length > 0 && !filters.severities.includes(hit.severity)) return false;
    if (filters.host && hit.host && !hit.host.includes(filters.host)) return false;
    if (filters.pod && hit.pod && !hit.pod.includes(filters.pod)) return false;
    if (filters.hasStackTrace && !hasStackTrace(hit.message)) return false;
    if (filters.hasAIExplanation && !hit.plainEnglish) return false;
    if (filters.environment !== 'ALL') {
      const env = hit.labels?.environment ?? hit.labels?.env ?? '';
      if (env && !env.toUpperCase().includes(filters.environment)) return false;
    }
    return true;
  });
}

export default function LogExplorer() {
  const urlSearch = useSearch({ strict: false }) as {
    service?: string;
    traceId?: string;
    from?: string;
    to?: string;
    q?: string;
    mode?: string;
  };
  const filterState = useFilterStore();
  const setCustomRange = useFilterStore((s) => s.setCustomRange);
  const user = useAuthStore((s) => s.user);
  const tenantId = useAuthStore((s) => s.tenantId);
  const { getTimeBounds, host, pod, services, severities, environment, hasStackTrace, hasAIExplanation } =
    filterState;
  const [query, setQuery] = useState(urlSearch.q ?? '');
  const [mode, setMode] = useState<SearchMode>(
    urlSearch.mode === 'regex' || urlSearch.mode === 'ai' ? urlSearch.mode : 'text',
  );
  const [selected, setSelected] = useState<LogHit | null>(null);
  const [liveTail, setLiveTail] = useState(false);
  const [tailPaused, setTailPaused] = useState(false);
  const [columns, setColumns] = useState<LogColumnPref[]>(() => loadLogColumns());
  const [saveName, setSaveName] = useState('');
  const [saved, setSaved] = useState<SavedSearch[]>([]);
  const parentRef = useRef<HTMLDivElement>(null);

  const storageTenant = tenantId ?? 'default';
  const storageUser = user?.id ?? 'anonymous';

  const reloadSaved = useCallback(() => {
    setSaved(listSavedSearches(storageTenant, storageUser));
  }, [storageTenant, storageUser]);

  useEffect(() => {
    reloadSaved();
  }, [reloadSaved]);

  useEffect(() => {
    if (urlSearch.traceId) setQuery(urlSearch.traceId);
    else if (urlSearch.service && !query) setQuery(`service:${urlSearch.service}`);
    if (urlSearch.q) setQuery(urlSearch.q);
  }, [urlSearch.traceId, urlSearch.service, urlSearch.q]);

  const { start, end } = useMemo(() => {
    if (urlSearch.from && urlSearch.to) {
      return { start: new Date(urlSearch.from), end: new Date(urlSearch.to) };
    }
    return getTimeBounds();
  }, [urlSearch.from, urlSearch.to, getTimeBounds]);

  const dashboardQuery = useQuery({
    queryKey: ['dashboard-overview'],
    queryFn: fetchDashboardOverview,
  });

  const serviceHealth = useMemo(() => {
    const map: Record<string, number> = {};
    (dashboardQuery.data?.serviceHealth ?? []).forEach((s) => {
      map[s.service] = s.score;
    });
    return map;
  }, [dashboardQuery.data]);

  const searchQuery = useQuery({
    queryKey: ['logs', query, mode, start.toISOString(), end.toISOString(), host, pod],
    queryFn: async () => {
      if (mode === 'ai') {
        const result = await aiSearch(query, 500);
        return result.results;
      }
      return searchLogs({
        query: mode === 'text' ? query : undefined,
        messageRegex: mode === 'regex' ? query : undefined,
        host: host || undefined,
        pod: pod || undefined,
        startTime: start.toISOString(),
        endTime: end.toISOString(),
        size: 1000,
      });
    },
    enabled: query.length > 0,
    refetchInterval: liveTail && !tailPaused ? 5000 : false,
  });

  const hits = useMemo(
    () => applyClientFilters(searchQuery.data?.hits ?? [], filterState),
    [searchQuery.data, services, severities, environment, hasStackTrace, hasAIExplanation, filterState],
  );

  const severityCounts = useMemo(() => {
    const counts: Record<string, number> = {};
    (searchQuery.data?.hits ?? []).forEach((h) => {
      counts[h.severity] = (counts[h.severity] ?? 0) + 1;
    });
    return counts;
  }, [searchQuery.data]);

  const serviceOptions = useMemo(() => {
    const buckets = searchQuery.data?.aggregations?.byService ?? [];
    if (buckets.length) return buckets.map((b) => b.key);
    return Object.keys(serviceHealth);
  }, [searchQuery.data, serviceHealth]);

  const rowVirtualizer = useVirtualizer({
    count: hits.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 44,
    overscan: 25,
  });

  const copyTrace = useCallback((traceId: string, e: MouseEvent) => {
    e.stopPropagation();
    void navigator.clipboard.writeText(traceId);
    toast.success('Trace ID copied');
  }, []);

  const shareLink = useCallback(() => {
    const link = buildShareLink(query, start.toISOString(), end.toISOString(), { mode });
    void navigator.clipboard.writeText(link);
    toast.success('Share link copied');
  }, [query, start, end, mode]);

  const clearTail = () => {
    searchQuery.refetch();
    toast.success('Stream cleared — showing latest results');
  };

  const onSaveSearch = () => {
    if (!saveName.trim() || !query) return;
    saveSearch(storageTenant, storageUser, { name: saveName.trim(), query, mode });
    setSaveName('');
    reloadSaved();
    toast.success('Search saved');
  };

  const resizeColumn = (id: LogColumnPref['id'], delta: number) => {
    const next = columns.map((c) =>
      c.id === id ? { ...c, width: Math.max(40, c.width + delta) } : c,
    );
    setColumns(next);
    saveLogColumns(next);
  };

  return (
    <div>
      <PageHeader title="Log Explorer" subtitle="Search and analyze logs with AI-powered filters" />

      <div className={`log-explorer ${selected ? 'log-explorer--detail' : ''}`}>
        <LogFiltersSidebar
          serviceOptions={serviceOptions}
          severityCounts={severityCounts}
          serviceHealth={serviceHealth}
        />

        <section className="log-stream">
          <div className="log-stream__header">
            <SearchInput
              placeholder="Search logs… or try 'payment failures after 10pm' (AI)"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              shortcut="⌘K"
            />
            <div className="log-stream__controls">
              {(['text', 'regex', 'ai'] as SearchMode[]).map((m) => (
                <Button key={m} variant={mode === m ? 'primary' : 'ghost'} size="sm" onClick={() => setMode(m)}>
                  {m === 'ai' ? 'AI/NLP' : m === 'regex' ? 'Regex' : 'Text'}
                </Button>
              ))}
              <label className="checkbox-row log-stream__live">
                <input type="checkbox" checked={liveTail} onChange={(e) => setLiveTail(e.target.checked)} />
                Live tail
              </label>
              {liveTail && (
                <Button variant="ghost" size="sm" onClick={clearTail}>
                  Clear
                </Button>
              )}
            </div>
            <div className="log-saved-searches">
              <Input
                placeholder="Name saved search"
                value={saveName}
                onChange={(e) => setSaveName(e.target.value)}
                aria-label="Saved search name"
              />
              <Button variant="secondary" size="sm" disabled={!query || !saveName} onClick={onSaveSearch}>
                <Bookmark size={14} /> Save
              </Button>
              {saved.map((s) => (
                <span key={s.id} className="log-saved-chip">
                  <button
                    type="button"
                    onClick={() => {
                      setQuery(s.query);
                      setMode(s.mode);
                    }}
                  >
                    {s.name}
                  </button>
                  <button
                    type="button"
                    aria-label={`Delete ${s.name}`}
                    onClick={() => {
                      deleteSavedSearch(storageTenant, storageUser, s.id);
                      reloadSaved();
                    }}
                  >
                    <Trash2 size={12} />
                  </button>
                </span>
              ))}
            </div>
          </div>

          {hits.length > 0 && (
            <LogHistogram
              hits={hits}
              onBrush={(s, e) => {
                setCustomRange(s, e);
                toast.success('Time range narrowed');
              }}
            />
          )}

          <div className="log-column-header">
            {columns
              .filter((c) => c.visible)
              .map((col) => (
                <span
                  key={col.id}
                  className="log-column-header__cell"
                  style={{ flex: col.id === 'message' ? 1 : `0 0 ${col.width}px` }}
                >
                  {col.id}
                  <button type="button" aria-label={`Resize ${col.id}`} onMouseDown={() => resizeColumn(col.id, 8)}>
                    ⋮
                  </button>
                </span>
              ))}
          </div>

          <div
            className="log-stream__body"
            ref={parentRef}
            onMouseEnter={() => liveTail && setTailPaused(true)}
            onMouseLeave={() => liveTail && setTailPaused(false)}
          >
            {searchQuery.isLoading && <LoadingState label="Searching logs…" />}
            {searchQuery.error && (
              <ErrorState message={getApiErrorMessage(searchQuery.error)} onRetry={() => searchQuery.refetch()} />
            )}
            {!searchQuery.isLoading && hits.length === 0 && query && (
              <DomainEmptyState domain="logs" onAction={() => setMode('ai')} actionLabel="Try AI search" />
            )}
            {!query && <DomainEmptyState domain="logs" />}

            <div style={{ height: rowVirtualizer.getTotalSize(), position: 'relative' }}>
              {rowVirtualizer.getVirtualItems().map((virtualRow) => {
                const hit = hits[virtualRow.index];
                return (
                  <div
                    key={hit.id}
                    className={`log-row severity-bar severity-bar--${hit.severity} ${selected?.id === hit.id ? 'log-row--selected' : ''}`}
                    style={{
                      position: 'absolute',
                      top: 0,
                      left: 0,
                      width: '100%',
                      transform: `translateY(${virtualRow.start}px)`,
                    }}
                    onClick={() => setSelected(hit)}
                  >
                    <Badge variant={severityVariant(hit.severity) as 'error'}>{hit.severity.slice(0, 1)}</Badge>
                    <span className="log-row__time" title={format(new Date(hit.timestamp), 'PPpp')}>
                      {formatDistanceToNow(new Date(hit.timestamp), { addSuffix: true })}
                    </span>
                    <button
                      type="button"
                      className="log-row__service"
                      onClick={(e) => {
                        e.stopPropagation();
                        filterState.toggleService(hit.service);
                      }}
                    >
                      {hit.service}
                    </button>
                    <span className="log-row__message">
                      {highlightMatches(hit.message.slice(0, 160), query, mode === 'regex')}
                    </span>
                    {hit.traceId && (
                      <Link
                        to="/traces/$traceId"
                        params={{ traceId: hit.traceId }}
                        className="log-row__trace"
                        title={hit.traceId}
                        onClick={(e) => e.stopPropagation()}
                      >
                        {hit.traceId.slice(0, 8)}
                      </Link>
                    )}
                    {hit.traceId && (
                      <button
                        type="button"
                        className="log-row__trace-copy"
                        title="Copy trace ID"
                        onClick={(e) => copyTrace(hit.traceId!, e)}
                      >
                        <Copy size={12} />
                      </button>
                    )}
                    {hit.plainEnglish && (
                      <button
                        type="button"
                        className="log-row__ai"
                        title="AI explanation available"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelected(hit);
                        }}
                      >
                        <Sparkles size={14} />
                      </button>
                    )}
                  </div>
                );
              })}
            </div>
          </div>

          <footer className="log-stream__footer">
            <span className="muted">
              {searchQuery.data?.total ?? hits.length} results · {searchQuery.data?.tookMs ?? 0}ms
              {liveTail && (tailPaused ? ' · tail paused' : ' · live')}
            </span>
            <div className="log-stream__export">
              <Button variant="ghost" size="sm" disabled={!hits.length} onClick={() => exportLogsJson(hits, query)}>
                <Download size={14} /> JSON
              </Button>
              <Button variant="ghost" size="sm" disabled={!hits.length} onClick={() => exportLogsCsv(hits)}>
                <Download size={14} /> CSV
              </Button>
              <Button variant="ghost" size="sm" disabled={!query} onClick={shareLink}>
                <Link2 size={14} /> Share
              </Button>
            </div>
          </footer>
        </section>

        <AnimatePresence>
          {selected && (
            <motion.div
              className="log-detail-panel-wrap"
              initial={{ x: 40, opacity: 0 }}
              animate={{ x: 0, opacity: 1 }}
              exit={{ x: 40, opacity: 0 }}
            >
              <LogDetailPanel log={selected} onClose={() => setSelected(null)} />
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </div>
  );
}
