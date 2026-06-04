import { Link, useRouter, useRouterState } from '@tanstack/react-router';
import { X } from 'lucide-react';
import { useFilterStore } from '../../store/filterStore';
import { useInvestigationStore } from '../../store/investigationStore';
import { formatRangeLabel } from './TimeRangePicker';
import { useMemo } from 'react';

const OBSERVE_PREFIXES = ['/logs', '/traces', '/metrics', '/incidents', '/alerts', '/query-workbench'];

export function InvestigationContextBar() {
  const router = useRouter();
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const environment = useFilterStore((s) => s.environment);
  const timeRange = useFilterStore((s) => s.timeRange);
  const customStart = useFilterStore((s) => s.customStart);
  const customEnd = useFilterStore((s) => s.customEnd);
  const services = useFilterStore((s) => s.services);
  const clearServices = useFilterStore((s) => s.clearServices);

  const service = useInvestigationStore((s) => s.service);
  const traceId = useInvestigationStore((s) => s.traceId);
  const txnId = useInvestigationStore((s) => s.txnId);
  const setService = useInvestigationStore((s) => s.setService);
  const setTraceId = useInvestigationStore((s) => s.setTraceId);
  const setTxnId = useInvestigationStore((s) => s.setTxnId);

  const clearSearchParam = (key: string) => {
    const params = new URLSearchParams(window.location.search);
    params.delete(key);
    const qs = params.toString();
    router.history.replace(`${pathname}${qs ? `?${qs}` : ''}${window.location.hash}`);
  };

  const onObservePage = OBSERVE_PREFIXES.some((p) => pathname === p || pathname.startsWith(`${p}/`));

  const chips = useMemo(() => {
    const list: { key: string; label: string; onClear: () => void }[] = [];
    if (environment !== 'ALL') {
      list.push({
        key: 'env',
        label: `Env: ${environment}`,
        onClear: () => useFilterStore.getState().setEnvironment('ALL'),
      });
    }
    list.push({
      key: 'time',
      label: formatRangeLabel(timeRange, customStart, customEnd),
      onClear: () => useFilterStore.getState().setTimeRange('24h'),
    });
    if (service) {
      list.push({
        key: 'service',
        label: `Service: ${service}`,
        onClear: () => {
          setService(null);
          clearSearchParam('service');
        },
      });
    }
    for (const svc of services) {
      list.push({
        key: `filter-${svc}`,
        label: `Filter: ${svc}`,
        onClear: () => useFilterStore.getState().toggleService(svc),
      });
    }
    if (traceId) {
      list.push({
        key: 'trace',
        label: `Trace: ${traceId.slice(0, 12)}…`,
        onClear: () => {
          setTraceId(null);
          clearSearchParam('traceId');
        },
      });
    }
    if (txnId) {
      list.push({
        key: 'txn',
        label: `Txn: ${txnId}`,
        onClear: () => {
          setTxnId(null);
          clearSearchParam('txnId');
        },
      });
    }
    return list;
  }, [
    environment,
    timeRange,
    customStart,
    customEnd,
    service,
    services,
    traceId,
    txnId,
    setService,
    setTraceId,
    setTxnId,
    clearSearchParam,
  ]);

  const showBar = onObservePage;

  if (!showBar) return null;

  return (
    <div className="investigation-bar" data-testid="investigation-context-bar">
      <span className="investigation-bar__label">Context</span>
      <div className="investigation-bar__chips">
        {chips.map((chip) => (
          <span key={chip.key} className="investigation-chip">
            {chip.label}
            {chip.key !== 'time' && (
              <button type="button" className="investigation-chip__clear" aria-label={`Clear ${chip.label}`} onClick={chip.onClear}>
                <X size={12} />
              </button>
            )}
          </span>
        ))}
      </div>
      <div className="investigation-bar__actions">
        {traceId && (
          <Link to="/traces/$traceId" params={{ traceId }} className="investigation-bar__link">
            Open trace
          </Link>
        )}
        {(service || traceId) && (
          <Link
            to="/logs"
            search={{ service: service ?? undefined, traceId: traceId ?? undefined }}
            className="investigation-bar__link"
          >
            Search logs
          </Link>
        )}
        {services.length > 0 && (
          <button type="button" className="investigation-bar__link investigation-bar__link--btn" onClick={clearServices}>
            Clear service filters
          </button>
        )}
      </div>
    </div>
  );
}
